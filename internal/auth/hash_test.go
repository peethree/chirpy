package auth

import (
	"reflect"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	pw := []byte("Password123")

	hash, err := bcrypt.GenerateFromPassword(pw, 10)
	if err != nil {
		t.Fatalf("error generating password hash: %v", err)
	}

	if reflect.DeepEqual(pw, hash) {
		t.Fatalf("password and hash should not match")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	pw := []byte("Password123")

	hash, err := bcrypt.GenerateFromPassword(pw, 10)
	if err != nil {
		t.Fatalf("error generating password hash: %v", err)
	}

	// CheckPasswordHash(password, hash string) error {
	err = CheckPasswordHash(string(pw), string(hash))
	if err != nil {
		t.Fatalf("error checking pw hash: %v", err)
	}
}
