package services

import (
	"math/rand"
	"rgm-planning-poker/pkg/models"
	"sync"
)

const UserIdMaxValue = 420_000

type SessionService struct {
	mutex *sync.Mutex
}

func NewSessionService() *SessionService {
	return &SessionService{&sync.Mutex{}}
}

func (s SessionService) JoinSession(session *models.Session, user *models.User) *models.User {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	user.Id = models.UserId(GenerateRandomId())

	session.Users[user.Id] = user

	return user
}

func (s SessionService) Vote(session *models.Session, vote *models.Vote) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	user, ok := session.Users[vote.UserId]
	if !ok {
		return models.ErrNoRecord
	}

	session.Votes[user.Id] = vote.Vote

	return nil
}

func (s SessionService) Clear(session *models.Session) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for v := range session.Votes {
		delete(session.Votes, v)
	}
}

func GenerateRandomId() int {
	return rand.Intn(UserIdMaxValue)
}
