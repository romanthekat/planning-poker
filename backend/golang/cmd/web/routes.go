package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func (app *Application) routes() http.Handler {
	topMux := mux.NewRouter()

	topMux.HandleFunc("/api/createSession", app.createSession)
	topMux.HandleFunc("/api/{sessionId}/get/{userId}", app.getSession)
	topMux.HandleFunc("/api/{sessionId}/join", app.joinSession)
	topMux.HandleFunc("/api/{sessionId}/vote", app.vote)

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	topMux.Handle("/static/", http.StripPrefix("/static", fileServer))

	//TODO use alice middleware?
	return app.logRequest(app.authorization(topMux))
}
