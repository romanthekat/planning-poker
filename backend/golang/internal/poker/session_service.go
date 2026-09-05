package poker

import (
	"context"
	"html"
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const UserIdMaxValue = 420_000

const SessionExpirationMin = 42.0
const UserExpirationSec = 20.0

// SessionStore is the minimal persistence contract SessionService needs.
// Kept here to avoid cyclic dependencies between poker and storage/memory packages.
// TODO somewhat wonky, validate how to avoid dependency between poker 'logical' structures and storage/memory.
type sessionStore interface {
	Create() (*Session, error)
	Get(id SessionId) (*Session, error)
	Delete(id SessionId) error
	List() []*Session
}

type SessionService struct {
	store sessionStore
	hub   *Hub

	errorLog *log.Logger
	infoLog  *log.Logger

	// cancelExpiration lets removeExpiredSessions stop a session's ping loop (see Hub)
	// the moment the session itself is removed, instead of leaving it running until the
	// process exits.
	cancelMu         sync.Mutex
	cancelExpiration map[SessionId]func()
}

func NewSessionService(store sessionStore, hub *Hub, errorLog *log.Logger, infoLog *log.Logger) *SessionService {
	s := &SessionService{
		store:            store,
		hub:              hub,
		errorLog:         errorLog,
		infoLog:          infoLog,
		cancelExpiration: make(map[SessionId]func()),
	}

	go s.removeExpiredSessions()
	go s.expireUsers()

	return s
}

func (s *SessionService) JoinSession(sessionId SessionId, user *User) (*User, error) {
	session, err := s.Get(sessionId)
	if err != nil {
		return nil, err
	}

	session.Mutex().Lock()
	s.UpdateUserActiveness(sessionId, user)
	user.Id = UserId(GenerateRandomId())
	session.Users[user.Id] = user
	session.Mutex().Unlock()

	if err := s.SendUpdates(sessionId); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *SessionService) Vote(sessionId SessionId, vote *VoteRequest) error {
	session, err := s.Get(sessionId)
	if err != nil {
		return err
	}

	session.Mutex().Lock()
	user, ok := session.Users[vote.UserId]
	if !ok {
		session.Mutex().Unlock()
		return ErrNoRecord
	}

	session.Votes[user.Id] = vote.Vote

	if s.allVotesObtained(session) {
		session.VotesHidden = false
	}
	session.Mutex().Unlock()

	//TODO controversial to send updates here; side-effect needed, but who's responsible?
	return s.SendUpdates(sessionId)
}

// allVotesObtained assumes the caller already holds session's lock.
func (s *SessionService) allVotesObtained(session *Session) bool {
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

func (s *SessionService) Clear(sessionId SessionId, userId UserId) error {
	session, err := s.Get(sessionId)
	if err != nil {
		return err
	}

	session.Mutex().Lock()
	_, ok := session.Users[userId]
	if !ok {
		session.Mutex().Unlock()
		return ErrNoRecord
	}

	for v := range session.Votes {
		delete(session.Votes, v)
	}

	session.VotesHidden = true
	session.Mutex().Unlock()

	return s.SendUpdates(sessionId)
}

func (s *SessionService) Create() (*Session, error) {
	session, err := s.store.Create()
	if err != nil {
		return nil, err
	}

	s.startPingLoop(session.Id)

	return session, nil
}

func (s *SessionService) startPingLoop(sessionId SessionId) {
	ctx, cancel := context.WithCancel(context.Background())

	s.cancelMu.Lock()
	s.cancelExpiration[sessionId] = cancel
	s.cancelMu.Unlock()

	go s.hub.PingLoop(ctx, sessionId)
}

// removeExpiredSessions is a domain rule ("inactive sessions expire"), not a storage concern,
// so it lives here rather than in the storage layer;
// it only asks the store to persist the outcome (Delete), never reaches into another
// session's fields without holding that session's own lock.
func (s *SessionService) removeExpiredSessions() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		for _, session := range s.store.List() {
			session.Mutex().RLock()
			expired := time.Since(session.LastActive).Minutes() > SessionExpirationMin
			s.infoLog.Printf("[%v] session last active at %s\n", session.Id, session.LastActive)
			session.Mutex().RUnlock()
			if !expired {
				continue
			}

			s.expireSession(session.Id)
		}
	}
}

func (s *SessionService) expireSession(id SessionId) {
	s.cancelMu.Lock()
	s.infoLog.Printf("[%v] expiring session due to inactivity\n", id)
	if cancel, ok := s.cancelExpiration[id]; ok {
		cancel()
		delete(s.cancelExpiration, id)
	}
	s.cancelMu.Unlock()

	s.hub.CloseSession(id)

	if err := s.store.Delete(id); err != nil {
		s.errorLog.Printf("failed to delete expired session %v: %s", id, err)
	}
}

// expireUsers is a domain rule ("inactive users are marked as gone"); it mutates
// session.Users under that session's own lock - the same lock SessionService uses for
// the same field.
func (s *SessionService) expireUsers() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for _, session := range s.store.List() {
			session.Mutex().Lock()

			for _, user := range session.Users {
				if time.Since(user.LastActive).Seconds() > UserExpirationSec && user.Active {
					s.infoLog.Printf("[%v] expire user: %+v\n", session.Id, user)
					user.Active = false

					delete(session.Users, user.Id)
					//TODO check whether session votes must be shown/all active users voted
				}
			}

			session.Mutex().Unlock()
		}
	}
}

func (s *SessionService) Get(id SessionId) (*Session, error) {
	session, err := s.store.Get(id)
	if err != nil {
		return nil, err
	}

	session.Mutex().Lock()
	s.infoLog.Printf("[%v] session requested; refreshing last active\n", id)
	session.LastActive = time.Now()
	session.Mutex().Unlock()

	return session, nil
}

func (s *SessionService) SaveConnectionForUser(sessionId SessionId, userId UserId, conn *websocket.Conn) error {
	session, err := s.Get(sessionId)
	if err != nil {
		return err
	}

	session.Mutex().RLock()
	_, ok := session.Users[userId]
	session.Mutex().RUnlock()
	if !ok {
		return ErrNoRecord
	}

	s.hub.AddConnection(sessionId, userId, conn)

	//naive reader from connection until error happens, otherwise pong handler won't work
	go s.hub.ReadLoop(conn)

	return s.SendUpdates(sessionId)
}

// GetMaskedSessionForUser assumes the caller already holds session's lock (Lock or RLock).
func (s *SessionService) GetMaskedSessionForUser(session Session, userId UserId) Session {
	var votes []Vote

	for displayUserId, user := range session.Users {
		if !user.Active {
			continue
		}

		userVote := session.Votes[displayUserId]
		isCurrentUser := displayUserId == userId

		vote := Vote{
			Name:        html.EscapeString(user.Name),
			Voted:       userVote != "",
			Vote:        getVoteToShow(userVote, session.VotesHidden, isCurrentUser),
			CurrentUser: isCurrentUser,
		}

		votes = append(votes, vote)
	}

	sort.Sort(VotesByName(votes))
	session.VotesInfo = votes
	return session
}

func (s *SessionService) Show(sessionId SessionId, userId UserId) error {
	session, err := s.store.Get(sessionId)
	if err != nil {
		return err
	}

	session.Mutex().Lock()
	_, ok := session.Users[userId]
	if !ok {
		session.Mutex().Unlock()
		return ErrNoRecord
	}

	session.VotesHidden = false
	session.Mutex().Unlock()

	return s.SendUpdates(sessionId)
}

func (s *SessionService) SendUpdates(sessionId SessionId) error {
	s.infoLog.Printf("[%v] send updates for session", sessionId)

	session, err := s.store.Get(sessionId)
	if err != nil {
		s.errorLog.Println(err)
		return err
	}

	// maskedSessionFor/onSent each take and release the session's own lock for a single field
	// access; Hub.Broadcast calls them around network I/O (wsjson.Write) done with no
	// lock held at all, so a slow or stuck client can't block every other request for
	// this session.
	maskedSessionFor := func(userId UserId) Session {
		session.Mutex().RLock()
		defer session.Mutex().RUnlock()

		return s.GetMaskedSessionForUser(*session, userId)
	}

	onSent := func(userId UserId) {
		session.Mutex().Lock()
		defer session.Mutex().Unlock()

		if user, ok := session.Users[userId]; ok {
			s.UpdateUserActiveness(sessionId, user)
		}
	}

	s.hub.Broadcast(sessionId, maskedSessionFor, onSent)

	return nil
}

func (s *SessionService) UpdateUserActiveness(sessionId SessionId, user *User) {
	//s.infoLog.Printf("[%v][%v] update user activeness", sessionId, user.Name)
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
