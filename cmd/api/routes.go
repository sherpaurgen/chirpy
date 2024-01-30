package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

var (
	visitCount int
	mux        sync.Mutex
)

type application struct {
	config config
	logger *log.Logger
}

func middlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (app *application) Routes() *chi.Mux {
	// Initialize a new chi router instance.
	mainrouter := chi.NewRouter()
	apirouter := chi.NewRouter()
	metricsrouter := chi.NewRouter()

	mainrouter.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	mainrouter.Use(middlewareLogger)
	mainrouter.Use(VisitCounter)

	mainrouter.Handle("/app/assets/", http.StripPrefix("/app/assets", http.FileServer(http.Dir("./assets/"))))
	mainrouter.Handle("/app", http.HandlerFunc(serveIndex))

	apirouter.Get("/healthz", app.healthcheckHandler)
	metricsrouter.Get("/metrics", metricsHandler)
	apirouter.Get("/reset", resetHandler)
	mainrouter.Mount("/api", apirouter)
	mainrouter.Mount("/admin", metricsrouter)

	return mainrouter
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	mux.Lock()
	w.Header().Set("Content-Type", "text/html")
	htmlContent := `
<html>
    <body>
        <h1>Welcome, Chirpy Admin</h1>
        <p>Chirpy has been visited %d times!</p>
    </body>
</html>`

	fmt.Fprintf(w, htmlContent, visitCount)
	mux.Unlock()
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	mux.Lock()
	visitCount = 0
	mux.Unlock()
}

func VisitCounter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/app" || r.URL.Path == "/app/assets/" {
			mux.Lock()
			visitCount++
			mux.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}

func serveIndex(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFile(w, r, "index.html")

}

//mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
