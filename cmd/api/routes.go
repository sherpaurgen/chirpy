package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	fsdatabase "github.com/sherpaurgen/boot/internal"
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

	metricsrouter.Get("/metrics", metricsHandler)

	apirouter.Get("/healthz", app.healthcheckHandler)
	apirouter.Get("/reset", resetHandler)
	apirouter.Post("/validate_chirp", validatechirpHandler)
	apirouter.Post("/chirps", saveChirpHandler)
	apirouter.Get("/chirps", getChirpHandler)

	mainrouter.Mount("/api", apirouter)
	mainrouter.Mount("/admin", metricsrouter)

	return mainrouter
}
func getChirpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fpath := "./data.json"
	log.Printf("getchiphandler called...%v\n", fpath)
	jsondata, _ := fsdatabase.ReadData(fpath)
	log.Print(string(jsondata))
	w.WriteHeader(http.StatusOK)
	w.Write(jsondata)
}
func saveChirpHandler(w http.ResponseWriter, r *http.Request) {
	fpath := "./data.json"
	w.Header().Set("Content-Type", "application/json")
	wordsToReplace := []string{"kerfuffle", "sharbert", "fornax"}
	type parameters struct {
		Body string `json:"body"`
	}
	type errorMsg struct {
		Error string `json:"error"`
	}
	type validBody struct {
		Id   int    `json:"id"`
		Body string `json:"body"`
	}

	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	fmt.Println(params.Body)
	if params.Body == "" || err != nil {
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

	//lowercasemsg := strings.ToLower(params.Body)
	cleanword := replaceWords(params.Body, wordsToReplace)
	//now that the tweet is valid
	sanitizedData := fsdatabase.Chirp{
		Body: cleanword,
		Id:   0,
	}

	jsondata, err := fsdatabase.WriteData(fpath, sanitizedData)
	if err != nil || jsondata == nil {
		respBody := errorMsg{
			Error: "Problem in encoding json fsdatabase",
		}
		w.WriteHeader(400)
		dat, _ := json.Marshal(respBody)
		w.Write(dat)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(jsondata)
}

func replaceWords(input string, wordsToReplace []string) string {

	pattern := fmt.Sprintf(`(?i)\b(%s)\b`, strings.Join(wordsToReplace, "|"))

	// Compile the regular expression
	regex := regexp.MustCompile(pattern)

	// Replace matched words with "****"
	replaced := regex.ReplaceAllStringFunc(input, func(match string) string {
		return "****"
	})

	return replaced
}

func validatechirpHandler(w http.ResponseWriter, r *http.Request) {
	wordsToReplace := []string{"kerfuffle", "sharbert", "fornax"}
	type parameters struct {
		Body string `json:"body"`
	}
	type errorMsg struct {
		Error string `json:"error"`
	}
	type validBody struct {
		Cleaned_body string `json:"cleaned_body"`
	}
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	fmt.Println(params.Body)
	if params.Body == "" || err != nil {
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

	//lowercasemsg := strings.ToLower(params.Body)
	cleanword := replaceWords(params.Body, wordsToReplace)
	//now that the tweet is valid
	respBody := validBody{
		Cleaned_body: cleanword,
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
