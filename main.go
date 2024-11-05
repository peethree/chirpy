package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

// struct for keeping track server hits, atomic.Int32 has various methods to change the value
type apiConfig struct {
	fileserverHits atomic.Int32
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
