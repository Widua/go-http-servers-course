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
	tokenSecret := os.Getenv("JWT_SECRET")
	polkaKey := os.Getenv("POLKA_KEY")
	dbconfig := database.InitializeDatabase(dbUrl)

	serveMux := http.NewServeMux()
	server := http.Server{
		Handler: serveMux,
		Addr:    ":8080",
	}
	config := api.ApiConfig{FileServerHits: atomic.Int32{}, JWT_Secret: tokenSecret, DB_Config: &dbconfig, POLKA_KEY: polkaKey}
	serveMux.Handle("/app/", config.MetricsMiddleware(api.HandleFileserver()))
	serveMux.HandleFunc("POST /admin/reset", config.HandleReset)
	serveMux.HandleFunc("GET /api/healthz", config.HandleHealthz)
	serveMux.HandleFunc("GET /admin/metrics", config.HandleMetrics)
	serveMux.HandleFunc("POST /api/users", config.HandleCreateUser)
	serveMux.HandleFunc("POST /api/chirps", config.HandleCreateChirp)
	serveMux.HandleFunc("GET /api/chirps", config.HandleGetChirps)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", config.HandleGetChirp)
	serveMux.HandleFunc("DELETE /api/chirps/{chirpID}", config.HandleDeleteChirp)
	serveMux.HandleFunc("POST /api/login", config.HandleLogin)
	serveMux.HandleFunc("POST /api/refresh", config.HandleRefreshToken)
	serveMux.HandleFunc("POST /api/revoke", config.HandleRevokeToken)
	serveMux.HandleFunc("PUT /api/users", config.HandleUpdateUser)
	serveMux.HandleFunc("POST /api/polka/webhooks", config.HandlePolkaWebhooks)
	server.ListenAndServe()
}
