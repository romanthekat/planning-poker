package main

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/romanthekat/planning-poker/pkg/models"
	"gopkg.in/validator.v2"
	"net/http"
	"strconv"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
	//TODO update origin policy to domain specific, in router too
	return true
}}

func (app *Application) createSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

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
	sessionId, err := getSessionId(r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	userId, err := getUserId(r)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
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

func (app *Application) checkSessionExists(w http.ResponseWriter, r *http.Request) {
	sessionId, err := getSessionId(r)
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
	var user *models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w)
		return
	}
	if err := validator.Validate(user); err != nil {
		app.clientErrorWithText(w, http.StatusBadRequest, err)
		return
	}

	sessionId, err := getSessionId(r)
	if err != nil {
		app.badRequest(w)
		return
	}

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
	var vote *models.Vote
	err := json.NewDecoder(r.Body).Decode(&vote)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	sessionId, err := getSessionId(r)
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
	sessionId, err := getSessionId(r)
	if err != nil {
		app.badRequest(w)
		return
	}

	err = app.sessionService.Show(sessionId)
	if err != nil {
		app.notFound(w)
		return
	}
}

func (app *Application) clear(w http.ResponseWriter, r *http.Request) {
	sessionId, err := getSessionId(r)
	if err != nil {
		app.badRequest(w)
		return
	}

	err = app.sessionService.Clear(sessionId)
	if err != nil {
		app.notFound(w)
		return
	}
}

func getSessionId(r *http.Request) (models.SessionId, error) {
	vars := mux.Vars(r)

	sessionIdStr := vars["sessionId"]
	sessionId, err := strconv.Atoi(sessionIdStr)
	if err != nil {
		return -1, err
	}

	return models.SessionId(sessionId), nil
}

func getUserId(r *http.Request) (models.UserId, error) {
	vars := mux.Vars(r)

	userIdStr := vars["userId"]
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return -1, err
	}

	return models.UserId(userId), nil
}
