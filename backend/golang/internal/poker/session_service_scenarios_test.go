package poker_test

import (
	"errors"
	"io"
	"log"
	"testing"

	"github.com/romanthekat/planning-poker/internal/poker"
	"github.com/romanthekat/planning-poker/internal/storage/memory"
)

// newTestService wires a SessionService against a fresh in-memory store and a Hub with
// no live connections, using discard loggers - enough for exercising the business logic
// covered by these tests without any real network I/O.
func newTestService() *poker.SessionService {
	discard := log.New(io.Discard, "", 0)
	store := memory.NewStore()
	hub := poker.NewHub(discard, discard)
	return poker.NewSessionService(store, hub, discard, discard)
}

func TestSessionService_JoinSession_Success(t *testing.T) {
	service := newTestService()
	session, err := service.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	user, err := service.JoinSession(session.Id, &poker.User{Name: "Alice"})
	if err != nil {
		t.Fatalf("JoinSession() error = %v", err)
	}

	if user.Name != "Alice" {
		t.Errorf("JoinSession() user.Name = %v, want Alice", user.Name)
	}
	if !user.Active {
		t.Errorf("JoinSession() user.Active = false, want true")
	}

	session.Mutex().RLock()
	_, ok := session.Users[user.Id]
	session.Mutex().RUnlock()
	if !ok {
		t.Errorf("JoinSession() did not add the user to the session")
	}
}

func TestSessionService_JoinSession_SessionNotFound(t *testing.T) {
	service := newTestService()

	_, err := service.JoinSession(poker.SessionId(999999), &poker.User{Name: "Alice"})
	if !errors.Is(err, poker.ErrNoRecord) {
		t.Errorf("JoinSession() error = %v, want %v", err, poker.ErrNoRecord)
	}
}

func TestSessionService_Vote_Success(t *testing.T) {
	service := newTestService()
	session, _ := service.Create()
	user, err := service.JoinSession(session.Id, &poker.User{Name: "Alice"})
	if err != nil {
		t.Fatalf("JoinSession() error = %v", err)
	}

	if err := service.Vote(session.Id, &poker.VoteRequest{UserId: user.Id, Vote: "5"}); err != nil {
		t.Fatalf("Vote() error = %v", err)
	}

	session.Mutex().RLock()
	vote, ok := session.Votes[user.Id]
	hidden := session.VotesHidden
	session.Mutex().RUnlock()

	if !ok || vote != "5" {
		t.Errorf("Vote() recorded = %q, ok=%v, want \"5\"", vote, ok)
	}
	// Alice is the only active user and she just voted, so all votes are in.
	if hidden {
		t.Errorf("Vote() VotesHidden = true after every active user voted, want false")
	}
}

func TestSessionService_Vote_WaitsForEveryActiveUser(t *testing.T) {
	service := newTestService()
	session, _ := service.Create()
	alice, _ := service.JoinSession(session.Id, &poker.User{Name: "Alice"})
	_, err := service.JoinSession(session.Id, &poker.User{Name: "Bob"})
	if err != nil {
		t.Fatalf("JoinSession() error = %v", err)
	}

	if err := service.Vote(session.Id, &poker.VoteRequest{UserId: alice.Id, Vote: "3"}); err != nil {
		t.Fatalf("Vote() error = %v", err)
	}

	session.Mutex().RLock()
	hidden := session.VotesHidden
	session.Mutex().RUnlock()

	if !hidden {
		t.Errorf("Vote() VotesHidden = false while Bob hasn't voted yet, want true")
	}
}

func TestSessionService_Vote_UserNotFound(t *testing.T) {
	service := newTestService()
	session, _ := service.Create()

	err := service.Vote(session.Id, &poker.VoteRequest{UserId: poker.UserId(123), Vote: "5"})
	if !errors.Is(err, poker.ErrNoRecord) {
		t.Errorf("Vote() error = %v, want %v", err, poker.ErrNoRecord)
	}
}

func TestSessionService_Vote_SessionNotFound(t *testing.T) {
	service := newTestService()

	err := service.Vote(poker.SessionId(999999), &poker.VoteRequest{UserId: poker.UserId(1), Vote: "5"})
	if !errors.Is(err, poker.ErrNoRecord) {
		t.Errorf("Vote() error = %v, want %v", err, poker.ErrNoRecord)
	}
}

func TestSessionService_Clear_Success(t *testing.T) {
	service := newTestService()
	session, _ := service.Create()
	user, _ := service.JoinSession(session.Id, &poker.User{Name: "Alice"})
	if err := service.Vote(session.Id, &poker.VoteRequest{UserId: user.Id, Vote: "5"}); err != nil {
		t.Fatalf("Vote() error = %v", err)
	}

	if err := service.Clear(session.Id, user.Id); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	session.Mutex().RLock()
	votesLeft := len(session.Votes)
	hidden := session.VotesHidden
	session.Mutex().RUnlock()

	if votesLeft != 0 {
		t.Errorf("Clear() left %d votes, want 0", votesLeft)
	}
	if !hidden {
		t.Errorf("Clear() VotesHidden = false, want true")
	}
}

func TestSessionService_Clear_UserNotFound(t *testing.T) {
	service := newTestService()
	session, _ := service.Create()

	err := service.Clear(session.Id, poker.UserId(123))
	if !errors.Is(err, poker.ErrNoRecord) {
		t.Errorf("Clear() error = %v, want %v", err, poker.ErrNoRecord)
	}
}

func TestSessionService_Show_Success(t *testing.T) {
	service := newTestService()
	session, _ := service.Create()
	user, _ := service.JoinSession(session.Id, &poker.User{Name: "Alice"})

	if err := service.Show(session.Id, user.Id); err != nil {
		t.Fatalf("Show() error = %v", err)
	}

	session.Mutex().RLock()
	hidden := session.VotesHidden
	session.Mutex().RUnlock()

	if hidden {
		t.Errorf("Show() VotesHidden = true, want false")
	}
}

func TestSessionService_Show_UserNotFound(t *testing.T) {
	service := newTestService()
	session, _ := service.Create()

	err := service.Show(session.Id, poker.UserId(123))
	if !errors.Is(err, poker.ErrNoRecord) {
		t.Errorf("Show() error = %v, want %v", err, poker.ErrNoRecord)
	}
}

func TestSessionService_GetMaskedSessionForUser_HidesOthersVotesUntilRevealed(t *testing.T) {
	service := newTestService()
	session, _ := service.Create()
	alice, _ := service.JoinSession(session.Id, &poker.User{Name: "Alice"})
	bob, _ := service.JoinSession(session.Id, &poker.User{Name: "Bob"})

	if err := service.Vote(session.Id, &poker.VoteRequest{UserId: bob.Id, Vote: "8"}); err != nil {
		t.Fatalf("Vote() error = %v", err)
	}

	session.Mutex().RLock()
	masked := service.GetMaskedSessionForUser(*session, alice.Id)
	session.Mutex().RUnlock()

	var bobVote poker.Vote
	found := false
	for _, v := range masked.VotesInfo {
		if v.Name == "Bob" {
			bobVote = v
			found = true
		}
	}
	if !found {
		t.Fatalf("GetMaskedSessionForUser() votes info = %+v, missing Bob", masked.VotesInfo)
	}
	if !bobVote.Voted {
		t.Errorf("GetMaskedSessionForUser() Bob.Voted = false, want true")
	}
	if bobVote.Vote != "" {
		t.Errorf("GetMaskedSessionForUser() from Alice's perspective revealed Bob's vote %q while votes are hidden, want empty", bobVote.Vote)
	}

	session.Mutex().RLock()
	maskedForBob := service.GetMaskedSessionForUser(*session, bob.Id)
	session.Mutex().RUnlock()

	for _, v := range maskedForBob.VotesInfo {
		if v.CurrentUser && v.Vote != "8" {
			t.Errorf("GetMaskedSessionForUser() hid the viewer's own vote: got %q, want \"8\"", v.Vote)
		}
	}
}

func TestSessionService_Create(t *testing.T) {
	service := newTestService()

	session, err := service.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if session.Users == nil || session.Votes == nil {
		t.Errorf("Create() session = %+v, want initialized Users/Votes maps", session)
	}
	if !session.VotesHidden {
		t.Errorf("Create() VotesHidden = false, want true for a fresh session")
	}
}
