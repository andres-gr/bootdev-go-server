package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/andres-gr/go-server/internal/database"
)

func respondJSON(w http.ResponseWriter, code int, payload interface{}) error {
	res, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if _, err := w.Write(res); err != nil {
		return err
	}

	return nil
}

func respondError(w http.ResponseWriter, code int, msg string) error {
	return respondJSON(w, code, map[string]string{"error": msg})
}

func closeReqBody(req *http.Request, msg string) {
	if err := req.Body.Close(); err != nil {
		log.Printf("%s - req.Body.Close: %v", msg, err)
	}
}

func respondWithInternalError(w http.ResponseWriter, msg ...string) {
	err := respondError(w, http.StatusInternalServerError, "Something went wrong")
	if err != nil {
		log.Printf("%s - respondError: %v", msg, err)
	}
}

func cleanUserResponse(user database.User, token string, refresh string) User {
	return User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		IsChirpyRed:  user.IsChirpyRed,
		Token:        token,
		RefreshToken: refresh,
	}
}
