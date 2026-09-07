package main

import (
	"log"
	"net/http"
)

type HealthHandler struct{}

func (HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte("OK")); err != nil {
		log.Printf("w.Write: %v", err)
	}
}
