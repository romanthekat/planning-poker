package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/romanthekat/planning-poker/pkg/models"
	"github.com/romanthekat/planning-poker/pkg/models/memory"
	"github.com/romanthekat/planning-poker/pkg/services"
)

type Application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	sessions       models.SessionModel
	sessionService *services.SessionService
}

func main() {
	addr := flag.String("addr", ":10080", "HTTP network address")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	sessionModel := memory.NewSessionModel()
	sessionService := services.NewSessionService(sessionModel, errorLog, infoLog)

	app := &Application{
		errorLog:       errorLog,
		infoLog:        infoLog,
		sessions:       sessionModel,
		sessionService: sessionService,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting HTTP server on %s", *addr)
	err := srv.ListenAndServe()
	if err != nil {
		errorLog.Fatal(err)
	}
}
