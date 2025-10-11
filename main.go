package main

import (
	"net/http"
	"os"
)

func main() {
	serveMux := http.NewServeMux()
	server := http.Server{
		Handler: serveMux,
		Addr:    ":8080",
	}
	serveMux.HandleFunc("/app", handleApp)
	serveMux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	serveMux.HandleFunc("/healthz", handleHealthz)
	serveMux.Handle("/", http.FileServer(http.Dir(".")))
	server.ListenAndServe()
}

func handleApp(responseWriter http.ResponseWriter, req *http.Request) {
	responseWriter.WriteHeader(200)
	responseWriter.Header().Add("Content-Type", "application/html")
	res, err := os.ReadFile("index.html")
	if err != nil {
		return
	}
	responseWriter.Write(res)
}

func handleHealthz(responseWriter http.ResponseWriter, req *http.Request) {
	responseWriter.WriteHeader(200)
	responseWriter.Header().Add("Content-Type", "text/plain")
	responseWriter.Write([]byte("OK"))
}
