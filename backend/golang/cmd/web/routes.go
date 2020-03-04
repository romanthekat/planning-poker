package main

import (
	"net/http"
)

func (app *Application) routes() http.Handler {
	mux := http.NewServeMux()

	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/createSession", app.createSession)
	apiMux.HandleFunc("/", app.joinSession)

	mux.Handle("/api/", http.StripPrefix("/api", app.postRequest(apiMux)))

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	//TODO use alice middleware?
	return app.logRequest(app.authorization(mux))
}
