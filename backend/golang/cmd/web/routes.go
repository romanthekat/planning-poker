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
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With, Content-Type, Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	return app.logRequest(handlers.CORS(headersOk, originsOk, methodsOk)(topMux))
}
