package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/andres-gr/go-server/internal/database"
)

func main() {
	const (
		port    = "8080"
		rootDir = "."
	)

	if err := godotenv.Load(); err != nil {
		log.Fatalf("godotenv.Load: %v", err)
		os.Exit(1)
	}

	dbUri := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbUri)
	if err != nil {
		log.Fatalf("sql.Open: %v", err)
		os.Exit(1)
	}

	dbQueries := database.New(db)
	conf := &apiConfig{
		db:             dbQueries,
		fileserverHits: atomic.Int32{},
		jwtSecret:      os.Getenv("JWT_SECRET"),
	}

	idleConnsClosed := make(chan struct{})
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	mux.Handle("/app/", conf.middlewareMetricsInc(AppHandler{rootDir: rootDir}))

	mux.HandleFunc("GET /api/healthz", handleHealthz)

	mux.HandleFunc("GET /api/chirps/{id}", conf.handleGetChirp)
	mux.HandleFunc("GET /api/chirps", conf.handleGetChirps)
	mux.HandleFunc("POST /api/chirps", conf.handleAddChirp)

	mux.HandleFunc("POST /api/users", conf.handleCreateUser)

	mux.HandleFunc("POST /api/login", conf.handleLoginUser)
	mux.HandleFunc("POST /api/refresh", conf.handleRefreshToken)
	mux.HandleFunc("POST /api/revoke", conf.handleRevokeToken)

	mux.HandleFunc("GET /admin/metrics", conf.handleMetrics)
	mux.HandleFunc("POST /admin/reset", conf.handleReset)

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
