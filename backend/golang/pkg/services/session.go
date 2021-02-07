package services

import (
	"github.com/EvilKhaosKat/planning-poker/pkg/models"
	"github.com/gorilla/websocket"
	"html"
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"
)

const UserIdMaxValue = 420_000
const pongWait = 60 * time.Second

const pingPeriod = (pongWait * 9) / 10

type SessionService struct {
	sessions models.SessionModel
	mutex    *sync.Mutex
	errorLog *log.Logger
	infoLog  *log.Logger
}

func NewSessionService(sessions models.SessionModel, errorLog *log.Logger, infoLog *log.Logger) *SessionService {
	return &SessionService{sessions, &sync.Mutex{}, errorLog, infoLog}
}

func (s SessionService) JoinSession(sessionId models.SessionId, user *models.User) (*models.User, error) {
	s.mutex.Lock()
	defer s.SendUpdates(sessionId)
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
	defer s.SendUpdates(sessionId)
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

	//TODO that's ugly and needs tests
	activeUsersCount := 0
	for _, user := range session.Users {
		if user.Active {
			activeUsersCount++
		}
	}

	activeUsersVotesCount := 0
	for userId := range session.Votes {
		if session.Users[userId].Active {
			activeUsersVotesCount++
		}
	}

	if activeUsersVotesCount == activeUsersCount {
		session.VotesHidden = false
	}

	return nil
}

func (s SessionService) Clear(sessionId models.SessionId) error {
	s.mutex.Lock()
	defer s.SendUpdates(sessionId)
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

func (s SessionService) SaveConnectionForUser(sessionId models.SessionId, userId models.UserId, conn *websocket.Conn) error {
	s.mutex.Lock()
	defer s.SendUpdates(sessionId)
	defer s.mutex.Unlock()

	session, err := s.Get(sessionId)
	if err != nil {
		return err
	}
	//TODO validate user id existence

	existingConn, ok := session.Connections[userId]
	if ok {
		existingConn.Close()
	}
	session.Connections[userId] = conn
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	s.SendUpdates(sessionId)

	return nil
}

func (s SessionService) GetMaskedSessionForUser(session models.Session, userId models.UserId) models.Session {
	user, ok := session.Users[userId]
	if ok {
		user.LastActive = time.Now()
		user.Active = true
	}

	var votesInfo []models.VoteInfo

	for displayUserId, user := range session.Users {
		if !user.Active {
			continue
		}

		userVote := session.Votes[displayUserId]
		isCurrentUser := displayUserId == userId

		voteInfo := models.VoteInfo{
			Name:        html.EscapeString(user.Name),
			Voted:       userVote != nil,
			Vote:        getVoteToShow(userVote, session.VotesHidden, isCurrentUser),
			CurrentUser: isCurrentUser,
		}

		votesInfo = append(votesInfo, voteInfo)
	}

	sort.Sort(models.VotesInfoByName(votesInfo))
	session.VotesInfo = votesInfo
	return session
}

func (s SessionService) Show(sessionId models.SessionId) error {
	s.mutex.Lock()
	defer s.SendUpdates(sessionId)
	defer s.mutex.Unlock()

	session, err := s.sessions.Get(sessionId)
	if err != nil {
		return err
	}

	session.VotesHidden = false

	return nil
}

func (s SessionService) SendUpdates(sessionId models.SessionId) error {
	s.infoLog.Printf("send updates for session %v\n", sessionId)

	session, err := s.sessions.Get(sessionId)
	if err != nil {
		s.errorLog.Println(err)
		return err
	}

	for userId, conn := range session.Connections {
		sessionToReturn := s.GetMaskedSessionForUser(*session, userId)
		err = conn.WriteJSON(sessionToReturn)
		if err != nil {
			s.errorLog.Printf("error for session %s user %s: %s\n", sessionId, userId, err)
		}
	}

	return nil
}

func getVoteToShow(vote *float32, votesHidden bool, sameUser bool) *float32 {
	if sameUser || !votesHidden {
		return vote
	} else {
		return nil
	}
}

func GenerateRandomId() int {
	return rand.Intn(UserIdMaxValue)
}
