// Package http contains everything that turns SessionService's business logic into an
// HTTP + websocket API: routing, request decoding/validation, and response helpers. It
// is the only package in this module that depends on net/http, so it is also the only
// place a second transport (e.g. gRPC) would need to be added alongside.
package http

import (
	"log"

	"github.com/romanthekat/planning-poker/internal/poker"
)

// Application wires together everything an HTTP handler needs; it replaces the
// package-main Application struct the handlers used to be methods on.
type Application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	sessionService *poker.SessionService
}

func NewApplication(sessionService *poker.SessionService, errorLog *log.Logger, infoLog *log.Logger) *Application {
	return &Application{
		errorLog:       errorLog,
		infoLog:        infoLog,
		sessionService: sessionService,
	}
}
