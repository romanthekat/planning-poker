package main

import (
	"net/http"
)

func (app *Application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/createSession", app.createSession)
	mux.HandleFunc("/api/joinSession", app.joinSession)

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	//TODO use alice middleware?
	return app.logRequest(app.authorization(mux))
}
