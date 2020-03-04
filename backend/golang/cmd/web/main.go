package main

import (
	"flag"
	"log"
	"math/rand"
	"net/http"
	"os"
	"rgm-planning-poker/pkg/models"
	"rgm-planning-poker/pkg/models/memory"
	"rgm-planning-poker/pkg/services"
	"time"
)

type Application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	sessions       models.SessionModel
	sessionService *services.SessionService
}

func main() {
	rand.Seed(time.Now().UnixNano())

	addr := flag.String("addr", ":10080", "HTTP network address")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	sessionModel := memory.NewSessionModel()
	sessionService := services.NewSessionService()

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
