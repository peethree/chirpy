package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	// SigningMethod: jwt.SigningMethodHS256
	// func NewWithClaims(method SigningMethod, claims Claims, opts ...TokenOption) *Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		Issuer:    "chirpy",
		Subject:   userID.String(),
	})

	// create JWT
	// func (t *Token) SignedString(key interface{}) (string, error)
	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// func to validate the created jwt
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {

	// func ParseWithClaims(tokenString string, claims Claims, keyFunc Keyfunc, options ...ParserOption) (*Token, error)
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	// return empty token + err in case of failed parse
	if err != nil {
		// 401 Unauthorized
		return uuid.Nil, err
	}

	// first get the claims interface
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return uuid.Nil, errors.New("error getting claim interface")
	}

	// get the subject from claims, parse the user id to type uuid.UUID
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	// information comes into the Authorization header
	// looks like: Bearer TOKEN_STRING
	auth := headers.Get("Authorization")

	// empty auth header
	if auth == "" {
		return "", errors.New("auth header missing")
	}

	// extract the token from the header
	splitAuth := strings.Split(auth, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "Bearer" {
		return "", errors.New("incorrect bearer token format")
	}

	// token should be 2nd index
	token := splitAuth[1]

	// remove white space
	cleanToken := strings.TrimSpace(token)

	return cleanToken, nil
}
