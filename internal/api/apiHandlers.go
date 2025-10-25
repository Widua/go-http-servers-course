package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/widua/go-http-server/internal/auth"
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
	database.DB_Config.Queries.ResetUsers(context.Background())
	database.DB_Config.Queries.ResetChirps(context.Background())
	RespondOk(out)
}

func ValidateChirp(body string) (string, error) {

	if len(body) > 140 {
		return "", errors.New("Chirp is too long")
	}
	censoredChirp := ChirpClearer(body)
	return censoredChirp, nil
}

func ChirpClearer(body string) string {
	var profaneWords []string = []string{"kerfuffle", "sharbert", "fornax"}

	splittedBody := strings.Split(body, " ")

	for ix, word := range splittedBody {
		if slices.Contains(profaneWords, strings.ToLower(word)) {
			splittedBody[ix] = "****"
		}
	}
	censoredBody := strings.Join(splittedBody, " ")

	return censoredBody
}

func HandleHealthz(out http.ResponseWriter, req *http.Request) {
	RespondOk(out)
}
func HandleCreateUser(out http.ResponseWriter, req *http.Request) {
	type createUserBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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
	passwdHash, _ := auth.HashPassword(parsedBody.Password)
	usr, err := database.DB_Config.Queries.CreateUser(context.Background(), database.CreateUserParams{Email: parsedBody.Email, HashedPassword: passwdHash})

	user := User{ID: usr.ID, CreatedAt: usr.CreatedAt, UpdatedAt: usr.UpdatedAt, Email: usr.Email}
	byteBody, err := json.Marshal(user)
	if err != nil {
		RespondWithError(out, 400, err.Error())
		return
	}

	RespondWithJSON(out, 201, byteBody)

}
func HandleCreateChirp(out http.ResponseWriter, req *http.Request) {
	type createChirpBody struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	parsedReqBody := createChirpBody{}
	decoder := json.NewDecoder(req.Body)

	err := decoder.Decode(&parsedReqBody)
	if err != nil {
		RespondWithError(out, 400, err.Error())
		return
	}
	if parsedReqBody == (createChirpBody{}) {
		RespondWithError(out, 400, "Invalid body")
		return
	}

	chirp, err := database.DB_Config.Queries.CreateChirp(context.Background(), database.CreateChirpParams{Body: parsedReqBody.Body, UserID: parsedReqBody.UserID})
	mappedChirp := Chirp{ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, UserID: chirp.UserID}
	byteBody, err := json.Marshal(mappedChirp)
	if err != nil {
		RespondWithError(out, 400, err.Error())
		return
	}
	RespondWithJSON(out, 201, byteBody)

}
func HandleGetChirps(out http.ResponseWriter, req *http.Request) {
	chirps, err := database.DB_Config.Queries.GetAllChirps(context.Background())
	if err != nil {
		RespondWithError(out, 400, err.Error())
		return
	}
	mappedChirps := make([]Chirp, len(chirps))

	for ix, chirp := range chirps {
		mappedChirps[ix] = Chirp{ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, UserID: chirp.UserID}
	}
	mappedChirpsBytes, _ := json.Marshal(mappedChirps)

	RespondWithJSON(out, 200, mappedChirpsBytes)

}

func HandleGetChirp(out http.ResponseWriter, req *http.Request) {
	chirpID := req.PathValue("chirpID")
	err := uuid.Validate(chirpID)

	if err != nil {
		RespondWithError(out, 400, "It's not valid chirp ID")
		return
	}

	chirp, err := database.DB_Config.Queries.GetChirpByID(context.Background(), uuid.MustParse(chirpID))

	if err != nil {
		RespondWithError(out, 404, err.Error())
		return
	}

	if chirp == (database.Chirp{}) {
		RespondWithError(out, 404, "Chirp does not exist")
		return
	}

	mappedChirp := Chirp{ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, UserID: chirp.ID, Body: chirp.Body}
	jsonChirp, err := json.Marshal(mappedChirp)

	if err != nil {
		RespondWithError(out, 400, err.Error())
		return
	}
	RespondWithJSON(out, 200, jsonChirp)
}

func HandleLogin(out http.ResponseWriter, req *http.Request) {
	type loginRequestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	parsedReqBody := loginRequestBody{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&parsedReqBody)
	if err != nil {
		RespondWithError(out, 400, "Error handling login data")
	}

}
