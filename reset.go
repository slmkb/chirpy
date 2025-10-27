package main

import (
	"context"
	"log"
	"net/http"
)

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	if err := cfg.db.DropUsers(context.Background()); err != nil {
		log.Printf("create user: %v", err)
		internalError(w)
		return
	}
	w.Write([]byte("counter was reset"))
}
