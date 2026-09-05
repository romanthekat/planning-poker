package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/coder/websocket"
	"github.com/romanthekat/planning-poker/internal/poker"
)

func (app *Application) getWebsocketConnection(w http.ResponseWriter, r *http.Request) {
	sessionId, err := getSessionIdFromPath(r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	userId, err := getUserIdFromPath(r)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OnPingReceived: nil,
		OnPongReceived: app.pingPongReceiver(sessionId, userId),
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.sessionService.SaveConnectionForUser(sessionId, userId, conn)
	if err != nil {
		if errors.Is(err, poker.ErrNoRecord) {
			_ = conn.Close(websocket.StatusPolicyViolation, "user not found")
		} else {
			_ = conn.Close(websocket.StatusPolicyViolation, "bad request")
		}
	}
}

func (app *Application) pingPongReceiver(sessionId poker.SessionId, userId poker.UserId) func(context.Context, []byte) {
	return func(ctx context.Context, payload []byte) {
		session, err := app.sessionService.Get(sessionId)
		if err != nil {
			return
		}

		session.Mutex().Lock()
		user, ok := session.Users[userId]
		if ok {
			app.sessionService.UpdateUserActiveness(sessionId, user)
		}
		session.Mutex().Unlock()
	}
}
