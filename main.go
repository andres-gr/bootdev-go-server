package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	const (
		port    = "8080"
		rootDir = "."
	)

	conf := &apiConfig{}

	idleConnsClosed := make(chan struct{})
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	mux.Handle("/app/", conf.middlewareMetricsInc(AppHandler{rootDir: rootDir}))
	mux.HandleFunc("/healthz", handleHealthz)
	mux.HandleFunc("/metrics", conf.handleMetrics)
	mux.HandleFunc("/reset", conf.handleReset)

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}

		close(idleConnsClosed)
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed
}
