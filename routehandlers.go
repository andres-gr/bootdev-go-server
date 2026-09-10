package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type AppHandler struct {
	rootDir string
}

func (a AppHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	http.StripPrefix("/app", http.FileServer(http.Dir(a.rootDir))).ServeHTTP(w, req)
}

func handleHealthz(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(http.StatusText(http.StatusOK))); err != nil {
		log.Printf("w.Write: %v", err)
	}
}

func (conf *apiConfig) handleMetrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	res := fmt.Sprintf(`
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
	`, conf.fileserverHits.Load())

	if _, err := w.Write([]byte(res)); err != nil {
		log.Printf("w.Write: %v", err)
	}
}

const devEnv = "dev"

func (conf *apiConfig) handleReset(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if env := os.Getenv("PLATFORM"); env != devEnv {
		w.WriteHeader(http.StatusForbidden)

		if _, err := w.Write([]byte(http.StatusText(http.StatusForbidden))); err != nil {
			log.Printf("w.Write: %v", err)
		}
		return
	}

	if err := conf.db.ResetUsers(req.Context()); err != nil {
		respondWithInternalError(w)
		return
	}

	w.WriteHeader(http.StatusOK)

	conf.fileserverHits.Store(0)

	if _, err := w.Write([]byte(http.StatusText(http.StatusOK))); err != nil {
		log.Printf("w.Write: %v", err)
	}
}

var invalidWords = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

func handleValidateChirp(w http.ResponseWriter, req *http.Request) {
	defer closeReqBody(req)

	bod := struct {
		Body string `json:"body"`
	}{}

	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&bod); err != nil {
		respondWithInternalError(w)
		return
	}

	if len(bod.Body) > 140 {
		err := respondError(w, http.StatusBadRequest, "Chirp is too long")
		if err != nil {
			log.Printf("respondError: %v", err)
		}
		return
	}

	type result struct {
		CleanedBody string `json:"cleaned_body"`
	}

	words := strings.Split(bod.Body, " ")

	for i, w := range words {
		if w == "" {
			continue
		}

		if _, ok := invalidWords[strings.ToLower(w)]; ok {
			words[i] = "****"
		}
	}

	err := respondJSON(w, http.StatusOK, result{
		CleanedBody: strings.Join(words, " "),
	})
	if err != nil {
		log.Printf("respondJSON: %v", err)
	}
}

func (conf *apiConfig) handleCreateUser(w http.ResponseWriter, req *http.Request) {
	defer closeReqBody(req)

	bod := struct {
		Email string `json:"email"`
	}{}

	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&bod); err != nil {
		respondWithInternalError(w)
		return
	}

	if bod.Email == "" {
		err := respondError(w, http.StatusBadRequest, "Email is required")
		if err != nil {
			log.Printf("respondError: %v", err)
		}
		return
	}

	var result User
	user, err := conf.db.CreateUser(req.Context(), bod.Email)
	if err != nil {
		respondWithInternalError(w)
		return
	}

	result = User(user)

	err = respondJSON(w, http.StatusCreated, result)
	if err != nil {
		log.Printf("respondJSON: %v", err)
	}
}
