package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/romanthekat/planning-poker/internal/poker"
)

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

func (app *Application) getSessionById(w http.ResponseWriter, r *http.Request) {
	sessionId, err := getSessionIdFromPath(r)
	if err != nil {
		app.badRequest(w)
		return
	}

	session, err := app.sessionService.Get(sessionId)
	if err != nil {
		if errors.Is(err, poker.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	userId, err := getUserIdFromParam(r)
	if errors.Is(err, poker.ErrNoUserId) {
		// in this specific case user id is optional
		app.noContent(w)
		return
	}
	if err != nil {
		app.badRequest(w)
		return
	}

	maskedSession := app.sessionService.GetMaskedSessionForUser(*session, userId)
	err = json.NewEncoder(w).Encode(maskedSession)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *Application) joinSession(w http.ResponseWriter, r *http.Request) {
	var createUserRequest *poker.CreateUserRequest
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

	user := &poker.User{Name: createUserRequest.Name}
	app.infoLog.Printf("[%v] join request for user '%v'", sessionId, user.Name)
	user, err = app.sessionService.JoinSession(sessionId, user)
	if err != nil {
		if errors.Is(err, poker.ErrNoRecord) {
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
	var vote *poker.VoteRequest
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
	var userRequest *poker.UserRequest
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

	var userRequest *poker.UserRequest
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

func getSessionIdFromPath(r *http.Request) (poker.SessionId, error) {
	sessionIdStr := r.PathValue("sessionId")
	sessionId, err := strconv.Atoi(sessionIdStr)
	if err != nil {
		return -1, err
	}

	return poker.SessionId(sessionId), nil
}

func getUserIdFromParam(r *http.Request) (poker.UserId, error) {
	userIdStr := r.URL.Query().Get("userId")
	if userIdStr == "" {
		return -1, poker.ErrNoUserId
	}

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return -1, err
	}

	return poker.UserId(userId), nil
}

func getUserIdFromPath(r *http.Request) (poker.UserId, error) {
	userIdStr := r.PathValue("userId")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return -1, err
	}

	return poker.UserId(userId), nil
}
