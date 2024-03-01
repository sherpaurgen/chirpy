package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	fsdatabase "github.com/sherpaurgen/boot/internal"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
}

func main() {
	var cfg config
	// default port will be 4000 and env will be "developement" if flag is not provided
	flag.IntVar(&cfg.port, "port", 8080, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	logger := log.New(os.Stdout, "Boot ", log.Ldate|log.Ltime)
	dbfilename := "tokendb.sqlite"
	db, err := fsdatabase.NewDatabase(dbfilename)
	if err != nil {
		log.Fatal("Error initializing database:", err)
	}
	defer db.Close()
	app := &application{
		config: cfg,
		logger: logger,
		db:     db,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.Routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Printf("Starting %s server on %s", cfg.env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)

}
