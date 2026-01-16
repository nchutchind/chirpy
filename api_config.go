package main

import (
	"sync/atomic"

	"github.com/nchutchind/chirpy/internal/database"
)

type apiConfig struct {
	platform	   string
	fileserverHits atomic.Int32
	db             *database.Queries
	jwtSecret      string
}

