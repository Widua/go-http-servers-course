package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/widua/go-http-server/api"
	"github.com/widua/go-http-server/internal/database"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

type databaseConfig struct {
	db_connection *sql.DB
	queries       *database.Queries
}

var profaneWords []string = []string{"kerfuffle", "sharbert", "fornax"}

func (cfg *apiConfig) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	godotenv.Load(".env")
	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbUrl)
	db_config := databaseConfig{db_connection: db, queries: database.New(db)}
	fmt.Printf("Succesfully connected to database: %v", db_config.queries)
	if err != nil {
		panic("Error while connecting to database")
	}
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
	api.RespondOk(responseWriter)
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
		api.RespondWithError(res, 400, err.Error())
		return
	}
	body := params.Body

	if len(body) > 140 {
		api.RespondWithError(res, 400, "Chirp is too long")
		return
	}
	chirpClearer(res, body)
}

func chirpClearer(out http.ResponseWriter, body string) {
	type cleanedBody struct {
		CleanedBody string `json:"cleaned_body"`
	}
	splittedBody := strings.Split(body, " ")

	for ix, word := range splittedBody {
		if slices.Contains(profaneWords, strings.ToLower(word)) {
			splittedBody[ix] = "****"
		}
	}
	censoredBody := cleanedBody{CleanedBody: strings.Join(splittedBody, " ")}

	byteBody, err := json.Marshal(censoredBody)

	if err != nil {
		api.RespondWithError(out, 400, err.Error())
		return
	}

	api.RespondWithJSON(out, 200, byteBody)
}

func handleHealthz(responseWriter http.ResponseWriter, req *http.Request) {
	api.RespondOk(responseWriter)
}
