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

func TestCheckPasswordHash_(t *testing.T) {
	// First, we need to create some hashed passwords for testing
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "Correct password",
			password: password1,
			hash:     hash1,
			wantErr:  false,
		},
		{
			name:     "Incorrect password",
			password: "wrongPassword",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Password doesn't match different hash",
			password: password1,
			hash:     hash2,
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Invalid hash",
			password: password1,
			hash:     "invalidhash",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
