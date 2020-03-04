package services

import (
	"math/rand"
	"rgm-planning-poker/pkg/models"
	"sync"
)

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

func GenerateRandomId() int {
	return rand.Intn(100000)
}
