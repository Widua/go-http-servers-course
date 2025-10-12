package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (cfg *apiConfig) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	serveMux := http.NewServeMux()
	server := http.Server{
		Handler: serveMux,
		Addr:    ":8080",
	}
	config := apiConfig{fileServerHits: atomic.Int32{}}
	serveMux.Handle("/app/", config.metricsMiddleware(handleFileserver()))
	serveMux.HandleFunc("POST /reset", config.handleReset)
	serveMux.HandleFunc("GET /healthz", handleHealthz)
	serveMux.HandleFunc("GET /metrics", config.handleMetrics)
	server.ListenAndServe()
}

func handleFileserver() http.Handler {
	return http.StripPrefix("/app", http.FileServer(http.Dir(".")))
}
func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileServerHits.Load())))
}
func (cfg *apiConfig) handleReset(responseWriter http.ResponseWriter, req *http.Request) {
	cfg.fileServerHits.Store(0)
	responseWriter.WriteHeader(200)
	responseWriter.Header().Add("Content-Type", "text/plain")
	responseWriter.Write([]byte("OK"))
}

func handleHealthz(responseWriter http.ResponseWriter, req *http.Request) {
	responseWriter.WriteHeader(200)
	responseWriter.Header().Add("Content-Type", "text/plain")
	responseWriter.Write([]byte("OK"))
}
