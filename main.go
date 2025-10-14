package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

type apiError struct {
	Error string `json:"error"`
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
	serveMux.HandleFunc("POST /admin/reset", config.handleReset)
	serveMux.HandleFunc("GET /api/healthz", handleHealthz)
	serveMux.HandleFunc("GET /admin/metrics", config.handleMetrics)
	serveMux.HandleFunc("POST /api/validate_chirp", validateChirp)
	server.ListenAndServe()
}

func handleFileserver() http.Handler {
	return http.StripPrefix("/app", http.FileServer(http.Dir(".")))
}
func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, req *http.Request) {
	metricsTemplate := ` 
	<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>
	`
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte(fmt.Sprintf(metricsTemplate, cfg.fileServerHits.Load())))
}
func (cfg *apiConfig) handleReset(responseWriter http.ResponseWriter, req *http.Request) {
	cfg.fileServerHits.Store(0)
	responseWriter.WriteHeader(200)
	responseWriter.Header().Add("Content-Type", "text/plain")
	responseWriter.Write([]byte("OK"))
}

func validateChirp(res http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type validConfirmation struct {
		Valid bool `json:"valid"`
	}
	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		res.WriteHeader(400)
		errorMess := apiError{Error: err.Error()}
		body, _ := json.Marshal(errorMess)
		res.Write(body)
		return
	}
	body := params.Body

	if len(body) > 140 {
		res.WriteHeader(400)
		errorMess := apiError{Error: "Chirp is too long"}
		body, _ := json.Marshal(errorMess)
		res.Write(body)
		return
	}

	resp := validConfirmation{Valid: true}
	res.WriteHeader(200)
	respBody, _ := json.Marshal(resp)
	res.Write(respBody)
}

func handleHealthz(responseWriter http.ResponseWriter, req *http.Request) {
	responseWriter.WriteHeader(200)
	responseWriter.Header().Add("Content-Type", "text/plain")
	responseWriter.Write([]byte("OK"))
}
