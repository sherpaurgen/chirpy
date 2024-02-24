package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
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
	apirouter.Get("/chirps", getAllChirpHandler)
	apirouter.Get("/chirps/{id}", getChirp)
	apirouter.Post("/validate_chirp", validatechirpHandler)
	apirouter.Post("/chirps", saveChirpHandler)
	apirouter.Post("/users", CreateUserHandler)
	apirouter.Put("/users", changeAccount)
	apirouter.Post("/login", loginHandler)

	mainrouter.Mount("/api", apirouter)
	mainrouter.Mount("/admin", metricsrouter)

	return mainrouter
}

func changeAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	//Extract JWT token from the Authorization header

	authHeader := r.Header.Get("Authorization")
	tokenString := strings.Split(authHeader, "Bearer ")[1]
	token, err := verifyJwt(tokenString)
	if err != nil {
		log.Println("error occuredin parsewithclaims err:", err)
		w.WriteHeader(401)
		w.Write([]byte("check bearer token for errors"))
		return
	}

	userIDString, err := token.Claims.GetSubject() //get userid from jwt
	handleErrorPrint("changeAccount", err)
	var userinfo fsdatabase.User
	err = json.NewDecoder(r.Body).Decode(&userinfo)
	handleErrorPrint("changeAccount", err)

	log.Println(userIDString)
	fpath := "./data.json"
	res, err := fsdatabase.ModifyUser(fpath, userIDString, userinfo)
	handleErrorPrint("changeAccount", err)
	w.WriteHeader(200)
	w.Write(res)
}

func verifyJwt(tokenString string) (*jwt.Token, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Cannot load env in changeAccount")
	}
	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		},
	)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fpath := "./data.json"
	var user fsdatabase.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, fmt.Sprintf("check json malformed body: %v", err),
			http.StatusBadRequest)
		return
	}
	jsondata, err := fsdatabase.AuthenticateUser(user, fpath)
	if err != nil {
		log.Println("error returned by fsdatabase.AuthenticateUser(user, fpath)", err)
		w.WriteHeader(401)
		return
	}
	//now that user login is successful we generate jwt
	// username := user.Email
	err = json.Unmarshal(jsondata, &user)
	log.Println("this is user: ", user)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte("error: unmarshaling jsondata"))
		return
	}
	userid := user.Id
	// useremail := user_v.Email
	expires_in_seconds := user.Expires_in_seconds

	tokenString, err := generateJWT(userid, expires_in_seconds)
	if err != nil {
		w.WriteHeader(401)
		log.Println("error returned by generateJWT(userid, expires_in_seconds)")
		return
	}
	//for successful auth check respond 200 and email+id
	w.Header().Set("Authorization", "Bearer "+tokenString)
	w.WriteHeader(200)
	w.Write(jsondata)
}

func generateJWT(userid int, expires_in_seconds int) (string, error) {
	//sampleSecretKey := []byte("JtRp8DrxkynDo7mfRqMDaSfntlDqleoKfaMkcp0Fh33aCMR0mA8pOGcqsexlEEC8BTfDX2U1dVjkIbc1qnkr4g")
	log.Println(expires_in_seconds)
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	sampleSecretKey := []byte(os.Getenv("JWT_SECRET_KEY"))
	ExpiresAt := jwt.NewNumericDate(time.Now().Add(time.Duration(expires_in_seconds) * time.Second))
	log.Println("Expiresat:", ExpiresAt)
	if expires_in_seconds < 1 {
		ExpiresAt = jwt.NewNumericDate(time.Now().Add(24 * time.Hour))
	}
	log.Println("Expiresat:", ExpiresAt)
	claims := jwt.RegisteredClaims{
		// A usual scenario is to set the expiration time relative to the current time
		ExpiresAt: ExpiresAt,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "chirpy",
		Subject:   fmt.Sprint(userid),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(sampleSecretKey)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return tokenString, nil
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	fpath := "./data.json"
	var user fsdatabase.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Error decoding request body",
			http.StatusInternalServerError)
		return
	}
	// User struct with email,pass is sent NOT json
	jsondata, err := fsdatabase.CreateUser(user, fpath)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsondata)
		return
	}
	w.WriteHeader(201)
	w.Write(jsondata)
}

func getChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fpath := "./data.json"
	log.Printf("getchiphandler called...%v\n", fpath)
	id := chi.URLParam(r, "id")
	jsondata, err := fsdatabase.ReadChirpData(fpath, id)
	if jsondata == nil || err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Println(err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsondata)
}

func getAllChirpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fpath := "./data.json"
	log.Printf("getchiphandler called...%v\n", fpath)
	id := "" //sending empty id string for getall chirp
	jsondata, _ := fsdatabase.ReadChirpData(fpath, id)
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

	jsondata, err := fsdatabase.WriteChirpData(fpath, sanitizedData)
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

func handleErrorFatal(funcname string, err error, data ...string) {
	if err != nil {
		log.Fatalf("error in function %v: %v\n", funcname, err)
	}
}

func handleErrorPrint(funcname string, err error, data ...string) {
	if err != nil {
		log.Printf("error in function %v: %v\n", funcname, err)
		return
	}

}
