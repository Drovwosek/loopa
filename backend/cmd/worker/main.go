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

	// SpeechKit ключи обязательны только при provider=speechkit
	if cfg.TranscriptionProvider == "speechkit" {
		if cfg.YandexSpeechKitAPIKey == "" {
			log.Fatal("YANDEX_SPEECHKIT_API_KEY is required for speechkit provider")
		}
		if cfg.YandexFolderId == "" {
			log.Fatal("YANDEX_FOLDER_ID is required for speechkit provider")
		}
		apiKeyPreview := cfg.YandexSpeechKitAPIKey
		if len(apiKeyPreview) > 10 {
			apiKeyPreview = apiKeyPreview[:5] + "..." + apiKeyPreview[len(apiKeyPreview)-5:]
		}
		log.Printf("Using API Key: %s (length: %d)", apiKeyPreview, len(cfg.YandexSpeechKitAPIKey))
		log.Printf("Using Folder ID: %s", cfg.YandexFolderId)
	}

	log.Printf("Transcription provider: %s", cfg.TranscriptionProvider)

	conn, err := db.Open(cfg.DBDSN)
	if err != nil {
		log.Fatalf("db open failed: %v", err)
	}
	defer conn.Close()

	migrationsDir := getEnv("MIGRATIONS_DIR", "./migrations")
	if err := db.Migrate(conn, migrationsDir); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	var s3cfg *worker.S3Config
	if cfg.HasObjectStorage() {
		s3cfg = &worker.S3Config{
			AccessKey: cfg.YandexStorageAccessKey,
			SecretKey: cfg.YandexStorageSecretKey,
			Bucket:    cfg.YandexStorageBucket,
		}
	}

	w := worker.New(conn, cfg.TranscriptionProvider, cfg.YandexSpeechKitAPIKey, cfg.YandexFolderId, cfg.UploadDir, cfg.MLServiceURL, s3cfg)

	stop := make(chan struct{})
	go w.Run(stop)

	if cfg.TranscriptionProvider == "whisper" {
		log.Println("Worker started with Faster-Whisper (ML-сервис)")
	} else if s3cfg != nil {
		log.Println("Worker started with Yandex SpeechKit (async mode for long audio)")
	} else {
		log.Println("Worker started with Yandex SpeechKit (chunked mode for long audio)")
	}

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
