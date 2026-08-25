package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/romanthekat/planning-poker/internal/poker"
	"github.com/romanthekat/planning-poker/internal/storage/memory"
	httptransport "github.com/romanthekat/planning-poker/internal/transport/http"
)

func main() {
	addr := flag.String("addr", ":10080", "HTTP network address")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	store := memory.NewStore()
	hub := poker.NewHub(errorLog, infoLog)
	sessionService := poker.NewSessionService(store, hub, errorLog, infoLog)

	app := httptransport.NewApplication(sessionService, errorLog, infoLog)

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.Routes(),
	}

	infoLog.Printf("Starting HTTP server on %s", *addr)
	err := srv.ListenAndServe()
	if err != nil {
		errorLog.Fatal(err)
	}
}
