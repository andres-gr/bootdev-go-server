// Package auth provides authentication and authorization
package auth

import (
	"fmt"

	"github.com/alexedwards/argon2id"
)

func HashPassword(pass string) (hash string, err error) {
	if pass == "" {
		return hash, fmt.Errorf("password cannot be empty")
	}

	hash, err = argon2id.CreateHash(pass, argon2id.DefaultParams)

	return
}

func CheckPassword(pass, hash string) (valid bool, err error) {
	if pass == "" {
		return valid, fmt.Errorf("password cannot be empty")
	}

	if hash == "" {
		return valid, fmt.Errorf("hash cannot be empty")
	}

	valid, err = argon2id.ComparePasswordAndHash(pass, hash)

	return
}
