package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func CheckPasswordHash(password string, hash string) (bool, error) {
	valid, _, err := argon2id.CheckHash(password, hash)
	return valid, err
}

func CreateJWTToken(userID uuid.UUID, tokenSecret string, expires time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Issuer: "chirpy", IssuedAt: jwt.NewNumericDate(time.Now().UTC()), ExpiresAt: jwt.NewNumericDate(time.Now().Add(expires).UTC()), Subject: userID.String()})
	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString string, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("Error while parsing claims: %v", err)
	}
	if issuer, err := token.Claims.GetIssuer(); issuer != "chirpy" {
		if err != nil {
			return uuid.UUID{}, err
		}
		return uuid.UUID{}, fmt.Errorf("Unexpected issuer: %v", issuer)
	}
	subject, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("Error while getting claims: %v", err)
	}
	userId, err := uuid.Parse(subject)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("Error while parsing UUID: %v", err)
	}
	return userId, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("No Authorization header")
	}
	if !strings.Contains(authHeader, "Bearer ") {
		return "", errors.New("No Bearer token in Authorization header")
	}
	return strings.Split(authHeader, " ")[1], nil
}
func MakeRefreshToken() (string, error) {

	refreshToken := make([]byte, 32)
	rand.Read(refreshToken)

	return hex.EncodeToString(refreshToken), nil
}
