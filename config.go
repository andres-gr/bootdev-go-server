package main

import (
	"sync/atomic"

	"github.com/andres-gr/go-server/internal/database"
)

type apiConfig struct {
	db             *database.Queries
	fileserverHits atomic.Int32
	jwtSecret      string
	polkaKey       string
}
