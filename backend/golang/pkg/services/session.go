package services

import (
	"context"
	"html"
	"io"
	"log"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/romanthekat/planning-poker/pkg/models"
)

const UserIdMaxValue = 420_000

const pongWait = 5 * time.Second
const pingPeriod = (pongWait * 9) / 10

type SessionService struct {
	sessions models.SessionModel
	errorLog *log.Logger
	infoLog  *log.Logger
}

func NewSessionService(sessions models.SessionModel, errorLog *log.Logger, infoLog *log.Logger) *SessionService {
	return &SessionService{sessions, errorLog, infoLog}
}

func (s SessionService) JoinSession(sessionId models.SessionId, user *models.User) (*models.User, error) {
	session, err := s.Get(sessionId)
	if err != nil {
		return nil, err
	}

	session.Mutex().Lock()

	s.UpdateUserActiveness(user)
	user.Id = models.UserId(GenerateRandomId())
	session.Users[user.Id] = user

	session.Mutex().Unlock()

	err = s.SendUpdates(sessionId)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s SessionService) Vote(sessionId models.SessionId, vote *models.VoteRequest) error {
	session, err := s.Get(sessionId)
	if err != nil {
		return err
	}

	session.Mutex().Lock()
	user, ok := session.Users[vote.UserId]
	if !ok {
		session.Mutex().Unlock()
		return models.ErrNoRecord
	}

	session.Votes[user.Id] = vote.Vote

	if s.allVotesObtained(session) {
		session.VotesHidden = false
	}
	session.Mutex().Unlock()

	//TODO controversial to send updates here; side-effect needed, but who's responsible?
	return s.SendUpdates(sessionId)
}

func (s SessionService) allVotesObtained(session *models.Session) bool {
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

	return activeUsersVotesCount == activeUsersCount
}

func (s SessionService) Clear(sessionId models.SessionId, userId models.UserId) error {
	session, err := s.Get(sessionId)
	if err != nil {
		return err
	}

	session.Mutex().Lock()
	_, ok := session.Users[userId]
	if !ok {
		session.Mutex().Unlock()
		return models.ErrNoRecord
	}

	for v := range session.Votes {
		delete(session.Votes, v)
	}

	session.VotesHidden = true
	session.Mutex().Unlock()

	return s.SendUpdates(sessionId)
}

func (s SessionService) Create() (*models.Session, error) {
	session, err := s.sessions.Create()
	if err != nil {
		return nil, err
	}

	go s.tickerFunctionForSession(session)()

	return session, nil
}

func (s SessionService) tickerFunctionForSession(session *models.Session) func() {
	return func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()

		for {
			select {
			case <-session.ExpirationChan:
				return
			case <-ticker.C:
				session.Mutex().Lock()
				for userId, conn := range session.Connections {
					err := conn.Ping(context.Background())
					if err != nil {
						s.errorLog.Printf("ping error: %s", err)
						delete(session.Connections, userId)
					}
				}
				session.Mutex().Unlock()
			}
		}
	}
}

func (s SessionService) Get(id models.SessionId) (*models.Session, error) {
	session, err := s.sessions.Get(id)
	if err != nil {
		return nil, err
	}

	session.Mutex().Lock()
	session.LastActive = time.Now()
	session.Mutex().Unlock()

	return session, nil
}

func (s SessionService) SaveConnectionForUser(sessionId models.SessionId, userId models.UserId, conn *websocket.Conn) error {
	session, err := s.Get(sessionId)
	if err != nil {
		return err
	}

	session.Mutex().Lock()
	_, ok := session.Users[userId]
	if !ok {
		session.Mutex().Unlock()
		return models.ErrNoRecord
	}

	existingConn, ok := session.Connections[userId]
	if ok {
		_ = existingConn.Close(websocket.StatusNormalClosure, "")
	}
	session.Connections[userId] = conn
	session.Mutex().Unlock()

	//naive reader from connection until error happens, otherwise pong handler won't work
	go s.websocketReaderFunction()(conn)

	return s.SendUpdates(sessionId)
}

func (s SessionService) websocketReaderFunction() func(c *websocket.Conn) {
	return func(c *websocket.Conn) {
		for {
			messageType, reader, err := c.Reader(context.Background())
			s.infoLog.Println("websocket messageType: ", messageType)
			if err != nil {
				s.errorLog.Println(err)
				_ = c.Close(websocket.StatusInternalError, "internal error")
				break
			}

			buf := new(strings.Builder)
			_, err = io.Copy(buf, reader)
			if err != nil {
				s.errorLog.Println(err)
				_ = c.Close(websocket.StatusInternalError, "internal error")
				break
			}

			s.infoLog.Println("websocket message: ", buf.String())
		}
	}
}

func (s SessionService) GetMaskedSessionForUser(session models.Session, userId models.UserId) models.Session {
	var votes []models.Vote

	for displayUserId, user := range session.Users {
		if !user.Active {
			continue
		}

		userVote := session.Votes[displayUserId]
		isCurrentUser := displayUserId == userId

		vote := models.Vote{
			Name:        html.EscapeString(user.Name),
			Voted:       userVote != "",
			Vote:        getVoteToShow(userVote, session.VotesHidden, isCurrentUser),
			CurrentUser: isCurrentUser,
		}

		votes = append(votes, vote)
	}

	sort.Sort(models.VotesByName(votes))
	session.VotesInfo = votes
	return session
}

func (s SessionService) Show(sessionId models.SessionId, userId models.UserId) error {
	session, err := s.sessions.Get(sessionId)
	if err != nil {
		return err
	}

	session.Mutex().Lock()
	_, ok := session.Users[userId]
	if !ok {
		session.Mutex().Unlock()
		return models.ErrNoRecord
	}

	session.VotesHidden = false
	session.Mutex().Unlock()

	return s.SendUpdates(sessionId)
}

func (s SessionService) SendUpdates(sessionId models.SessionId) error {
	s.infoLog.Printf("[%v] send updates for session", sessionId)

	session, err := s.sessions.Get(sessionId)
	if err != nil {
		s.errorLog.Println(err)
		return err
	}

	// Snapshot per-viewer masked views and the target connections while holding a
	// read lock, then release it before doing network I/O (wsjson.Write) - a slow or
	// stuck client must not be able to block every other request for this session.
	type recipient struct {
		userId  models.UserId
		conn    *websocket.Conn
		session models.Session
	}

	session.Mutex().RLock()

	recipients := make([]recipient, 0, len(session.Connections))
	for userId, conn := range session.Connections {
		recipients = append(recipients, recipient{userId, conn, s.GetMaskedSessionForUser(*session, userId)})
	}

	session.Mutex().RUnlock()

	for _, r := range recipients {
		err = wsjson.Write(context.Background(), r.conn, r.session)
		if err != nil {
			s.errorLog.Printf("[%v] user %v: %s\n", sessionId, r.userId, err)
			continue
		}

		session.Mutex().Lock()
		if user, ok := session.Users[r.userId]; ok {
			s.UpdateUserActiveness(user)
		}
		session.Mutex().Unlock()
	}

	return nil
}

func (s SessionService) UpdateUserActiveness(user *models.User) {
	user.LastActive = time.Now()
	user.Active = true
}

func getVoteToShow(vote string, votesHidden bool, sameUser bool) string {
	if sameUser || !votesHidden {
		return vote
	}

	return ""
}

func GenerateRandomId() int {
	return rand.Intn(UserIdMaxValue)
}
