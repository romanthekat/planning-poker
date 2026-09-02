package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/romanthekat/planning-poker/internal/poker"
	"github.com/romanthekat/planning-poker/internal/storage/memory"
	httptransport "github.com/romanthekat/planning-poker/internal/transport/http"
)

// newTestServer wires a full Application (real in-memory store, real Hub, real
// SessionService) behind an httptest.Server, so these tests exercise the HTTP layer
// exactly as it is wired in cmd/web/main.go, end to end through routing/decoding/
// validation, without mocking any collaborator.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	discard := log.New(io.Discard, "", 0)
	store := memory.NewStore()
	hub := poker.NewHub(discard, discard)
	sessionService := poker.NewSessionService(store, hub, discard, discard)
	app := httptransport.NewApplication(sessionService, discard, discard)

	srv := httptest.NewServer(app.Routes())
	t.Cleanup(srv.Close)
	return srv
}

func postJSON(t *testing.T, url string, body any) *http.Response {
	t.Helper()

	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(b)
	}

	resp, err := http.Post(url, "application/json", reader)
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func createSession(t *testing.T, srv *httptest.Server) poker.Session {
	t.Helper()

	resp := postJSON(t, srv.URL+"/api/sessions", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("createSession() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var session poker.Session
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	return session
}

func joinSession(t *testing.T, srv *httptest.Server, sessionId poker.SessionId, name string) poker.User {
	t.Helper()

	url := fmt.Sprintf("%s/api/sessions/%d/join", srv.URL, sessionId)
	resp := postJSON(t, url, poker.CreateUserRequest{Name: name})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("joinSession() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var user poker.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		t.Fatalf("decode user: %v", err)
	}
	return user
}

func TestCreateSession(t *testing.T) {
	srv := newTestServer(t)

	session := createSession(t, srv)

	if session.Id == 0 {
		t.Errorf("createSession() got zero-value session id")
	}
	if !session.VotesHidden {
		t.Errorf("createSession() VotesHidden = false, want true for a fresh session")
	}
}

func TestCheckSessionExists_Found(t *testing.T) {
	srv := newTestServer(t)
	session := createSession(t, srv)

	resp, err := http.Get(fmt.Sprintf("%s/api/sessions/%d?userId=%d", srv.URL, session.Id, poker.UserId(42)))
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("checkSessionExists() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestCheckSessionExists_NotFound(t *testing.T) {
	srv := newTestServer(t)

	resp, err := http.Get(srv.URL + "/api/sessions/999999?userId=999999")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("checkSessionExists() status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestCheckSessionExists_BadRequest(t *testing.T) {
	srv := newTestServer(t)

	resp, err := http.Get(srv.URL + "/api/sessions/not-a-number")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("checkSessionExists() status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestJoinSession_Success(t *testing.T) {
	srv := newTestServer(t)
	session := createSession(t, srv)

	user := joinSession(t, srv, session.Id, "Alice")

	if user.Name != "Alice" {
		t.Errorf("joinSession() user.Name = %v, want Alice", user.Name)
	}
	if user.Id == 0 {
		t.Errorf("joinSession() got zero-value user id")
	}
}

func TestJoinSession_InvalidBody(t *testing.T) {
	srv := newTestServer(t)
	session := createSession(t, srv)

	url := fmt.Sprintf("%s/api/sessions/%d/join", srv.URL, session.Id)
	resp := postJSON(t, url, poker.CreateUserRequest{Name: ""})

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("joinSession() with empty name status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestJoinSession_SessionNotFound(t *testing.T) {
	srv := newTestServer(t)

	url := srv.URL + "/api/sessions/999999/join"
	resp := postJSON(t, url, poker.CreateUserRequest{Name: "Alice"})

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("joinSession() for missing session status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestVote_Success(t *testing.T) {
	srv := newTestServer(t)
	session := createSession(t, srv)
	user := joinSession(t, srv, session.Id, "Alice")

	url := fmt.Sprintf("%s/api/sessions/%d/vote", srv.URL, session.Id)
	resp := postJSON(t, url, poker.VoteRequest{UserId: user.Id, Vote: "5"})

	if resp.StatusCode != http.StatusOK {
		t.Errorf("vote() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestVote_MissingVoteField(t *testing.T) {
	srv := newTestServer(t)
	session := createSession(t, srv)
	user := joinSession(t, srv, session.Id, "Alice")

	url := fmt.Sprintf("%s/api/sessions/%d/vote", srv.URL, session.Id)
	resp := postJSON(t, url, poker.VoteRequest{UserId: user.Id, Vote: ""})

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("vote() with empty vote status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestVote_UserNotFound(t *testing.T) {
	srv := newTestServer(t)
	session := createSession(t, srv)

	url := fmt.Sprintf("%s/api/sessions/%d/vote", srv.URL, session.Id)
	resp := postJSON(t, url, poker.VoteRequest{UserId: poker.UserId(123), Vote: "5"})

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("vote() for unknown user status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestShow_Success(t *testing.T) {
	srv := newTestServer(t)
	session := createSession(t, srv)
	user := joinSession(t, srv, session.Id, "Alice")

	url := fmt.Sprintf("%s/api/sessions/%d/show", srv.URL, session.Id)
	resp := postJSON(t, url, poker.UserRequest{UserId: user.Id})

	if resp.StatusCode != http.StatusOK {
		t.Errorf("show() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestShow_UserNotFound(t *testing.T) {
	srv := newTestServer(t)
	session := createSession(t, srv)

	url := fmt.Sprintf("%s/api/sessions/%d/show", srv.URL, session.Id)
	resp := postJSON(t, url, poker.UserRequest{UserId: poker.UserId(123)})

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("show() for unknown user status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestClear_Success(t *testing.T) {
	srv := newTestServer(t)
	session := createSession(t, srv)
	user := joinSession(t, srv, session.Id, "Alice")

	url := fmt.Sprintf("%s/api/sessions/%d/clear", srv.URL, session.Id)
	resp := postJSON(t, url, poker.UserRequest{UserId: user.Id})

	if resp.StatusCode != http.StatusOK {
		t.Errorf("clear() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestClear_UserNotFound(t *testing.T) {
	srv := newTestServer(t)
	session := createSession(t, srv)

	url := fmt.Sprintf("%s/api/sessions/%d/clear", srv.URL, session.Id)
	resp := postJSON(t, url, poker.UserRequest{UserId: poker.UserId(123)})

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("clear() for unknown user status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}
