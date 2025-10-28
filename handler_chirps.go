package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type returnVals struct {
	ID         uuid.UUID `json:"id,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
	CleaneBody string    `json:"body,omitempty"`
	UserID     uuid.UUID `json:"user_id,omitempty"`
	Error      string    `json:"error,omitempty"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("validate chirp: %v", err)
		internalError(w)
		return
	}
	if token == "" {
		http.Error(w, "Unathorized", http.StatusUnauthorized)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		log.Printf("validate chirp: %v", err)
		http.Error(w, "Unathorized", http.StatusUnauthorized)
		return
	}
	if userID == uuid.Nil {
		http.Error(w, "Unathorized", http.StatusUnauthorized)
		return
	}

	type parmeters struct {
		Chirp string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parmeters{}
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("validate chirp: %v", err)
		internalError(w)
		return
	}

	if len(params.Chirp) > 140 {
		respBody := returnVals{
			Error: "Chrip is too long",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("validate chrip: %v", err)
			internalError(w)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(dat)
		return
	}

	respBody := returnVals{
		CleaneBody: filterChrip(params.Chirp),
	}

	chirp, err := cfg.db.CreateChirp(context.Background(), database.CreateChirpParams{
		Body:   respBody.CleaneBody,
		UserID: userID,
	})
	if err != nil {
		log.Printf("validate chrip: %v %+v", err, params)
		internalError(w)
		return
	}

	respBody.ID = chirp.ID
	respBody.CreatedAt = chirp.CreatedAt
	respBody.UpdatedAt = chirp.UpdatedAt
	respBody.UserID = chirp.UserID
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("validate chrip: %v", err)
		internalError(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(dat)
}

func filterChrip(text string) string {
	profanities := []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}

	var filtered []string
	for _, s := range strings.Fields(text) {
		for _, p := range profanities {
			if p == strings.ToLower(s) {
				s = "****"
			}
		}
		filtered = append(filtered, s)
	}
	return strings.Join(filtered, " ")
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetAllChirps(context.Background())
	if err != nil {
		log.Printf("GetAllChirps: %v", err)
		internalError(w)
		return
	}

	var respBody []returnVals
	for _, chirp := range chirps {
		respBody = append(respBody, returnVals{
			ID:         chirp.ID,
			CreatedAt:  chirp.CreatedAt,
			UpdatedAt:  chirp.UpdatedAt,
			CleaneBody: chirp.Body,
			UserID:     chirp.UserID,
		})
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("validate chrip: %v", err)
		internalError(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}

func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	parsedId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("validate chrip: %v", err)
		internalError(w)
		return
	}
	chirp, err := cfg.db.GetChirpByID(context.Background(), parsedId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "", http.StatusNotFound)
			return
		}
		log.Printf("validate chrip: %v", err)
		internalError(w)
		return
	}
	var respBody returnVals
	respBody.ID = chirp.ID
	respBody.CreatedAt = chirp.CreatedAt
	respBody.UpdatedAt = chirp.UpdatedAt
	respBody.CleaneBody = chirp.Body
	respBody.UserID = chirp.UserID
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("validate chrip: %v", err)
		internalError(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	parsedId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("validate chrip: %v", err)
		internalError(w)
		return
	}
	chirp, err := cfg.db.GetChirpByID(context.Background(), parsedId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "", http.StatusNotFound)
			return
		}
		log.Printf("validate chrip: %v", err)
		internalError(w)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("delete chrip: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		log.Printf("delete chrip: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if userID != chirp.UserID {
		log.Printf("delete chrip: %v", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err = cfg.db.DeleteChirp(context.Background(), database.DeleteChirpParams{
		UserID: userID,
		ID:     chirp.ID,
	})
	if err != nil {
		log.Printf("delete chrip: %v", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
