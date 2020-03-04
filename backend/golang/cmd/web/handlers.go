package main

import (
	"encoding/json"
	"net/http"
	"rgm-planning-poker/pkg/models"
	"strconv"
	"strings"
)

func (app *Application) createSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	sessionId, err := app.sessions.Create()
	if err != nil {
		app.serverError(w, err)
	}

	app.infoLog.Printf("new session created: %+v \n", sessionId)
	err = json.NewEncoder(w).Encode(sessionId)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *Application) joinSession(w http.ResponseWriter, r *http.Request) {
	var user *models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	sessionId, err := getSessionId(r)
	if err != nil {
		app.clientError(w, 400)
	}

	session, err := app.sessions.Get(sessionId)
	if err != nil {
		app.clientError(w, 404)
	}

	user = app.sessionService.JoinSession(session, user)

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func getSessionId(r *http.Request) (models.SessionId, error) {
	sessionIdStr := strings.TrimPrefix(r.URL.Path, "/api/")
	sessionId, err := strconv.Atoi(sessionIdStr)
	if err != nil {
		return -1, err
	}

	return models.SessionId(sessionId), nil
}
