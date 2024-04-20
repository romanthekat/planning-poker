package memory

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/romanthekat/planning-poker/pkg/models"
	"math/rand"
	"sync"
	"time"
)

const SessionExpirationMin = 42.0
const UserExpirationSec = 20.0
const MaxSessionId = 420_000

type SessionModel struct {
	sessions map[models.SessionId]*models.Session
	mutex    *sync.Mutex
}

func NewSessionModel() *SessionModel {
	sessionModel := &SessionModel{make(map[models.SessionId]*models.Session), &sync.Mutex{}}

	go removeExpiredSessions(sessionModel)
	go expireUsers(sessionModel)

	return sessionModel
}

func removeExpiredSessions(sessionModel *SessionModel) {
	ticker := time.NewTicker(time.Minute)

	for range ticker.C {
		sessionModel.mutex.Lock()

		for _, session := range sessionModel.sessions {
			if time.Since(session.LastActive).Minutes() > SessionExpirationMin {
				session.ExpirationChan <- struct{}{}

				for _, conn := range session.Connections {
					conn.Close()
				}
				delete(sessionModel.sessions, session.Id)
			}
		}

		sessionModel.mutex.Unlock()
	}
}

func expireUsers(sessionModel *SessionModel) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		sessionModel.mutex.Lock()

		for _, session := range sessionModel.sessions {
			for _, user := range session.Users {
				if time.Since(user.LastActive).Seconds() > UserExpirationSec && user.Active {
					fmt.Printf("expire user: %+v, session: %d\n", user, session.Id)
					user.Active = false

					delete(session.Connections, user.Id)
					delete(session.Users, user.Id)
					//TODO check whether session votes must be shown/all active users voted
				}
			}
		}

		sessionModel.mutex.Unlock()
	}
}

func (s SessionModel) Create() (*models.Session, error) {
	id := models.SessionId(generateRandomId())

	session := &models.Session{
		Id:             id,
		Users:          make(map[models.UserId]*models.User),
		Votes:          make(map[models.UserId]*float32),
		VotesInfo:      []models.VoteInfo{},
		VotesHidden:    true,
		LastActive:     time.Now(),
		Connections:    make(map[models.UserId]*websocket.Conn),
		ExpirationChan: make(chan interface{}, 1),
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
	return rand.Intn(MaxSessionId)
}

// TODO usage looks ugly
func (s SessionModel) update(callback func()) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	callback()
}
