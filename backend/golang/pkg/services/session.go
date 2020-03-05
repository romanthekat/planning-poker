package services

import (
	"math/rand"
	"rgm-planning-poker/pkg/models"
	"sync"
)

const UserIdMaxValue = 420_000

type SessionService struct {
	sessions models.SessionModel
	mutex    *sync.Mutex
}

func NewSessionService(sessions models.SessionModel) *SessionService {
	return &SessionService{sessions, &sync.Mutex{}}
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

func (s SessionService) Create() (*models.Session, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.sessions.Create()
}

func (s SessionService) Get(id models.SessionId) (*models.Session, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.sessions.Get(id)
}

func (s SessionService) GetMaskedSessionForUser(session *models.Session, id models.UserId) *models.Session {
	//TODO impl
	return session
}

func GenerateRandomId() int {
	return rand.Intn(UserIdMaxValue)
}
