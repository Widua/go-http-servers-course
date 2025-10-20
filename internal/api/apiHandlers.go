package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/widua/go-http-server/internal/database"
)

type ApiConfig struct {
	FileServerHits atomic.Int32
}

func HandleFileserver() http.Handler {
	return http.StripPrefix("/app", http.FileServer(http.Dir(".")))
}
func (cfg *ApiConfig) HandleMetrics(out http.ResponseWriter, req *http.Request) {
	metricsTemplate := ` 
	<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>
	`
	out.Header().Add("Content-Type", "text/html; charset=utf-8")
	out.WriteHeader(http.StatusOK)
	out.Write([]byte(fmt.Sprintf(metricsTemplate, cfg.FileServerHits.Load())))
}

func (cfg *ApiConfig) HandleReset(out http.ResponseWriter, req *http.Request) {
	cfg.FileServerHits.Store(0)
	RespondOk(out)
}

func ValidateChirp(out http.ResponseWriter, req *http.Request) {
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
		RespondWithError(out, 400, err.Error())
		return
	}
	body := params.Body

	if len(body) > 140 {
		RespondWithError(out, 400, "Chirp is too long")
		return
	}
	ChirpClearer(out, body)
}

func ChirpClearer(out http.ResponseWriter, body string) {
	var profaneWords []string = []string{"kerfuffle", "sharbert", "fornax"}
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
		RespondWithError(out, 400, err.Error())
		return
	}

	RespondWithJSON(out, 200, byteBody)
}

func HandleHealthz(out http.ResponseWriter, req *http.Request) {
	RespondOk(out)
}
func HandleCreateUser(out http.ResponseWriter, req *http.Request) {
	type createUserBody struct {
		Email string `json:"email"`
	}
	parsedBody := createUserBody{}
	decoder := json.NewDecoder(req.Body)

	err := decoder.Decode(&parsedBody)
	if err != nil {
		RespondWithError(out, 400, err.Error())
		return
	}
	if parsedBody == (createUserBody{}) {
		RespondWithError(out, 400, "Invalid body")
		return
	}
	usr, err := database.DB_Config.Queries.CreateUser(context.Background(), parsedBody.Email)

	user := User{ID: usr.ID, CreatedAt: usr.CreatedAt, UpdatedAt: usr.UpdatedAt, Email: usr.Email}
	byteBody, err := json.Marshal(user)
	if err != nil {
		RespondWithError(out, 400, err.Error())
		return
	}

	RespondWithJSON(out, 201, byteBody)

}
