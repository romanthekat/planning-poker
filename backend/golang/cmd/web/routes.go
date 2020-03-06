package main

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"net/http"
)

func (app *Application) routes() http.Handler {
	topMux := mux.NewRouter()

	topMux.HandleFunc("/api/sessions", app.createSession)
	topMux.HandleFunc("/api/sessions/{sessionId}/get/{userId}", app.getSession)
	topMux.HandleFunc("/api/sessions/{sessionId}/join", app.joinSession)
	topMux.HandleFunc("/api/sessions/{sessionId}/vote", app.vote)

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	topMux.Handle("/static/", http.StripPrefix("/static", fileServer))

	//TODO use alice middleware?
	corsObj := handlers.AllowedOrigins([]string{"*"})
	return handlers.CORS(corsObj)(app.logRequest(app.authorization(topMux)))
}
