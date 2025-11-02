package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/widua/go-http-server/internal/auth"
	"github.com/widua/go-http-server/internal/database"
)

type ApiConfig struct {
	FileServerHits atomic.Int32
	JWT_Secret     string
	DB_Config      *database.DatabaseConfig
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
	cfg.DB_Config.Queries.ResetUsers(context.Background())
	cfg.DB_Config.Queries.ResetChirps(context.Background())
	RespondOk(out)
}

func (cfg *ApiConfig) ValidateChirp(body string) (string, error) {

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

func (cfg *ApiConfig) HandleHealthz(out http.ResponseWriter, req *http.Request) {
	RespondOk(out)
}
func (cfg *ApiConfig) HandleCreateUser(out http.ResponseWriter, req *http.Request) {
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
	usr, err := cfg.DB_Config.Queries.CreateUser(context.Background(), database.CreateUserParams{Email: parsedBody.Email, HashedPassword: passwdHash})

	user := RegisterFromDatabaseUser(usr)
	byteBody, err := json.Marshal(user)
	if err != nil {
		RespondWithError(out, 400, err.Error())
		return
	}

	RespondWithJSON(out, 201, byteBody)

}
func (cfg *ApiConfig) HandleCreateChirp(out http.ResponseWriter, req *http.Request) {
	type createChirpBody struct {
		Body string `json:"body"`
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
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		RespondWithError(out, 401, err.Error())
		return
	}
	tokenuuid, err := auth.ValidateJWT(token, cfg.JWT_Secret)
	if err != nil {
		RespondWithError(out, 401, err.Error())
		return
	}

	chirp, err := cfg.DB_Config.Queries.CreateChirp(context.Background(), database.CreateChirpParams{Body: parsedReqBody.Body, UserID: tokenuuid})
	mappedChirp := Chirp{ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, UserID: chirp.UserID}
	byteBody, err := json.Marshal(mappedChirp)
	if err != nil {
		RespondWithError(out, 400, err.Error())
		return
	}
	RespondWithJSON(out, 201, byteBody)

}
func (cfg *ApiConfig) HandleGetChirps(out http.ResponseWriter, req *http.Request) {
	chirps, err := cfg.DB_Config.Queries.GetAllChirps(context.Background())
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

func (cfg *ApiConfig) HandleGetChirp(out http.ResponseWriter, req *http.Request) {
	chirpID := req.PathValue("chirpID")
	err := uuid.Validate(chirpID)

	if err != nil {
		RespondWithError(out, 400, "It's not valid chirp ID")
		return
	}

	chirp, err := cfg.DB_Config.Queries.GetChirpByID(context.Background(), uuid.MustParse(chirpID))

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

func (cfg *ApiConfig) HandleLogin(out http.ResponseWriter, req *http.Request) {
	type loginRequestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	parsedReqBody := loginRequestBody{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&parsedReqBody)
	if err != nil {
		RespondWithError(out, 400, "Error handling login data")
		return
	}
	usr, err := cfg.DB_Config.Queries.GetUserByEmail(context.Background(), parsedReqBody.Email)
	if err != nil {
		RespondWithError(out, 400, "User does not exist")
		return
	}
	valid, _ := auth.CheckPasswordHash(parsedReqBody.Password, usr.HashedPassword)
	if !valid {
		RespondWithError(out, 401, "Wrong password")
		return
	}
	token, err := auth.CreateJWTToken(usr.ID, cfg.JWT_Secret, 3600*time.Second)
	if err != nil {
		RespondWithError(out, 401, err.Error())
		return
	}

	refreshToken, _ := auth.MakeRefreshToken()
	refreshTokenDB, err := cfg.DB_Config.Queries.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{Token: refreshToken, UserID: usr.ID})
	if err != nil {
		RespondWithError(out, 400, err.Error())
		return
	}
	fmt.Printf("REFRESHTOKEN: %v", refreshToken)

	user := FromDatabaseUser(usr, token, refreshTokenDB.Token)
	jsonUser, err := json.Marshal(user)
	if err != nil {
		RespondWithError(out, 400, err.Error())
		return
	}

	RespondWithJSON(out, 200, jsonUser)
}

func (cfg *ApiConfig) HandleRefreshToken(out http.ResponseWriter, req *http.Request) {
	type tokenResponse struct {
		Token string `json:"token"`
	}
	token := tokenResponse{}

	refreshToken, err := auth.GetBearerToken(req.Header)

	if err != nil {
		RespondWithError(out, 400, err.Error())
		return
	}

	refreshTokenData, err := cfg.DB_Config.Queries.GetRefreshTokenByToken(context.Background(), refreshToken)

	if err != nil {
		RespondWithError(out, 401, err.Error())
		return
	}
	if refreshTokenData.RevokedAt != (sql.NullTime{}) {
		RespondWithError(out, 401, "That refresh token is revoked")
		return
	}

	jwt, err := auth.CreateJWTToken(refreshTokenData.UserID, cfg.JWT_Secret, 3600*time.Second)
	if err != nil {
		RespondWithError(out, 401, err.Error())
		return

	}
	token.Token = jwt

	resByte, _ := json.Marshal(token)
	RespondWithJSON(out, 200, resByte)
}

func (cfg *ApiConfig) HandleRevokeToken(out http.ResponseWriter, req *http.Request) {
	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		RespondWithError(out, 400, err.Error())
		return
	}
	err = cfg.DB_Config.Queries.RevokeAccessToToken(context.Background(), refreshToken)
	if err != nil {
		RespondWithError(out, 400, err.Error())
		return
	}

	RespondNoContent(out, 204)
}

func (cfg *ApiConfig) HandleUpdateUser(out http.ResponseWriter, req *http.Request) {
	type updateData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	reqUpdateData := updateData{}
	apiToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		RespondWithError(out, 401, "Token missing")
		return
	}
	decoder := json.NewDecoder(req.Body)
	decoder.Decode(&reqUpdateData)

	userId, err := auth.ValidateJWT(apiToken, cfg.JWT_Secret)
	if err != nil {
		RespondWithError(out, 401, "Invalid Token")
		return
	}

	hashedPassword, _ := auth.HashPassword(reqUpdateData.Password)

	err = cfg.DB_Config.Queries.UpdateUser(context.Background(), database.UpdateUserParams{Email: reqUpdateData.Email, HashedPassword: hashedPassword, ID: userId})
	if err != nil {
		RespondWithError(out, 401, "Problem while updating User")
		fmt.Printf("%v", err)
		return
	}

	updatedUser, _ := cfg.DB_Config.Queries.GetUserByID(context.Background(), userId)
	mappedUpser := RegisterFromDatabaseUser(updatedUser)
	parsedJsonUser, _ := json.Marshal(mappedUpser)

	RespondWithJSON(out, 200, parsedJsonUser)
}

func (cfg *ApiConfig) HandleDeleteChirp(out http.ResponseWriter, req *http.Request) {
	apiToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		RespondWithError(out, 401, "Token missing")
		return
	}
	userId, err := auth.ValidateJWT(apiToken, cfg.JWT_Secret)
	if err != nil {
		RespondWithError(out, 401, "Invalid token")
		return
	}

	chirpID := req.PathValue("chirpID")
	parsedChirp, err := uuid.Parse(chirpID)
	if err != nil {
		RespondWithError(out, 404, "Invalid ChirpID")
		return
	}

	chirp, err := cfg.DB_Config.Queries.GetChirpByID(context.Background(), parsedChirp)

	if err != nil {
		RespondWithError(out, 404, "Chirp does not exist")
		return
	}

	if chirp.UserID != userId {
		RespondWithError(out, 403, "It isn't your chirp")
		return
	}

	err = cfg.DB_Config.Queries.DeleteChirpByID(context.Background(), chirp.ID)

	if err != nil {
		RespondWithError(out, 400, err.Error())
		return
	}

	RespondNoContent(out, 204)
}
