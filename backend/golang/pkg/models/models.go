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
	Id          SessionId          `json:"id"`
	Users       map[UserId]*User   `json:"users"`
	Votes       map[UserId]float32 `json:"votes"`
	VotesHidden bool               `json:"votesHidden"`
}

type User struct {
	Id         UserId `json:"id"`
	Name       string `json:"name"`
	lastActive time.Time
}

//SessionModel defines model/DAO methods for Session
type SessionModel interface {
	Create() (*Session, error)
	Get(id SessionId) (*Session, error)
	Remove(id SessionId) (int64, error)
}
