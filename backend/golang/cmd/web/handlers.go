package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/coder/websocket"
	"github.com/go-playground/validator/v10"
	"github.com/romanthekat/planning-poker/pkg/models"
)

var validate = validator.New()

func (app *Application) createSession(w http.ResponseWriter, r *http.Request) {
	app.infoLog.Println("creating new session")
	session, err := app.sessionService.Create()
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.infoLog.Printf("new session created: %+v \n", session)
	err = json.NewEncoder(w).Encode(session)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

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
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.badRequest(w)
		}
	}
}

func (app *Application) pingPongReceiver(sessionId models.SessionId, userId models.UserId) func(context.Context, []byte) {
	return func(ctx context.Context, payload []byte) {
		session, err := app.sessionService.Get(sessionId)
		if err != nil {
			return
		}

		session.Mutex().Lock()
		user, ok := session.Users[userId]
		if ok {
			app.sessionService.UpdateUserActiveness(user)
		}
		session.Mutex().Unlock()
	}
}

func (app *Application) checkSessionExists(w http.ResponseWriter, r *http.Request) {
	sessionId, err := getSessionIdFromPath(r)
	if err != nil {
		app.badRequest(w)
		return
	}

	_, err = app.sessionService.Get(sessionId)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.noContent(w)
}

func (app *Application) joinSession(w http.ResponseWriter, r *http.Request) {
	var createUserRequest *models.CreateUserRequest
	err := json.NewDecoder(r.Body).Decode(&createUserRequest)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w)
		return
	}
	if err := validate.Struct(createUserRequest); err != nil {
		app.clientErrorWithText(w, http.StatusBadRequest, err)
		return
	}

	sessionId, err := getSessionIdFromPath(r)
	if err != nil {
		app.badRequest(w)
		return
	}

	user := &models.User{Name: createUserRequest.Name}
	app.infoLog.Printf("[%v] join request for user '%v'", sessionId, user.Name)
	user, err = app.sessionService.JoinSession(sessionId, user)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	app.infoLog.Printf("[%v] joined id:%v, name:%v", sessionId, user.Id, user.Name)

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *Application) vote(w http.ResponseWriter, r *http.Request) {
	var vote *models.VoteRequest
	err := json.NewDecoder(r.Body).Decode(&vote)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w)
		return
	}
	if err := validate.Struct(vote); err != nil {
		app.clientErrorWithText(w, http.StatusBadRequest, err)
		return
	}

	sessionId, err := getSessionIdFromPath(r)
	if err != nil {
		app.badRequest(w)
		return
	}

	app.infoLog.Printf("[%v] vote %+v", sessionId, vote)
	err = app.sessionService.Vote(sessionId, vote)
	if err != nil {
		app.notFound(w)
		return
	}

	err = json.NewEncoder(w).Encode(vote)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *Application) show(w http.ResponseWriter, r *http.Request) {
	var userRequest *models.UserRequest
	err := json.NewDecoder(r.Body).Decode(&userRequest)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w)
		return
	}
	if err := validate.Struct(userRequest); err != nil {
		app.clientErrorWithText(w, http.StatusBadRequest, err)
		return
	}

	sessionId, err := getSessionIdFromPath(r)
	if err != nil {
		app.badRequest(w)
		return
	}

	err = app.sessionService.Show(sessionId, userRequest.UserId)
	if err != nil {
		app.notFound(w)
		return
	}
}

func (app *Application) clear(w http.ResponseWriter, r *http.Request) {
	sessionId, err := getSessionIdFromPath(r)
	if err != nil {
		app.badRequest(w)
		return
	}

	var userRequest *models.UserRequest
	err = json.NewDecoder(r.Body).Decode(&userRequest)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w)
		return
	}
	if err := validate.Struct(userRequest); err != nil {
		app.clientErrorWithText(w, http.StatusBadRequest, err)
		return
	}

	err = app.sessionService.Clear(sessionId, userRequest.UserId)
	if err != nil {
		app.notFound(w)
		return
	}
}

func getSessionIdFromPath(r *http.Request) (models.SessionId, error) {
	sessionIdStr := r.PathValue("sessionId")
	sessionId, err := strconv.Atoi(sessionIdStr)
	if err != nil {
		return -1, err
	}

	return models.SessionId(sessionId), nil
}

func getUserIdFromPath(r *http.Request) (models.UserId, error) {
	userIdStr := r.PathValue("userId")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return -1, err
	}

	return models.UserId(userId), nil
}
