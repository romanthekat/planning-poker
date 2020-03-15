package memory

import (
	"math/rand"
	"rgm-planning-poker/pkg/models"
	"sync"
	"time"
)

const SessionExpirationMin = 42.0
const UserExpirationSec = 7.0

type SessionModel struct {
	sessions map[models.SessionId]*models.Session
	mutex    *sync.Mutex
}

func NewSessionModel() *SessionModel {
	sessionModel := &SessionModel{make(map[models.SessionId]*models.Session), &sync.Mutex{}}

	go removeExpiredSessions(sessionModel)
	go removeExpiredUsers(sessionModel)

	return sessionModel
}

func removeExpiredSessions(sessionModel *SessionModel) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sessionModel.mutex.Lock()

		for _, session := range sessionModel.sessions {
			if time.Now().Sub(session.LastActive).Minutes() > SessionExpirationMin {
				delete(sessionModel.sessions, session.Id)
			}
		}

		sessionModel.mutex.Unlock()
	}
}

func removeExpiredUsers(sessionModel *SessionModel) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		sessionModel.mutex.Lock()

		for _, session := range sessionModel.sessions {
			for _, user := range session.Users {
				if time.Now().Sub(user.LastActive).Seconds() > UserExpirationSec {
					delete(session.Users, user.Id)
					delete(session.Votes, user.Id)
					//TODO check whether session must be shown
				}
			}
		}

		sessionModel.mutex.Unlock()
	}
}

func (s SessionModel) Create() (*models.Session, error) {
	id := models.SessionId(generateRandomId())

	session := &models.Session{
		Id:          id,
		Users:       make(map[models.UserId]*models.User),
		Votes:       make(map[models.UserId]*float32),
		VotesHidden: true,
		LastActive:  time.Now(),
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
