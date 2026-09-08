package main

import "net/http"

func (conf *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conf.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
