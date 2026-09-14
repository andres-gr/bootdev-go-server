package auth

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	ApiKeyPrefix        = "ApiKey "
)

func GetBearerToken(headers http.Header) (token string, err error) {
	bearer := headers.Get(AuthorizationHeader)

	if bearer == "" {
		err = fmt.Errorf("authorization header not found")
		return
	}

	if !strings.HasPrefix(bearer, BearerPrefix) {
		err = fmt.Errorf("invalid authorization header")
		return
	}

	token = strings.TrimPrefix(bearer, BearerPrefix)

	if token == "" {
		err = fmt.Errorf("token not found")
		return
	}

	return
}

func GetApiKey(headers http.Header) (key string, err error) {
	apiKey := headers.Get(AuthorizationHeader)

	if apiKey == "" {
		err = fmt.Errorf("authorization header not found")
		return
	}

	if !strings.HasPrefix(apiKey, ApiKeyPrefix) {
		err = fmt.Errorf("invalid authorization header")
		return
	}

	key = strings.TrimPrefix(apiKey, ApiKeyPrefix)

	if key == "" {
		err = fmt.Errorf("api key not found")
		return
	}

	return
}
