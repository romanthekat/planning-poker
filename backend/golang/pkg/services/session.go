package services

import (
	"github.com/gorilla/websocket"
	"github.com/romanthekat/planning-poker/pkg/models"
	"html"
	"io"
	"log"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"
)

const UserIdMaxValue = 420_000

const pongWait = 5 * time.Second
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

	s.updateUserActiveness(user)

	session, err := s.Get(sessionId)
	if err != nil {
		return nil, err
	}

	user.Id = models.UserId(GenerateRandomId())

	session.Users[user.Id] = user

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

	session, err := s.sessions.Create()
	go func() {
		ticker := time.NewTicker(pingPeriod)

		for {
			select {
			case <-session.ExpirationChan:
				break
			case <-ticker.C:
				for userId, conn := range session.Connections {
					err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(pingPeriod))
					if err != nil {
						s.errorLog.Printf("ping error: %s", err)
						delete(session.Connections, userId)
					}
				}
			}
		}
	}()

	return session, err
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
	defer s.mutex.Unlock()

	session, err := s.Get(sessionId)
	if err != nil {
		return err
	}

	_, ok := session.Users[userId]
	if !ok {
		return models.ErrNoRecord
	}

	defer s.SendUpdates(sessionId)

	existingConn, ok := session.Connections[userId]
	if ok {
		existingConn.Close()
	}
	session.Connections[userId] = conn

	//naive reader from connection until error happens, otherwise pong handler won't work
	go func(c *websocket.Conn) {
		for {
			messageType, reader, err := c.NextReader()
			s.infoLog.Println("websocket messageType: ", messageType)
			if err != nil {
				c.Close()
				break
			}

			buf := new(strings.Builder)
			_, err = io.Copy(buf, reader)
			if err != nil {
				c.Close()
				break
			}

			s.infoLog.Println("websocket message: ", buf.String())
		}
	}(conn)

	conn.SetPongHandler(func(appData string) error {
		user, ok := session.Users[userId]
		if ok {
			s.updateUserActiveness(user)
		}

		//conn.SetReadDeadline(time.Now().Add(pongWait));

		return nil
	})

	return nil
}

func (s SessionService) GetMaskedSessionForUser(session models.Session, userId models.UserId) models.Session {
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
			s.errorLog.Printf("error for session %v user %v: %s\n", sessionId, userId, err)
		} else {
			user, ok := session.Users[userId]
			if ok {
				s.updateUserActiveness(user)
			}
		}
	}

	return nil
}

func (s SessionService) updateUserActiveness(user *models.User) {
	user.LastActive = time.Now()
	user.Active = true
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
