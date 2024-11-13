package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/peethree/chirpy/internal/auth"
	"github.com/peethree/chirpy/internal/database"
)

// struct for keeping track server hits, atomic.Int32 has various methods to change the value
// *database.Queries generated by sqlc
type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	JWTsecret      string
}

type loginParams struct {
	Password           string        `json:"password"`
	Email              string        `json:"email"`
	Expires_in_seconds time.Duration `json:"expires_in_seconds"`
}

// response struct for creating new users/ logging in
type User struct {
	Id         uuid.UUID `json:"id"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
	Email      string    `json:"email"`
	Token      string    `json:"token"`
}

// struct for making new users and getting their email address
type requestUserParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// chirp request parameters
type Chirp struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

// struct for responding to api/chirps
type responseChirp struct {
	Error      string    `json:"error"`
	Valid      bool      `json:"valid"`
	ID         uuid.UUID `json:"id"`
	Body       string    `json:"body"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
	User_id    uuid.UUID `json:"user_id"`
}

func main() {
	// get .env file
	godotenv.Load()

	// get the connection url from .env file
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	// check if .env platform is set to dev
	platformCheck := os.Getenv("PLATFORM")

	jwtSecret := os.Getenv("SECRET")

	// open connection to the db
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	// create new *database.Queries and store it in apiCfg
	dbQueries := database.New(db)

	// initialize apiCfg
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platformCheck,
		JWTsecret:      jwtSecret,
	}

	// create new serve mux
	mux := http.NewServeMux()

	// register handlers
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("GET /api/chirps", apiCfg.loadChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.loadChirpByIDHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.adminMetricsHandler)

	mux.HandleFunc("POST /api/login", apiCfg.loginHandler)
	mux.HandleFunc("POST /api/users", apiCfg.createUserHandler)
	mux.HandleFunc("POST /api/chirps", apiCfg.chirpHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)

	// use serve mux method to register fileserver handler for rootpath "/app/"
	// strip prefix from the request path before passing it to the fileserver handler
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	// create new http.Server struct
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Use the server's ListenAndServe method to start the server
	server.ListenAndServe()
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	// decode JSON body
	decoder := json.NewDecoder(r.Body)
	params := loginParams{}
	err := decoder.Decode(&params)
	if err != nil {
		http.Error(w, "Invalid Json", 400)
		return
	}

	// authenticate user:
	// check to see if email is in the table then compare password
	userExist, err := cfg.db.Login(r.Context(), params.Email)
	if err != nil {
		fmt.Println("%s", err)
		// 401 unauthorized
		http.Error(w, "This email does not match the database", http.StatusUnauthorized)
		return
	}

	// check if the hash matches password if the user's email exists
	if auth.CheckPasswordHash(params.Password, userExist.HashedPassword) != nil {
		fmt.Println("%s", err)
		// 401 unauthorized
		http.Error(w, "Wrong password", http.StatusUnauthorized)
		return
	}

	// token
	// func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	jwt, err := auth.MakeJWT(userExist.ID, cfg.JWTsecret, params.Expires_in_seconds)
	if err != nil {
		fmt.Println("%s", err)
		// 401 unauthorized
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// when the user exists and the password matches the hash -> encode response (login user)
	response := User{
		Id:         userExist.ID,
		Created_at: userExist.CreatedAt,
		Updated_at: userExist.UpdatedAt,
		Email:      userExist.Email,
		Token:      jwt,
	}

	// encode response
	dat, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}

func (cfg *apiConfig) loadChirpByIDHandler(w http.ResponseWriter, r *http.Request) {
	// You can get the string value of the path parameter like in Go with the http.Request.PathValue method.
	pathValue := r.PathValue("chirpID")
	// pathvalue returns a string, LoadChirpByID expects a uuid.UUID type input parameter
	chirpID, err := uuid.Parse(pathValue)
	if err != nil {
		fmt.Println("%s", err)
	}

	// sqlc generated helper function based on query: SELECT * FROM chirps WHERE id = $1;
	chirp, err := cfg.db.LoadChirpByID(r.Context(), chirpID)
	if err != nil {
		fmt.Println("%s", err)
		// 404
		http.Error(w, "Can't find this chirp", 404)
		return
	}

	response := responseChirp{
		ID:         chirp.ID,
		Created_at: chirp.CreatedAt,
		Updated_at: chirp.UpdatedAt,
		Body:       chirp.Body,
		User_id:    chirp.UserID,
	}

	encodeResponse(w, response, 200)
}

// retrieves all chirps in ascending order by created_at (oldest first)
func (cfg *apiConfig) loadChirpsHandler(w http.ResponseWriter, r *http.Request) {
	// sqlc generated function for loading all chirps based on: SELECT * FROM chirps;
	loadedChirps, err := cfg.db.LoadChirps(r.Context())
	if err != nil {
		http.Error(w, "Can't load chirps", 400)
	}

	// declare response variable as a slice of responseChirp structs
	var response []responseChirp

	// loop over all loaded chirps, fill up a responseChirp struct for each chirp, append it to the response slice
	for _, chirp := range loadedChirps {
		individualChirp := responseChirp{
			Body:       chirp.Body,
			Created_at: chirp.CreatedAt,
			Updated_at: chirp.UpdatedAt,
			User_id:    chirp.UserID,
		}
		response = append(response, individualChirp)
	}

	// sort the chirps by created_at
	sort.Slice(response, func(i, j int) bool {
		return response[i].Created_at.Before(response[j].Created_at)
	})

	// marshal the chirps, encoderesponse function does not work with a slice of responseChirp as a parameter
	dat, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}

// create chirp handler
func (cfg *apiConfig) chirpHandler(w http.ResponseWriter, r *http.Request) {
	// decode the JSON body
	decoder := json.NewDecoder(r.Body)
	params := Chirp{}
	err := decoder.Decode(&params)
	if err != nil {
		http.Error(w, "Invalid Json", 400)
		return
	}

	// to create a chirp, a user needs to have a valid jwt
	// get the header for the bearer token
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "Can't get bearer token", http.StatusUnauthorized)
		return
	}

	// check jwt for validity
	_, err = auth.ValidateJWT(bearerToken, cfg.JWTsecret)
	if err != nil {
		http.Error(w, "Invalid JWT", http.StatusUnauthorized)
		return
	}

	// check length of json body, cannot exceed 140 chars
	if len(params.Body) <= 140 {

		// use the helperfunction to clean up profanity
		removed_profanity := replaceProfanity(params.Body)

		// insert the chirp into the db with the sqlc generated createchirp function
		chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
			Body:   removed_profanity,
			UserID: params.UserID,
		})

		if err != nil {
			fmt.Println("%s", err)
			http.Error(w, "Invalid chirp", 400)
			return
		}

		// response for accepted body
		response := responseChirp{
			Valid:      true,
			ID:         chirp.ID,
			Body:       chirp.Body,
			Created_at: chirp.CreatedAt,
			Updated_at: chirp.UpdatedAt,
			User_id:    chirp.UserID,
		}
		statusCode := 201
		// encode response
		encodeResponse(w, response, statusCode)

	} else { // when the body of the request has more than 140 characters
		response := responseChirp{
			Error: "Chirp is too long",
			Valid: false,
		}
		statusCode := 400
		encodeResponse(w, response, statusCode)
	}
}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	// decode JSON for email
	decoder := json.NewDecoder(r.Body)
	params := requestUserParams{}
	err := decoder.Decode(&params)
	if err != nil {
		http.Error(w, "Invalid Json", 400)
		return
	}

	// password length checking
	if len(params.Password) >= 4 {
		// hash the user's password
		hashedPw, err := auth.HashPassword(params.Password)
		if err != nil {
			fmt.Println("%s", err)
		}

		// use the generated CreateUser function to create a user in the database
		new_user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
			Email:          params.Email,
			HashedPassword: hashedPw,
		})
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		// fill up the response fields with the data from the database
		response := User{
			Id:         new_user.ID,
			Created_at: new_user.CreatedAt,
			Updated_at: new_user.UpdatedAt,
			Email:      new_user.Email,
		}

		// encode response
		dat, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		w.Write(dat)
	} else {
		// in case the password length is too short
		http.Error(w, "Password not strong enough", 400)
		return
	}

}

// custom handler function
func healthzHandler(w http.ResponseWriter, r *http.Request) {
	//write Content-Type: text/plain; charset=utf-8
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// write status code
	w.WriteHeader(http.StatusOK)

	// write body
	w.Write([]byte("OK"))
}

// helper function to clean profanity
func replaceProfanity(p string) string {
	// the no-no words
	profanity := []string{"kerfuffle", "sharbert", "fornax"}

	// split up string
	inputStringList := strings.Split(p, " ")

	// check the input string for profanity, replace with **** if it matches profanity
	for _, word := range profanity {
		for i, input := range inputStringList {
			if strings.EqualFold(word, input) {
				inputStringList[i] = "****"
			}
		}
	}

	// add all the seperate (now filtered) words back together in a result string
	result := strings.Join(inputStringList, " ")

	return result
}

// helper function to reduce copying code
func encodeResponse(w http.ResponseWriter, response responseChirp, statusCode int) {
	dat, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(dat)
}

func (cfg *apiConfig) adminMetricsHandler(w http.ResponseWriter, r *http.Request) {
	// set header to html so page knows how to render it
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// template
	template := `
    <html>
      <body>
        <h1>Welcome, Chirpy Admin</h1>
        <p>Chirpy has been visited %d times!</p>
      </body>
    </html>`

	// amount of visits
	hits := cfg.fileserverHits.Load()

	// populate %d of tge template
	html := fmt.Sprintf(template, hits)

	w.WriteHeader(http.StatusOK)

	w.Write([]byte(html))
}

// reset method handler that sets hitnumber to 0 and removes all the users
func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		// forbidden
		w.WriteHeader(403)
		w.Write([]byte("No permission for this endpoint"))
	} else {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		// sets hits to 0
		cfg.fileserverHits.Store(0)
		// delete users
		cfg.db.Reset(r.Context())
		w.Write([]byte("Hits reset to 0, users deleted"))
	}
}

// middleware method that increments the fileserverHits counter every time it's called
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
