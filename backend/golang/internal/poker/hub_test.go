package poker_test

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/romanthekat/planning-poker/internal/poker"
)

// wsPair spins up a local httptest server that upgrades a single incoming request to a
// websocket connection, dials it, and hands back both ends - this lets Hub's tests
// exercise real network I/O (writes, pings, closes) instead of mocking the websocket
// library, which is exactly the kind of thing Hub actually depends on.
//
// Cleanup uses CloseNow (no close handshake) so that a test which never reads a
// close frame on one side doesn't stall for the library's 5-10s handshake timeout;
// tests that specifically exercise a graceful Close/handshake set up their own reader.
func wsPair(t *testing.T) (serverConn, clientConn *websocket.Conn) {
	t.Helper()

	serverConnCh := make(chan *websocket.Conn, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		serverConnCh <- c
	}))
	t.Cleanup(srv.Close)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	clientConn, _, err := websocket.Dial(context.Background(), wsURL, nil)
	if err != nil {
		t.Fatalf("client dial: %v", err)
	}
	t.Cleanup(func() { _ = clientConn.CloseNow() })

	select {
	case serverConn = <-serverConnCh:
	case <-time.After(time.Second):
		t.Fatal("server never accepted the websocket connection")
	}
	t.Cleanup(func() { _ = serverConn.CloseNow() })

	return serverConn, clientConn
}

func newTestHub() *poker.Hub {
	discard := log.New(io.Discard, "", 0)
	return poker.NewHub(discard, discard)
}

func TestHub_AddConnection_ReplacesExisting(t *testing.T) {
	hub := newTestHub()
	const sessionId = poker.SessionId(1)
	const userId = poker.UserId(1)

	firstServer, firstClient := wsPair(t)
	hub.AddConnection(sessionId, userId, firstServer)

	// Start reading before triggering the replacement: AddConnection's Close() on the
	// old connection performs a real close handshake and needs a peer that is actively
	// reading to acknowledge it promptly, exactly like a real client would.
	readErr := make(chan error, 1)
	go func() {
		_, _, err := firstClient.Read(context.Background())
		readErr <- err
	}()

	secondServer, _ := wsPair(t)
	hub.AddConnection(sessionId, userId, secondServer)

	select {
	case err := <-readErr:
		if err == nil {
			t.Errorf("Read() on the replaced connection = nil error, want the connection to have been closed")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("replaced connection was never closed")
	}
}

func TestHub_Broadcast_DeliversMaskedPayloadAndCallsOnSent(t *testing.T) {
	hub := newTestHub()
	const sessionId = poker.SessionId(1)
	const userId = poker.UserId(1)

	serverConn, clientConn := wsPair(t)
	hub.AddConnection(sessionId, userId, serverConn)

	want := poker.Session{Id: sessionId, VotesHidden: true}
	maskFor := func(gotUserId poker.UserId) poker.Session {
		if gotUserId != userId {
			t.Errorf("Broadcast() maskFor called with userId = %v, want %v", gotUserId, userId)
		}
		return want
	}

	var mu sync.Mutex
	var sentTo []poker.UserId
	onSent := func(gotUserId poker.UserId) {
		mu.Lock()
		sentTo = append(sentTo, gotUserId)
		mu.Unlock()
	}

	hub.Broadcast(sessionId, maskFor, onSent)

	var got poker.Session
	readCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := wsjson.Read(readCtx, clientConn, &got); err != nil {
		t.Fatalf("client read: %v", err)
	}
	if got.Id != want.Id {
		t.Errorf("Broadcast() delivered session.Id = %v, want %v", got.Id, want.Id)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(sentTo) != 1 || sentTo[0] != userId {
		t.Errorf("Broadcast() onSent calls = %v, want exactly [%v]", sentTo, userId)
	}
}

func TestHub_Broadcast_NoConnectionsIsNoop(t *testing.T) {
	hub := newTestHub()

	called := false
	hub.Broadcast(poker.SessionId(42), func(poker.UserId) poker.Session {
		called = true
		return poker.Session{}
	}, nil)

	if called {
		t.Errorf("Broadcast() invoked maskFor with no registered connections")
	}
}

func TestHub_Ping_RemovesDeadConnection(t *testing.T) {
	hub := newTestHub()
	const sessionId = poker.SessionId(1)
	const userId = poker.UserId(1)

	serverConn, clientConn := wsPair(t)
	hub.AddConnection(sessionId, userId, serverConn)

	// ReadLoop is what makes Ping/pong handling work at all (see hub.go and the
	// websocket library's own docs: "you must always read from the connection"),
	// exactly as production wires it up via SaveConnectionForUser.
	go hub.ReadLoop(serverConn)

	// Kill the underlying connection outright (no handshake) so the server's read
	// fails fast, instead of waiting on a graceful close it will never receive.
	_ = clientConn.CloseNow()

	// Ping's failure is detected asynchronously by ReadLoop's error handling, so poll
	// for the connection's removal instead of asserting after a single Ping call.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		hub.Ping(sessionId)

		called := false
		hub.Broadcast(sessionId, func(poker.UserId) poker.Session {
			called = true
			return poker.Session{}
		}, nil)
		if !called {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("Ping() never removed the dead connection")
}

func TestHub_PingLoop_StopsOnContextCancel(t *testing.T) {
	hub := newTestHub()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		hub.PingLoop(ctx, poker.SessionId(1))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("PingLoop() did not return after its context was cancelled")
	}
}

func TestHub_CloseSession_ClosesAllConnections(t *testing.T) {
	hub := newTestHub()
	const sessionId = poker.SessionId(1)

	serverA, aliceClient := wsPair(t)
	serverB, bobClient := wsPair(t)

	hub.AddConnection(sessionId, poker.UserId(1), serverA)
	hub.AddConnection(sessionId, poker.UserId(2), serverB)

	// Start reading on both clients before closing: CloseSession performs a real close
	// handshake per connection and needs an actively-reading peer to ack it promptly.
	errA := make(chan error, 1)
	errB := make(chan error, 1)
	go func() {
		_, _, err := aliceClient.Read(context.Background())
		errA <- err
	}()
	go func() {
		_, _, err := bobClient.Read(context.Background())
		errB <- err
	}()

	hub.CloseSession(sessionId)

	timeout := time.After(5 * time.Second)
	select {
	case err := <-errA:
		if err == nil {
			t.Errorf("CloseSession() left the first client's connection open")
		}
	case <-timeout:
		t.Fatal("CloseSession() never closed the first client's connection")
	}
	select {
	case err := <-errB:
		if err == nil {
			t.Errorf("CloseSession() left the second client's connection open")
		}
	case <-timeout:
		t.Fatal("CloseSession() never closed the second client's connection")
	}

	// Broadcasting afterwards must be a no-op: CloseSession should have forgotten the
	// session's connections entirely, not just closed them.
	called := false
	hub.Broadcast(sessionId, func(poker.UserId) poker.Session {
		called = true
		return poker.Session{}
	}, nil)
	if called {
		t.Errorf("CloseSession() did not forget the session's connections")
	}
}
