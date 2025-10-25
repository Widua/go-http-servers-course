package main

import (
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/widua/go-http-server/internal/api"
	"github.com/widua/go-http-server/internal/database"
)

func main() {
	godotenv.Load(".env")
	dbUrl := os.Getenv("DB_URL")
	database.InitializeDatabase(dbUrl)

	serveMux := http.NewServeMux()
	server := http.Server{
		Handler: serveMux,
		Addr:    ":8080",
	}
	config := api.ApiConfig{FileServerHits: atomic.Int32{}}
	serveMux.Handle("/app/", config.MetricsMiddleware(api.HandleFileserver()))
	serveMux.HandleFunc("POST /admin/reset", config.HandleReset)
	serveMux.HandleFunc("GET /api/healthz", api.HandleHealthz)
	serveMux.HandleFunc("GET /admin/metrics", config.HandleMetrics)
	serveMux.HandleFunc("POST /api/users", api.HandleCreateUser)
	serveMux.HandleFunc("POST /api/chirps", api.HandleCreateChirp)
<<<<<<< HEAD
	serveMux.HandleFunc("GET /api/chirps", api.HandleGetChirps)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", api.HandleGetChirp)
=======
	serveMux.HandleFunc("POST /api/login", api.HandleLogin)
>>>>>>> 4d8728a (auth 1)
	server.ListenAndServe()
}
