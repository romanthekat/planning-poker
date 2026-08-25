package memory

import (
	"math/rand"
	"sync"

	"github.com/romanthekat/planning-poker/internal/poker"
)

const MaxSessionId = 420_000

// Store is a pure in-memory CRUD adapter for poker.Session.
// Implements poker.sessionStore
type Store struct {
	mutex    sync.Mutex
	sessions map[poker.SessionId]*poker.Session
}

func NewStore() *Store {
	return &Store{
		sessions: make(map[poker.SessionId]*poker.Session),
	}
}

func (s *Store) Create() (*poker.Session, error) {
	id := poker.SessionId(generateRandomId())
	session := poker.NewSession(id)

	s.mutex.Lock()
	s.sessions[id] = session
	s.mutex.Unlock()

	return session, nil
}

func (s *Store) Get(id poker.SessionId) (*poker.Session, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	session, ok := s.sessions[id]
	if !ok {
		return nil, poker.ErrNoRecord
	}

	return session, nil
}

func (s *Store) Delete(id poker.SessionId) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.sessions, id)

	return nil
}

// List returns a snapshot of every currently stored session; used by the domain layer
// to drive its own expiration checks (see poker.SessionService).
func (s *Store) List() []*poker.Session {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	sessions := make([]*poker.Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

func generateRandomId() int {
	return rand.Intn(MaxSessionId)
}
