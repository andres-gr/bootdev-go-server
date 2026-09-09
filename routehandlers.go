package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

func (conf *apiConfig) handleReset(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	conf.fileserverHits.Store(0)
}

func handleValidateChirp(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := req.Body.Close(); err != nil {
			log.Printf("req.Body.Close: %v", err)
		}
	}()

	bod := struct {
		Body string `json:"body"`
	}{}

	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&bod); err != nil {
		err = respondError(w, http.StatusInternalServerError, "Something went wrong")
		if err != nil {
			log.Printf("respondError: %v", err)
		}
		return
	}

	if len(bod.Body) > 140 {
		err := respondError(w, http.StatusBadRequest, "Chirp is too long")
		if err != nil {
			log.Printf("respondError: %v", err)
		}
		return
	}

	err := respondJSON(w, http.StatusOK, struct {
		Valid bool `json:"valid"`
	}{
		Valid: true,
	})
	if err != nil {
		log.Printf("respondJSON: %v", err)
	}
}
