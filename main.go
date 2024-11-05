package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

// struct for keeping track server hits, atomic.Int32 has various methods to change the value
type apiConfig struct {
	fileserverHits atomic.Int32
}

type requestParameters struct {
	Body string `json:"body"`
}

type responseParameters struct {
	Error        string `json:"error"`
	Valid        bool   `json:"valid"`
	Cleaned_body string `json:"cleaned_body"`
}

func main() {
	// initialize apiCfg
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}

	// create new serve mux
	mux := http.NewServeMux()

	// register handlers
	mux.HandleFunc("GET /api/healthz", handlerHealthz)
	// mux.HandleFunc("GET /api/metrics", apiCfg.serverHitsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHitHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.adminMetrics)
	mux.HandleFunc("POST /api/validate_chirp", validateChirp)

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

// custom handler function
func handlerHealthz(w http.ResponseWriter, r *http.Request) {
	//write Content-Type: text/plain; charset=utf-8
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// write status code
	w.WriteHeader(http.StatusOK)

	// write body
	w.Write([]byte("OK"))
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	// decode the JSON body
	decoder := json.NewDecoder(r.Body)
	params := requestParameters{}
	err := decoder.Decode(&params)
	if err != nil {
		http.Error(w, "Invalid Json", http.StatusBadRequest)
	}

	// check length of json body, cannot exceed 140 chars
	if len(params.Body) <= 140 && len(params.Body) > 0 {

		// response for accepted body
		response := responseParameters{
			Valid:        true,
			Cleaned_body: replaceProfanity(params.Body),
		}
		statusCode := 200
		// encode response
		encodeResponse(w, response, statusCode)
	}

	if len(params.Body) == 0 {
		response := responseParameters{
			Error: "Chirp can't be 0 characters",
			Valid: false,
		}
		statusCode := 400
		encodeResponse(w, response, statusCode)
	}

	if len(params.Body) > 140 {
		response := responseParameters{
			Error: "Chirp is too long",
			Valid: false,
		}
		statusCode := 400
		encodeResponse(w, response, statusCode)
	}
}

// helper function to clean profanity
func replaceProfanity(p string) string {

	// the no-no words
	profanity := []string{"kerfuffle", "sharbert", "fornax"}
	// case insensitive
	str := strings.ToLower(p)

	for _, word := range profanity {
		if strings.Contains(str, word) {
			// replace the profanity with "****"
			str = strings.ReplaceAll(str, word, "****")
		}
	}

	return str
}

// helper function to reduce copying code
func encodeResponse(w http.ResponseWriter, response responseParameters, statusCode int) {
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

// method on apiConfig struct handler
// func (cfg *apiConfig) serverHitsHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
// 	w.WriteHeader(http.StatusOK)

// 	// write the amount of server hits
// 	hitNumber := fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())
// 	w.Write([]byte(hitNumber))
// }

func (cfg *apiConfig) adminMetrics(w http.ResponseWriter, r *http.Request) {
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

// reset method handler that sets hitnumber to 0
func (cfg *apiConfig) resetHitHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	w.Write([]byte("Hits reset to 0"))
}

// middleware method that increments the fileserverHits counter every time it's called
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
