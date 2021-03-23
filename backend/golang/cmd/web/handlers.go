package main

import (
	"encoding/json"
	"github.com/EvilKhaosKat/planning-poker/pkg/models"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
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

func (app *Application) getSession(w http.ResponseWriter, r *http.Request) {
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
		if err == models.ErrNoRecord {
			app.clientError(w, http.StatusNotFound)
		} else {
			app.clientError(w, http.StatusBadRequest)
		}
	}
}

func (app *Application) joinSession(w http.ResponseWriter, r *http.Request) {
	var user *models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	if err := validator.Validate(user); err != nil {
		app.clientErrorWithText(w, http.StatusBadRequest, err)
		return
	}

	sessionId, err := getSessionId(r)

	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	app.infoLog.Printf("join session %v for user %+v", sessionId, user)
	user, err = app.sessionService.JoinSession(sessionId, user)
	if err != nil {
		if err == models.ErrNoRecord {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

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
		app.clientError(w, http.StatusBadRequest)
		return
	}

	app.infoLog.Printf("vote %+v in session %v", vote, sessionId)
	err = app.sessionService.Vote(sessionId, vote)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
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
		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = app.sessionService.Show(sessionId)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}
}

func (app *Application) clear(w http.ResponseWriter, r *http.Request) {
	sessionId, err := getSessionId(r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = app.sessionService.Clear(sessionId)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
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
