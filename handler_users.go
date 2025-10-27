package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"chirpy/internal/session"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		log.Printf("create user: %v", err)
		internalError(w)
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("create user: %v", err)
		internalError(w)
		return
	}

	user, err := cfg.db.CreateUser(context.Background(), database.CreateUserParams{
		Email:        req.Email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		log.Printf("create user: %v", err)
		internalError(w)
		return
	}

	dat, err := json.Marshal(struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})

	if err != nil {
		log.Printf("create user: %v", err)
		internalError(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(dat)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		log.Printf("login: %v", err)
		internalError(w)
		return
	}

	user, err := cfg.db.GetUser(context.Background(), req.Email)
	if err != nil {
		log.Printf("login: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = auth.CheckPasswordHash(req.Password, user.PasswordHash)
	if err != nil {
		log.Printf("login: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	expiresIn := time.Hour
	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, expiresIn)
	if err != nil {
		log.Printf("login: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	refreshToken, err := session.MakeRefreshToken()
	if err != nil {
		log.Printf("login: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	cfg.db.CreateRefreshToken(context.Background(),
		database.CreateRefreshTokenParams{
			Token:     refreshToken,
			UserID:    user.ID,
			ExpiresAt: time.Now().AddDate(0, 0, 60),
		},
	)

	dat, err := json.Marshal(struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
	})

	if err != nil {
		log.Printf("create user: %v", err)
		internalError(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("refresh handler: %v", err)
		internalError(w)
		return
	}

	q, err := cfg.db.GetRefreshToken(context.Background(), refreshToken)
	if err != nil {
		log.Printf("refresh handler: %v", err)
		internalError(w)
		return
	}

	if q.RevokedAt.Valid {
		log.Printf("refresh token revoked: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	expiresIn := time.Hour
	token, err := auth.MakeJWT(q.UserID, cfg.jwtSecret, expiresIn)
	if err != nil {
		log.Printf("refresh handler: %v", err)
		internalError(w)
		return
	}

	dat, err := json.Marshal(struct {
		Token string `json:"token"`
	}{
		Token: token,
	})
	if err != nil {
		log.Printf("create user: %v", err)
		internalError(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("revoke handler: %v", err)
		internalError(w)
		return
	}

	err = cfg.db.RevokeRefreshToken(context.Background(), refreshToken)
	if err != nil {
		log.Printf("revoke handler: %v", err)
		internalError(w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
