package poker

import (
	"context"
	"io"
	"log"
	"maps"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const pongWait = 5 * time.Second
const pingPeriod = (pongWait * 9) / 10

// Hub owns every live websocket connection and the periodic ping/broadcast logic for
// every Session
// TODO consider if it should websocket specific package, not poker/logic
type Hub struct {
	mu          sync.RWMutex
	connections map[SessionId]map[UserId]*websocket.Conn

	errorLog *log.Logger
	infoLog  *log.Logger
}

func NewHub(errorLog, infoLog *log.Logger) *Hub {
	return &Hub{
		connections: make(map[SessionId]map[UserId]*websocket.Conn),
		errorLog:    errorLog,
		infoLog:     infoLog,
	}
}

// AddConnection registers conn as userId's live connection for sessionId, closing any
// previous connection already held for that user.
func (h *Hub) AddConnection(sessionId SessionId, userId UserId, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns, ok := h.connections[sessionId]
	if !ok {
		conns = make(map[UserId]*websocket.Conn)
		h.connections[sessionId] = conns
	}

	if existing, ok := conns[userId]; ok {
		_ = existing.Close(websocket.StatusNormalClosure, "")
	}
	conns[userId] = conn
}

func (h *Hub) removeConnection(sessionId SessionId, userId UserId) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.connections[sessionId], userId)
}

// TODO overkill?
// connectionsSnapshot returns a copy of sessionId's live connections, so callers can do
// network I/O (writes, pings) without holding the hub's lock for the duration.
func (h *Hub) connectionsSnapshot(sessionId SessionId) map[UserId]*websocket.Conn {
	h.mu.RLock()
	defer h.mu.RUnlock()

	snapshot := make(map[UserId]*websocket.Conn, len(h.connections[sessionId]))
	maps.Copy(snapshot, h.connections[sessionId])
	return snapshot
}

func (h *Hub) Broadcast(sessionId SessionId, maskedSessionFor func(userId UserId) Session, onSent func(userId UserId)) {
	for userId, conn := range h.connectionsSnapshot(sessionId) {
		err := wsjson.Write(context.Background(), conn, maskedSessionFor(userId))
		if err != nil {
			h.errorLog.Printf("[%v] user %v: %s\n", sessionId, userId, err)
			continue
		}

		if onSent != nil {
			onSent(userId)
		}
	}
}

// PingLoop periodically pings sessionId's connections until ctx is canceled (normally
// when the session expires, see SessionService)
func (h *Hub) PingLoop(ctx context.Context, sessionId SessionId) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.Ping(sessionId)
		}
	}
}

func (h *Hub) Ping(sessionId SessionId) {
	for userId, conn := range h.connectionsSnapshot(sessionId) {
		if err := conn.Ping(context.Background()); err != nil {
			h.errorLog.Printf("ping error: %s", err)
			h.removeConnection(sessionId, userId)
		}
	}
}

// CloseSession closes every live connection for sessionId and forgets about them
func (h *Hub) CloseSession(sessionId SessionId) {
	h.mu.Lock()
	conns := h.connections[sessionId]
	delete(h.connections, sessionId)
	h.mu.Unlock()

	for _, conn := range conns {
		_ = conn.Close(websocket.StatusGoingAway, "session has expired")
	}
}

// ReadLoop naively drains messages from conn until it errors; this is required for the
// underlying websocket library's pong handler to keep working, even though the messages
// themselves aren't currently used for anything.
// TODO check if it's still relevant given change in websocket lib
func (h *Hub) ReadLoop(c *websocket.Conn) {
	for {
		messageType, reader, err := c.Reader(context.Background())
		h.infoLog.Println("websocket messageType: ", messageType)
		if err != nil {
			h.errorLog.Println(err)
			_ = c.Close(websocket.StatusInternalError, "internal error")
			return
		}

		buf := new(strings.Builder)
		_, err = io.Copy(buf, reader)
		if err != nil {
			h.errorLog.Println(err)
			_ = c.Close(websocket.StatusInternalError, "internal error")
			return
		}

		h.infoLog.Println("websocket message: ", buf.String())
	}
}
