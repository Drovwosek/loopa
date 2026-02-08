package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBDSN                  string
	UploadDir              string
	MaxUploadBytes         int64
	TranscriptionProvider  string // "whisper" (default) или "speechkit"
	YandexSpeechKitAPIKey  string
	YandexFolderId         string
	// Yandex Object Storage для длинных аудио
	YandexStorageAccessKey string
	YandexStorageSecretKey string
	YandexStorageBucket    string
	// ML-сервис
	MLServiceURL string
}

func Load() Config {
	return Config{
		DBDSN:                  getEnv("DB_DSN", "root:root@tcp(mysql:3306)/loopa?parseTime=true"),
		UploadDir:              getEnv("UPLOAD_DIR", "/data/uploads"),
		MaxUploadBytes:         getEnvInt64("MAX_UPLOAD_BYTES", 1073741824),
		TranscriptionProvider:  getEnv("TRANSCRIPTION_PROVIDER", "whisper"),
		YandexSpeechKitAPIKey:  getEnv("YANDEX_SPEECHKIT_API_KEY", ""),
		YandexFolderId:         getEnv("YANDEX_FOLDER_ID", ""),
		YandexStorageAccessKey: getEnv("YANDEX_STORAGE_ACCESS_KEY", ""),
		YandexStorageSecretKey: getEnv("YANDEX_STORAGE_SECRET_KEY", ""),
		YandexStorageBucket:    getEnv("YANDEX_STORAGE_BUCKET", ""),
		MLServiceURL:           getEnv("ML_SERVICE_URL", "http://ml-service:8001"),
	}
}

// HasObjectStorage проверяет, настроен ли Object Storage для длинных аудио.
func (c *Config) HasObjectStorage() bool {
	return c.YandexStorageAccessKey != "" &&
		c.YandexStorageSecretKey != "" &&
		c.YandexStorageBucket != ""
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			return parsed
		}
	}
	return fallback
}
