package memory

import (
	"math/rand"
	"rgm-planning-poker/pkg/models"
	"sync"
)

type SessionModel struct {
	sessions map[models.SessionId]*models.Session
	mutex    *sync.Mutex
}

func NewSessionModel() *SessionModel {
	return &SessionModel{make(map[models.SessionId]*models.Session), &sync.Mutex{}}
}

func (s SessionModel) Create() (*models.Session, error) {
	id := models.SessionId(generateRandomId())

	session := &models.Session{
		Id:          id,
		Users:       map[models.UserId]*models.User{},
		Votes:       make(map[models.UserId]float32),
		VotesHidden: false,
	}

	s.update(func() { s.sessions[id] = session })

	return session, nil
}

func (s SessionModel) Get(id models.SessionId) (*models.Session, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	session, ok := s.sessions[id]
	if !ok {
		return nil, models.ErrNoRecord
	}

	return session, nil
}

func (s SessionModel) Remove(id models.SessionId) (int64, error) {
	s.update(func() { delete(s.sessions, id) })
	return 1, nil
}

func generateRandomId() int {
	return rand.Intn(100000)
}

//TODO usage looks ugly
func (s SessionModel) update(callback func()) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	callback()
}