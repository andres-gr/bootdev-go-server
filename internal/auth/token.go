package auth

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
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
