package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"loopa/backend/internal/config"
	"loopa/backend/internal/db"
	"loopa/backend/internal/worker"
)

func main() {
	cfg := config.Load()

	if cfg.YandexSpeechKitAPIKey == "" {
		log.Fatal("YANDEX_SPEECHKIT_API_KEY is required")
	}

	if cfg.YandexFolderId == "" {
		log.Fatal("YANDEX_FOLDER_ID is required")
	}

	// Отладка: показываем какие значения используются
	apiKeyPreview := cfg.YandexSpeechKitAPIKey
	if len(apiKeyPreview) > 10 {
		apiKeyPreview = apiKeyPreview[:5] + "..." + apiKeyPreview[len(apiKeyPreview)-5:]
	}
	log.Printf("Using API Key: %s (length: %d)", apiKeyPreview, len(cfg.YandexSpeechKitAPIKey))
	log.Printf("Using Folder ID: %s", cfg.YandexFolderId)

	conn, err := db.Open(cfg.DBDSN)
	if err != nil {
		log.Fatalf("db open failed: %v", err)
	}
	defer conn.Close()

	migrationsDir := getEnv("MIGRATIONS_DIR", "./migrations")
	if err := db.Migrate(conn, migrationsDir); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	w := worker.New(conn, cfg.YandexSpeechKitAPIKey, cfg.YandexFolderId, cfg.UploadDir)

	stop := make(chan struct{})
	go w.Run(stop)

	log.Println("Worker started with Yandex SpeechKit (chunked mode for long audio)")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	close(stop)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
