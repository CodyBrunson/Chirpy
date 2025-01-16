package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	const PORT = "8080"
	const FILEPATHROOT = "."

	cfg := &apiConfig{
		fileserverHits: atomic.Int32{},
	}

	mux := http.NewServeMux()

	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(FILEPATHROOT)))))
	mux.Handle("/metrics/", http.HandlerFunc(cfg.handlerMetrics))
	mux.Handle("/reset/", http.HandlerFunc(cfg.handlerReset))
	mux.HandleFunc("/healthz", handlerReadiness)

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + PORT,
	}

	log.Printf("Serving on port: %s\n", PORT)
	log.Fatal(server.ListenAndServe())
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
