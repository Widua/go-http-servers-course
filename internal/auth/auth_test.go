package auth

import (
	"testing"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
)

func TestHashing(t *testing.T) {
	password := "testPasswd1"

	hash, err := HashPassword(password)
	if err != nil {
		t.Errorf("Error occurs while hashing: %v", err)
	}
	valid, _, err := argon2id.CheckHash(password, hash)
	if err != nil {
		t.Errorf("Error occurs while checking hash: %v", err)
	}
	if !valid {
		t.Errorf("Password, and created hash does not match")
	}
}

func TestJWTTokenCycle(t *testing.T) {
	userId := uuid.New()
	tokenSecret := "secret"
	expires := time.Second * 5

	token, err := CreateJWTToken(userId, tokenSecret, expires)
	if err != nil {
		t.Errorf("Function should create JWT token, but produces error: %v", err)
	}

	validatedUUID, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Errorf("JWT validation should successfully validate JWT produced with CreateJWTToken, but produces error: %v", err)
	}
	if validatedUUID != userId {
		t.Errorf("%v should be equal to %v", userId, validatedUUID)
	}
}

func TestJWTTokenCycleExpired(t *testing.T) {
	userId := uuid.New()
	tokenSecret := "secret"
	expires := time.Millisecond

	token, _ := CreateJWTToken(userId, tokenSecret, expires)
	time.Sleep(time.Microsecond * 2)
	validatedUUID, err := ValidateJWT(token, tokenSecret)
	if err == nil {
		t.Errorf("JWT Validation should terminate validation when the time from expires parameter pass, and not return with validated UUID, but returned: %v", validatedUUID)
	}

}
