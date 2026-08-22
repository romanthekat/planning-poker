package models

import (
	"errors"
	"time"

	"github.com/coder/websocket"
)

var ErrNoRecord = errors.New("models: no matching record found")

type SessionId int
type UserId int

type Session struct {
	Id             SessionId                  `json:"id"`
	Users          map[UserId]*User           `json:"-"`
	Votes          map[UserId]string          `json:"-"`
	VotesInfo      []Vote                     `json:"votes_info"`
	VotesHidden    bool                       `json:"votes_hidden"`
	LastActive     time.Time                  `json:"-"`
	Connections    map[UserId]*websocket.Conn `json:"-"`
	ExpirationChan chan any                   `json:"-"`
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
	UserId UserId `json:"user_id"`
}

type VoteRequest struct {
	UserId UserId `json:"user_id"`
	Vote   string `json:"vote"`
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

// SessionModel defines model/DAO methods for Session
type SessionModel interface {
	Create() (*Session, error)
	Get(id SessionId) (*Session, error)
	Remove(id SessionId) (int64, error)
}
