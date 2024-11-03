package main

import "net/http"

func main() {
	// create new serve mux
	mux := http.NewServeMux()

	// register handler
	mux.HandleFunc("/healthz", handlerHealthz)

	// create new http.Server struct
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// use serve mux method to add fileserver handler for rootpath "/app/"
	// strip prefix from the request path before passing it to the fileserver handler
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))

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
