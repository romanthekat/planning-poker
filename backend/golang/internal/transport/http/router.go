package http

import (
	"net/http"

	"github.com/rs/cors"
)

func (app *Application) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/sessions", app.createSession)

	mux.HandleFunc("GET /api/sessions/{sessionId}", app.checkSessionExists)
	mux.HandleFunc("POST /api/sessions/{sessionId}/join", app.joinSession)
	//TODO mux can't separate /number vs /text
	mux.HandleFunc("GET /api/sessions/{sessionId}/get/{userId}", app.getWebsocketConnection)
	mux.HandleFunc("POST /api/sessions/{sessionId}/vote", app.vote)
	mux.HandleFunc("POST /api/sessions/{sessionId}/clear", app.clear)
	mux.HandleFunc("POST /api/sessions/{sessionId}/show", app.show)

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"X-Requested-With", "Content-Type", "Authorization"},
		//AllowCredentials: true,
	})

	return app.logRequest(c.Handler(mux))
}
