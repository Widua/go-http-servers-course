package auth

import (
	"fmt"
	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func CheckPasswordHash(password string, hash string) (bool, error) {
	valid, _, err := argon2id.CheckHash(password, hash)
	return valid, err
}

func CreateJWTToken(userID uuid.UUID, tokenSecret string, expires time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Issuer: "chirpy", IssuedAt: jwt.NewNumericDate(time.Now()), ExpiresAt: jwt.NewNumericDate(time.Now().Add(expires)), Subject: userID.String()})
	return token.SignedString(tokenSecret)
}

func ValidateJWT(tokenString string, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenSecret, jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}
		return tokenSecret, nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}
	if issuer, err := token.Claims.GetIssuer(); issuer != "chirpy" {
		if err != nil {
			return uuid.UUID{}, err
		}
		return uuid.UUID{}, fmt.Errorf("Unexpected issuer: %v", issuer)
	}
	subject, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}
	userId, err := uuid.Parse(subject)
	if err != nil {
		return uuid.UUID{}, err
	}
	return userId, nil

}
