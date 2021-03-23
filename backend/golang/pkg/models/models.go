package models

import (
	"errors"
	"github.com/gorilla/websocket"
	"time"
)

var ErrNoRecord = errors.New("models: no matching record found")

type SessionId int
type UserId int

//Session
type Session struct {
	Id             SessionId           `json:"id"`
	Users          map[UserId]*User    `json:"-"`
	Votes          map[UserId]*float32 `json:"-"`
	VotesInfo      []VoteInfo          `json:"votes_info"`
	VotesHidden    bool                `json:"votes_hidden"`
	LastActive     time.Time
	Connections    map[UserId]*websocket.Conn `json:"-"`
	ExpirationChan chan interface{}           `json:"-"`
}

type User struct {
	Id         UserId    `json:"id"`
	Name       string    `json:"name" validate:"min=1,max=20"`
	LastActive time.Time `json:"last_active"`
	Active     bool      `json:"active"`
}

type Vote struct {
	UserId UserId  `json:"user_id"`
	Vote   float32 `json:"vote"`
}

type VoteInfo struct {
	Name        string   `json:"name"`
	Voted       bool     `json:"is_voted"`
	Vote        *float32 `json:"vote"`
	CurrentUser bool     `json:"is_current_user"`
}

type VotesInfoByName []VoteInfo

func (v VotesInfoByName) Len() int {
	return len(v)
}

func (v VotesInfoByName) Less(i, j int) bool {
	return v[i].Name < v[j].Name
}

func (v VotesInfoByName) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

//SessionModel defines model/DAO methods for Session
type SessionModel interface {
	Create() (*Session, error)
	Get(id SessionId) (*Session, error)
	Remove(id SessionId) (int64, error)
}
