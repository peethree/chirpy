package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	// extract api key from auth header
	apiKeyString := headers.Get("Authorization")

	if apiKeyString == "" {
		return "", errors.New("empty authorization field")
	}

	splitStrings := strings.Split(apiKeyString, " ")

	//format: Authorization: ApiKey THE_KEY_HERE, key is 2nd index of the splice
	apiKey := splitStrings[1]

	if apiKey == "" {
		return "", errors.New("no api key provided in request")
	}

	return apiKey, nil
}
