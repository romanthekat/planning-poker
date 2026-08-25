package poker

import (
	"sync"
	"time"
)

type SessionId int
type UserId int

// Session holds a pure domain state for a single planning-poker session.
// Relevant live websocket connections belong to Hub,
// so Session has no dependency on any transport/websocket library and trivially JSON-marshalable.
//
// Every read or write, from any package, must happen while holding the session's own lock via Mutex().
type Session struct {
	Id          SessionId         `json:"id"`
	Users       map[UserId]*User  `json:"-"`
	Votes       map[UserId]string `json:"-"`
	VotesInfo   []Vote            `json:"votes_info"`
	VotesHidden bool              `json:"votes_hidden"`
	LastActive  time.Time         `json:"-"`

	mu *sync.RWMutex `json:"-"`
}

// NewSession creates a new Session with all of its maps/lock initialized.
func NewSession(id SessionId) *Session {
	return &Session{
		Id:          id,
		Users:       make(map[UserId]*User),
		Votes:       make(map[UserId]string),
		VotesInfo:   []Vote{},
		VotesHidden: true,
		LastActive:  time.Now(),
		mu:          &sync.RWMutex{},
	}
}

// Mutex returns the lock that must be held (via Lock/Unlock, or RLock/RUnlock for
// read-only access) around any read or write of this session's mutable fields.
func (s *Session) Mutex() *sync.RWMutex {
	return s.mu
}

type CreateUserRequest struct {
	Name string `json:"name" validate:"min=1,max=20"`
}

type User struct {
	Id         UserId    `json:"id"`
	Name       string    `json:"name"`
	LastActive time.Time `json:"last_active"`
	Active     bool      `json:"active"`
}

// TODO consider getting user id from header, not body
type UserRequest struct {
	UserId UserId `json:"user_id" validate:"required"`
}

type VoteRequest struct {
	UserId UserId `json:"user_id" validate:"required"`
	Vote   string `json:"vote" validate:"required"`
}

type Vote struct {
	Name        string `json:"name"`
	Voted       bool   `json:"is_voted"`
	Vote        string `json:"vote"`
	CurrentUser bool   `json:"is_current_user"`
}

type VotesByName []Vote

func (v VotesByName) Len() int {
	return len(v)
}

func (v VotesByName) Less(i, j int) bool {
	return v[i].Name < v[j].Name
}

func (v VotesByName) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}
