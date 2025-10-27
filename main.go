package main

import (
	"chirpy/internal/database"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	port    = "8080"
	fileDir = "."
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	jwtSecret      []byte
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("could not load .env: %v", err)
	}
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("could not open sql: %v", err)
	}
	dbQueries := database.New(db)
	mux := http.NewServeMux()
	srv := http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}
	b64jwtSecret := os.Getenv("JWT_SECRET")
	jwtSecret, err := base64.StdEncoding.DecodeString(b64jwtSecret)
	if err != nil {
		log.Fatalf("error decoding jwt secret: %v", err)
	}

	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		jwtSecret:      jwtSecret,
	}

	mux.Handle("/app/", cfg.middlewareMetricInc(http.StripPrefix("/app", http.FileServer(http.Dir(fileDir)))))
	mux.Handle("GET /api/healthz", http.HandlerFunc(healthz))
	mux.Handle("GET /api/chirps", http.HandlerFunc(cfg.handlerGetAllChirps))
	mux.Handle("POST /api/chirps", http.HandlerFunc(cfg.handlerCreateChirp))
	mux.Handle("GET /api/chirps/{id}/", http.HandlerFunc(cfg.handlerGetChirpByID))
	mux.Handle("POST /api/users", http.HandlerFunc(cfg.handlerCreateUser))
	mux.Handle("POST /api/login", http.HandlerFunc(cfg.handlerLogin))
	mux.Handle("GET /admin/metrics", http.HandlerFunc(cfg.metrics))
	mux.Handle("POST /admin/reset", http.HandlerFunc(cfg.reset))
	mux.Handle("POST /api/refresh", http.HandlerFunc(cfg.handlerRefresh))
	mux.Handle("POST /api/revoke", http.HandlerFunc(cfg.handlerRevoke))

	log.Printf("Staring server on port %s", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server failed")
	}
}

func internalError(w http.ResponseWriter) {
	respBody := struct {
		Error string `json:"error"`
	}{
		Error: "Something went wrong",
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("generic error: %v", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(dat)
}
