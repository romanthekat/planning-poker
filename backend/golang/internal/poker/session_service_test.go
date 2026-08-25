package poker_test

import (
	"io"
	"log"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/romanthekat/planning-poker/internal/poker"
	"github.com/romanthekat/planning-poker/internal/storage/memory"
)

// TestConcurrentSessionAccess sets up a scenario which used to be a data race: before
// the concurrency fix, SessionService methods (JoinSession/Vote/Show/Clear) mutated a
// Session's Users/Votes maps from request-handling goroutines while the storage layer's
// background expireUsers goroutine mutated the very same maps directly, guarded by an
// unrelated lock. Now that expiration lives in SessionService itself and every access
// goes through Session's own lock, this exercises exactly that overlap.
//
// To check with race condition validation, use:
//
//	go test -race ./internal/poker/...
func TestConcurrentSessionAccess(t *testing.T) {
	discard := log.New(io.Discard, "", 0)
	store := memory.NewStore()
	hub := poker.NewHub(discard, discard)
	service := poker.NewSessionService(store, hub, discard, discard)

	session, err := service.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Run long enough to overlap with at least one tick of the background
	// expireUsers goroutine (every second), which is the concurrent mutator that
	// raced with SessionService in the original bug.
	deadline := time.Now().Add(1500 * time.Millisecond)

	const workers = 20
	var wg sync.WaitGroup
	for i := range workers {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()

			user, err := service.JoinSession(session.Id, &poker.User{Name: "worker-" + strconv.Itoa(worker)})
			if err != nil {
				// The session may already be gone if a background job removed it
				// concurrently; that's a valid outcome for this stress test, not a failure.
				return
			}

			for time.Now().Before(deadline) {
				_ = service.Vote(session.Id, &poker.VoteRequest{UserId: user.Id, Vote: "1"})
				_ = service.Show(session.Id, user.Id)
				_ = service.Clear(session.Id, user.Id)
			}
		}(i)
	}

	wg.Wait()
}
