package models

import (
	"errors"
	"time"
)

var ErrNoRecord = errors.New("models: no matching record found")

type SessionId int
type UserId int

//Session
type Session struct {
	Id          SessionId           `json:"id"`
	Users       map[UserId]*User    `json:"users"`
	Votes       map[UserId]*float32 `json:"-"`
	VotesInfo   map[string]string   `json:"votes_info"`
	VotesHidden bool                `json:"votes_hidden"`
	LastActive  time.Time
}

type User struct {
	Id         UserId `json:"id"`
	Name       string `json:"name"`
	lastActive time.Time
}

type Vote struct {
	UserId UserId  `json:"user_id"`
	Vote   float32 `json:"vote"`
}

//SessionModel defines model/DAO methods for Session
type SessionModel interface {
	Create() (*Session, error)
	Get(id SessionId) (*Session, error)
	Remove(id SessionId) (int64, error)
}
