package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

const (
	uploadDir = "uploads"

	maxFiles      = 10
	maxUploadSize = (1 << 20) * maxFiles // n << 20 == n MB
	maxMemory     = 1 << 20
)

func main() {
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		http.ServeFile(w, r, "./ui/static/index.html")
	})
	mux.HandleFunc("POST /file", fileUploadHandler)

	srv := http.Server{
		Addr:         ":3232",
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
