package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"loopa/backend/internal/api"
	"loopa/backend/internal/config"
	"loopa/backend/internal/db"
	"loopa/backend/internal/storage"
)

func main() {
	cfg := config.Load()

	if err := storage.EnsureDir(cfg.UploadDir); err != nil {
		log.Fatalf("failed to create upload dir: %v", err)
	}

	conn, err := db.Open(cfg.DBDSN)
	if err != nil {
		log.Fatalf("db open failed: %v", err)
	}
	defer conn.Close()

	migrationsDir := getEnv("MIGRATIONS_DIR", "./migrations")
	if err := db.Migrate(conn, migrationsDir); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	server := api.NewServer(conn, cfg)
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      server.Router(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("API listening on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
