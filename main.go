package main

import (
	"log"
	"net/http"
)

func endpointHandler(w http.ResponseWriter,r *http.Request)  {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))	
	}

func main() {
	const filepathRoot = "."
	const port = "8080"

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app",http.FileServer(http.Dir(filepathRoot))))
	mux.HandleFunc("/healthz", endpointHandler)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files form %s on port %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())

}
