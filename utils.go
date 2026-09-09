package main

import (
	"encoding/json"
	"net/http"
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
