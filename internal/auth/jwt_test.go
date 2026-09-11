package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

var (
	expires = time.Duration(2 * time.Second)
	secret  = "testing_secret"
)

func TestJWT(t *testing.T) {
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}

	token, err := MakeJWT(id, secret, expires)
	if err != nil {
		t.Fatal(err)
	}

	res, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatal(err)
	}

	if res != id {
		t.Fatalf("expected %s, got %s", id, res)
	}
}

func TestExpiredJWT(t *testing.T) {
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}

	token, err := MakeJWT(id, secret, expires)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(3 * time.Second)

	_, err = ValidateJWT(token, secret)
	if err == nil {
		t.Fatal("expected error, got none")
	}
}
