package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(id uuid.UUID, secret string, expires time.Duration) (token string, err error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expires)),
		Subject:   id.String(),
	}
	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))

	return
}

func ValidateJWT(token, secret string) (id uuid.UUID, err error) {
	claims := jwt.RegisteredClaims{}
	t, err := jwt.ParseWithClaims(token, &claims, func(*jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return
	}

	if !t.Valid {
		err = fmt.Errorf("invalid token")
		return
	}

	date, err := t.Claims.GetExpirationTime()
	if err != nil {
		return
	}

	if date.Before(time.Now().UTC()) {
		err = fmt.Errorf("token expired")
		return
	}

	sub, err := t.Claims.GetSubject()
	if err != nil {
		return
	}

	id, err = uuid.Parse(sub)

	return
}
