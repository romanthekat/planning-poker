package services

import (
	"io"
	"log"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/romanthekat/planning-poker/pkg/models"
	"github.com/romanthekat/planning-poker/pkg/models/memory"
)

// TestConcurrentSessionAccess set up a scenario, which could be a data race.
// SessionService methods (JoinSession/Vote/Show/Clear) mutate a Session's Users/Votes maps
// from request-handling goroutines.
// Meanwhile, the storage layer's background expireUsers goroutine
// (started by memory.NewSessionModel, ticking every second) mutates the very same maps directly.
//
// To check with rece condition validation, use:
//
//	go test -race ./pkg/services/...
func TestConcurrentSessionAccess(t *testing.T) {
	discard := log.New(io.Discard, "", 0)
	sessionModel := memory.NewSessionModel()
	service := NewSessionService(sessionModel, discard, discard)

	session, err := service.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Run long enough to overlap with at least one tick of the background
	// expireUsers goroutine (every second), which is the concurrent mutator
	// that raced with SessionService in the original bug.
	deadline := time.Now().Add(1500 * time.Millisecond)

	const workers = 20
	var wg sync.WaitGroup
	for i := range workers {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()

			user, err := service.JoinSession(session.Id, &models.User{Name: "worker-" + strconv.Itoa(worker)})
			if err != nil {
				// The session may already be gone if a background job removed it
				// concurrently; that's a valid outcome for this stress test, not a failure.
				return
			}

			for time.Now().Before(deadline) {
				_ = service.Vote(session.Id, &models.VoteRequest{UserId: user.Id, Vote: "1"})
				_ = service.Show(session.Id, user.Id)
				_ = service.Clear(session.Id, user.Id)
			}
		}(i)
	}

	wg.Wait()
}
