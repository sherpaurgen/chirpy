package main

import (
	"encoding/json"
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
	apirouter.Post("/validate_chirp", validatechirpHandler)
	mainrouter.Mount("/api", apirouter)
	mainrouter.Mount("/admin", metricsrouter)

	return mainrouter
}
func validatechirpHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type errorMsg struct {
		Error string `json:"error"`
	}
	type validBody struct {
		Valid bool `json:"valid"`
	}
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respBody := errorMsg{
			Error: "Something went wrong",
		}
		w.WriteHeader(400)
		dat, _ := json.Marshal(respBody)
		w.Write(dat)
		return
	}
	chirpLen := len(params.Body)
	if chirpLen > 140 {
		respBody := errorMsg{
			Error: "Chirp is too long",
		}
		w.WriteHeader(400)
		dat, _ := json.Marshal(respBody)
		w.Write(dat)
		return
	}
	//now that the tweet is valid
	respBody := validBody{
		Valid: true,
	}
	w.WriteHeader(http.StatusOK)
	dat, _ := json.Marshal(respBody)
	w.Write(dat)
	return
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
