package main

import (
	"fmt"
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

	mux.Handle("/app/", middlewareMetricsInc(cfg, http.StripPrefix("/app", http.FileServer(http.Dir(FILEPATHROOT)))))
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

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fileserverHits := cfg.fileserverHits.Load()
	fmt.Fprintf(w, "Hits: %d\n", fileserverHits)
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	w.Write([]byte("Hits counter reset to 0"))
}

func middlewareMetricsInc(cfg *apiConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
