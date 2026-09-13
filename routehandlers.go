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
		log.Printf("GET /api/healthz - w.Write: %v", err)
	}
}

// GET /admin/metrics
func (conf *apiConfig) handleMetrics(w http.ResponseWriter, req *http.Request) {
	const msg = "GET /admin/metrics"

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
		log.Printf("%s - w.Write: %v", msg, err)
	}
}

const devEnv = "dev"

// POST /admin/reset
func (conf *apiConfig) handleReset(w http.ResponseWriter, req *http.Request) {
	const msg = "POST /admin/reset"

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if env := os.Getenv("PLATFORM"); env != devEnv {
		w.WriteHeader(http.StatusForbidden)

		if _, err := w.Write([]byte(http.StatusText(http.StatusForbidden))); err != nil {
			log.Printf("%s - w.Write: %v", msg, err)
		}
		return
	}

	if err := conf.db.ResetUsers(req.Context()); err != nil {
		respondWithInternalError(w, msg)
		return
	}

	w.WriteHeader(http.StatusOK)

	conf.fileserverHits.Store(0)

	if _, err := w.Write([]byte(http.StatusText(http.StatusOK))); err != nil {
		log.Printf("%s - w.Write: %v", msg, err)
	}
}

// POST /api/users
func (conf *apiConfig) handleCreateUser(w http.ResponseWriter, req *http.Request) {
	const msg = "POST /api/users"

	defer closeReqBody(req, msg)

	bod := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&bod); err != nil {
		respondWithInternalError(w, msg)
		return
	}

	if bod.Email == "" || bod.Password == "" {
		err := respondError(w, http.StatusBadRequest, "Email and password are required")
		if err != nil {
			log.Printf("%s - respondError: %v", msg, err)
		}
		return
	}

	if len(bod.Password) < 3 || strings.Contains(bod.Password, " ") || len(bod.Password) > 16 {
		err := respondError(w, http.StatusBadRequest, "Password must be between 3 and 16 characters and cannot contain spaces")
		if err != nil {
			log.Printf("%s - respondError: %v", msg, err)
		}
		return
	}

	pass, err := auth.HashPassword(bod.Password)
	if err != nil {
		respondWithInternalError(w, msg)
		return
	}

	user, err := conf.db.CreateUser(req.Context(), database.CreateUserParams{
		Email:          bod.Email,
		HashedPassword: pass,
	})
	if err != nil {
		respondWithInternalError(w, msg)
		return
	}

	token, err := auth.MakeJWT(user.ID, conf.jwtSecret, 1*time.Hour)
	if err != nil {
		respondWithInternalError(w, msg)
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
		log.Printf("%s - respondJSON: %v", msg, err)
	}
}

// POST /api/login
func (conf *apiConfig) handleLoginUser(w http.ResponseWriter, req *http.Request) {
	const msg = "POST /api/login"

	defer closeReqBody(req, msg)

	bod := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&bod); err != nil {
		respondWithInternalError(w, msg)
		return
	}

	if bod.Email == "" || bod.Password == "" {
		err := respondError(w, http.StatusBadRequest, "Email and password are required")
		if err != nil {
			log.Printf("%s - respondError: %v", msg, err)
		}
		return
	}

	user, err := conf.db.GetUserByEmail(req.Context(), bod.Email)
	if err != nil {
		respondWithInternalError(w, msg)
		return
	}

	valid, err := auth.CheckPassword(bod.Password, user.HashedPassword)
	if err != nil {
		respondWithInternalError(w, msg)
		return
	}

	if !valid {
		err := respondError(w, http.StatusUnauthorized, "incorrect email or password")
		if err != nil {
			log.Printf("%s - respondError: %v", msg, err)
		}
		return
	}

	token, err := auth.MakeJWT(user.ID, conf.jwtSecret, 1*time.Hour)
	if err != nil {
		respondWithInternalError(w, msg)
		return
	}

	rToken := auth.MakeRefreshToken()

	refresh, err := conf.db.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{
		Token:     rToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		respondWithInternalError(w, msg)
		return
	}

	res := cleanUserResponse(user, token, refresh)

	err = respondJSON(w, http.StatusOK, res)
	if err != nil {
		log.Printf("%s - respondJSON: %v", msg, err)
	}
}

// POST /api/refresh
func (conf *apiConfig) handleRefreshToken(w http.ResponseWriter, req *http.Request) {
	const msg = "POST /api/refresh"

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithInternalError(w, msg)
		return
	}

	rToken, err := conf.db.GetUserFromRefreshToken(req.Context(), token)
	if err != nil {
		err = respondError(w, http.StatusUnauthorized, "invalid refresh token")
		if err != nil {
			log.Printf("%s - respondError: %v", msg, err)
		}
		return
	}

	valid := rToken.ExpiresAt.After(time.Now()) && !rToken.RevokedAt.Valid
	if !valid {
		err = respondError(w, http.StatusUnauthorized, "invalid refresh token")
		if err != nil {
			log.Printf("%s - respondError: %v", msg, err)
		}
		return
	}

	jToken, err := auth.MakeJWT(rToken.UserID, conf.jwtSecret, time.Hour)
	if err != nil {
		respondWithInternalError(w, msg)
		return
	}

	err = respondJSON(w, http.StatusOK, struct {
		Token string `json:"token"`
	}{
		Token: jToken,
	})
	if err != nil {
		log.Printf("%s - respondJSON: %v", msg, err)
	}
}

// POST /api/revoke
func (conf *apiConfig) handleRevokeToken(w http.ResponseWriter, req *http.Request) {
	const msg = "POST /api/revoke"

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithInternalError(w, msg)
		return
	}

	_, err = conf.db.RevokeRefreshToken(req.Context(), token)
	if err != nil {
		respondWithInternalError(w, msg)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
	const msg = "POST /api/chirps"

	defer closeReqBody(req, msg)

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithInternalError(w, msg)
		return
	}

	id, err := auth.ValidateJWT(token, conf.jwtSecret)
	if err != nil {
		err = respondError(w, http.StatusUnauthorized, err.Error())
		if err != nil {
			log.Printf("%s - respondError: %v", msg, err)
		}
		return
	}

	content := struct {
		Body string `json:"body"`
	}{}

	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&content); err != nil {
		respondWithInternalError(w, msg)
		return
	}

	chirp, err := validateChirp(content.Body)
	if err != nil {
		err = respondError(w, http.StatusBadRequest, err.Error())
		if err != nil {
			log.Printf("%s - respondError: %v", msg, err)
		}
		return
	}

	res, err := conf.db.CreateChirp(req.Context(), database.CreateChirpParams{
		Body:   chirp,
		UserID: id,
	})
	if err != nil {
		respondWithInternalError(w, msg)
		return
	}

	err = respondJSON(w, http.StatusCreated, Chirp(res))
	if err != nil {
		log.Printf("%s - respondJSON: %v", msg, err)
	}
}

// GET /api/chirps
func (conf *apiConfig) handleGetChirps(w http.ResponseWriter, req *http.Request) {
	const msg = "GET /api/chirps"

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	res, err := conf.db.GetChirps(req.Context())
	if err != nil {
		respondWithInternalError(w, msg)
		return
	}

	var chirps []Chirp

	for _, chirp := range res {
		chirps = append(chirps, Chirp(chirp))
	}

	err = respondJSON(w, http.StatusOK, chirps)
	if err != nil {
		log.Printf("%s - respondJSON: %v", msg, err)
	}
}

// GET /api/chirps/{id}
func (conf *apiConfig) handleGetChirp(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	id, err := uuid.Parse(req.PathValue("id"))
	msg := "GET /api/chirps/" + id.String()

	if err != nil {
		respondWithInternalError(w, msg)
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
		log.Printf("%s - respondJSON: %v", msg, err)
	}
}
