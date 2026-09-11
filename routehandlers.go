package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/andres-gr/go-server/internal/auth"
	"github.com/andres-gr/go-server/internal/database"
)

type AppHandler struct {
	rootDir string
}

// GET /app/
func (a AppHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	http.StripPrefix("/app", http.FileServer(http.Dir(a.rootDir))).ServeHTTP(w, req)
}

// GET /api/healthz
func handleHealthz(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(http.StatusText(http.StatusOK))); err != nil {
		log.Printf("w.Write: %v", err)
	}
}

// GET /admin/metrics
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

// POST /admin/reset
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

// POST /api/users
func (conf *apiConfig) handleCreateUser(w http.ResponseWriter, req *http.Request) {
	defer closeReqBody(req)

	bod := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&bod); err != nil {
		respondWithInternalError(w)
		return
	}

	if bod.Email == "" || bod.Password == "" {
		err := respondError(w, http.StatusBadRequest, "Email and password are required")
		if err != nil {
			log.Printf("respondError: %v", err)
		}
		return
	}

	if len(bod.Password) < 3 || strings.Contains(bod.Password, " ") || len(bod.Password) > 16 {
		err := respondError(w, http.StatusBadRequest, "Password must be between 3 and 16 characters and cannot contain spaces")
		if err != nil {
			log.Printf("respondError: %v", err)
		}
		return
	}

	pass, err := auth.HashPassword(bod.Password)
	if err != nil {
		respondWithInternalError(w)
		return
	}

	user, err := conf.db.CreateUser(req.Context(), database.CreateUserParams{
		Email:          bod.Email,
		HashedPassword: pass,
	})
	if err != nil {
		respondWithInternalError(w)
		return
	}

	token, err := auth.MakeJWT(user.ID, conf.jwtSecret, time.Hour)
	if err != nil {
		respondWithInternalError(w)
		return
	}

	res := struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
	}{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	}

	err = respondJSON(w, http.StatusCreated, res)
	if err != nil {
		log.Printf("respondJSON: %v", err)
	}
}

// POST /api/login
func (conf *apiConfig) handleLoginUser(w http.ResponseWriter, req *http.Request) {
	defer closeReqBody(req)

	bod := struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds,omitempty"`
	}{}

	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&bod); err != nil {
		respondWithInternalError(w)
		return
	}

	if bod.Email == "" || bod.Password == "" {
		err := respondError(w, http.StatusBadRequest, "Email and password are required")
		if err != nil {
			log.Printf("respondError: %v", err)
		}
		return
	}

	user, err := conf.db.GetUserByEmail(req.Context(), bod.Email)
	if err != nil {
		respondWithInternalError(w)
		return
	}

	valid, err := auth.CheckPassword(bod.Password, user.HashedPassword)
	if err != nil {
		respondWithInternalError(w)
		return
	}

	if !valid {
		err := respondError(w, http.StatusUnauthorized, "incorrect email or password")
		if err != nil {
			log.Printf("respondError: %v", err)
		}
		return
	}

	expires := time.Duration(bod.ExpiresInSeconds) * time.Second
	if expires == 0 || expires > time.Hour {
		expires = time.Hour
	}

	token, err := auth.MakeJWT(user.ID, conf.jwtSecret, expires)
	if err != nil {
		respondWithInternalError(w)
		return
	}

	res := cleanUserResponse(user, token)

	err = respondJSON(w, http.StatusOK, res)
	if err != nil {
		log.Printf("respondJSON: %v", err)
	}
}

var invalidWords = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

func validateChirp(chirp string) (string, error) {
	if len(chirp) > 140 {
		return "", fmt.Errorf("Chirp is too long")
	}

	if len(chirp) == 0 {
		return "", fmt.Errorf("Chirp is empty")
	}

	words := strings.Split(chirp, " ")

	for i, w := range words {
		if w == "" {
			continue
		}

		if _, ok := invalidWords[strings.ToLower(w)]; ok {
			words[i] = "****"
		}
	}

	return strings.Join(words, " "), nil
}

// POST /api/chirps
func (conf *apiConfig) handleAddChirp(w http.ResponseWriter, req *http.Request) {
	defer closeReqBody(req)

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithInternalError(w)
		return
	}

	id, err := auth.ValidateJWT(token, conf.jwtSecret)
	if err != nil {
		err = respondError(w, http.StatusUnauthorized, err.Error())
		if err != nil {
			log.Printf("respondError: %v", err)
		}
		return
	}

	content := struct {
		Body string `json:"body"`
	}{}

	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&content); err != nil {
		respondWithInternalError(w)
		return
	}

	chirp, err := validateChirp(content.Body)
	if err != nil {
		err = respondError(w, http.StatusBadRequest, err.Error())
		if err != nil {
			log.Printf("respondError: %v", err)
		}
		return
	}

	res, err := conf.db.CreateChirp(req.Context(), database.CreateChirpParams{
		Body:   chirp,
		UserID: id,
	})
	if err != nil {
		respondWithInternalError(w)
		return
	}

	err = respondJSON(w, http.StatusCreated, Chirp(res))
	if err != nil {
		log.Printf("respondJSON: %v", err)
	}
}

// GET /api/chirps
func (conf *apiConfig) handleGetChirps(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	res, err := conf.db.GetChirps(req.Context())
	if err != nil {
		respondWithInternalError(w)
		return
	}

	var chirps []Chirp

	for _, chirp := range res {
		chirps = append(chirps, Chirp(chirp))
	}

	err = respondJSON(w, http.StatusOK, chirps)
	if err != nil {
		log.Printf("respondJSON: %v", err)
	}
}

// GET /api/chirps/{id}
func (conf *apiConfig) handleGetChirp(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	id, err := uuid.Parse(req.PathValue("id"))
	if err != nil {
		respondWithInternalError(w)
		return
	}

	res, err := conf.db.GetChirp(req.Context(), id)
	if err != nil {
		err = respondError(w, http.StatusNotFound, err.Error())
		if err != nil {
			log.Printf("respondError: %v", err)
		}
		return
	}

	err = respondJSON(w, http.StatusOK, Chirp(res))
	if err != nil {
		log.Printf("respondJSON: %v", err)
	}
}
