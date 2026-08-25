package http_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/romanthekat/planning-poker/internal/poker"
)

func TestGetWebsocketConnection_Success(t *testing.T) {
	srv := newTestServer(t)
	session := createSession(t, srv)
	user := joinSession(t, srv, session.Id, "Alice")

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	url := fmt.Sprintf("%s/api/sessions/%d/get/%d", wsURL, session.Id, user.Id)

	dialCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(dialCtx, url, nil)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer conn.CloseNow()

	// Connecting triggers SaveConnectionForUser, which immediately pushes a session
	// snapshot to this new connection.
	var got poker.Session
	readCtx, cancelRead := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelRead()
	if err := wsjson.Read(readCtx, conn, &got); err != nil {
		t.Fatalf("read update: %v", err)
	}

	if got.Id != session.Id {
		t.Errorf("first update session.Id = %v, want %v", got.Id, session.Id)
	}
}

func TestGetWebsocketConnection_UnknownUser_ConnectionIsClosed(t *testing.T) {
	srv := newTestServer(t)
	session := createSession(t, srv)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	url := fmt.Sprintf("%s/api/sessions/%d/get/%d", wsURL, session.Id, poker.UserId(123))

	dialCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(dialCtx, url, nil)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer conn.CloseNow()

	// The handler upgrades the connection to a websocket before it can know the user
	// doesn't exist, so the negative response can no longer be a plain HTTP status -
	// the observable effect for the client is simply that it never receives an update.
	readCtx, cancelRead := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancelRead()
	_, _, err = conn.Read(readCtx)
	if err == nil {
		t.Errorf("Read() on a connection for an unknown user = nil error, want no update to ever be sent")
	}
}
