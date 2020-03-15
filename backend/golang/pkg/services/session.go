package services

import (
	"fmt"
	"math/rand"
	"rgm-planning-poker/pkg/models"
	"sync"
	"time"
)

const UserIdMaxValue = 420_000

type SessionService struct {
	sessions models.SessionModel
	mutex    *sync.Mutex
}

func NewSessionService(sessions models.SessionModel) *SessionService {
	return &SessionService{sessions, &sync.Mutex{}}
}

func (s SessionService) JoinSession(sessionId models.SessionId, user *models.User) (*models.User, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	session, err := s.Get(sessionId)
	if err != nil {
		return nil, err
	}

	user.Id = models.UserId(GenerateRandomId())

	session.Users[user.Id] = user
	session.VotesHidden = true

	return user, nil
}

func (s SessionService) Vote(sessionId models.SessionId, vote *models.Vote) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	session, err := s.Get(sessionId)
	if err != nil {
		return err
	}

	user, ok := session.Users[vote.UserId]
	if !ok {
		return models.ErrNoRecord
	}

	session.Votes[user.Id] = &vote.Vote

	if len(session.Votes) == len(session.Users) {
		session.VotesHidden = false
	}

	return nil
}

func (s SessionService) Clear(sessionId models.SessionId) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	session, err := s.Get(sessionId)
	if err != nil {
		return err
	}

	for v := range session.Votes {
		delete(session.Votes, v)
	}

	session.VotesHidden = true

	return nil
}

func (s SessionService) Create() (*models.Session, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.sessions.Create()
}

func (s SessionService) Get(id models.SessionId) (*models.Session, error) {
	session, err := s.sessions.Get(id)
	if err != nil {
		return nil, err
	}

	session.LastActive = time.Now()

	return session, err
}

func (s SessionService) GetMaskedSessionForUser(session models.Session, userId models.UserId) models.Session {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	votesInfo := make(map[string]string)
	for votedUserId, vote := range session.Votes {
		name := session.Users[votedUserId].Name
		votesInfo[name] = getVoteToShow(vote, session.VotesHidden, votedUserId == userId)
	}

	session.VotesInfo = votesInfo
	return session
}

func (s SessionService) Show(sessionId models.SessionId) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	session, err := s.sessions.Get(sessionId)
	if err != nil {
		return err
	}

	session.VotesHidden = false

	return nil
}

func getVoteToShow(vote *float32, votesHidden bool, sameUser bool) string {
	if sameUser || !votesHidden {
		return fmt.Sprintf("%.2f", *vote)
	} else if vote != nil {
		return "+"
	} else {
		return "-"
	}
}

func GenerateRandomId() int {
	return rand.Intn(UserIdMaxValue)
}
