package api

import (
	"encoding/json"
	"net/http"
)

func RespondWithError(out http.ResponseWriter, statusCode int, errorMsg string) {
	type apiError struct {
		Error string `json:"error"`
	}

	out.WriteHeader(statusCode)
	errorMess := apiError{Error: errorMsg}
	body, _ := json.Marshal(errorMess)
	out.Write(body)
}

func RespondWithJSON(out http.ResponseWriter, statusCode int, responseBody []byte) {
	out.WriteHeader(statusCode)
	out.Header().Add("Content-Type", "Application/Json")
	out.Write(responseBody)

}

func RespondNoContent(out http.ResponseWriter, statusCode int) {
	out.WriteHeader(statusCode)
}

func RespondOk(out http.ResponseWriter) {
	out.WriteHeader(200)
	out.Header().Add("Content-Type", "text/plain")
	out.Write([]byte("OK"))
}
