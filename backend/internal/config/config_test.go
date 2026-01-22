package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear all env vars that could affect the test
	os.Unsetenv("DB_DSN")
	os.Unsetenv("UPLOAD_DIR")
	os.Unsetenv("MAX_UPLOAD_BYTES")
	os.Unsetenv("YANDEX_SPEECHKIT_API_KEY")
	os.Unsetenv("YANDEX_STORAGE_ACCESS_KEY")
	os.Unsetenv("YANDEX_STORAGE_SECRET_KEY")
	os.Unsetenv("YANDEX_STORAGE_BUCKET")

	cfg := Load()

	assert.Equal(t, "root:root@tcp(mysql:3306)/loopa?parseTime=true", cfg.DBDSN)
	assert.Equal(t, "/data/uploads", cfg.UploadDir)
	assert.Equal(t, int64(1073741824), cfg.MaxUploadBytes) // 1GB
	assert.Equal(t, "", cfg.YandexSpeechKitAPIKey)
}

func TestLoad_FromEnv(t *testing.T) {
	os.Setenv("DB_DSN", "custom:custom@tcp(localhost:3306)/test")
	os.Setenv("UPLOAD_DIR", "/custom/uploads")
	os.Setenv("MAX_UPLOAD_BYTES", "512000")
	os.Setenv("YANDEX_SPEECHKIT_API_KEY", "test-api-key")
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("UPLOAD_DIR")
		os.Unsetenv("MAX_UPLOAD_BYTES")
		os.Unsetenv("YANDEX_SPEECHKIT_API_KEY")
	}()

	cfg := Load()

	assert.Equal(t, "custom:custom@tcp(localhost:3306)/test", cfg.DBDSN)
	assert.Equal(t, "/custom/uploads", cfg.UploadDir)
	assert.Equal(t, int64(512000), cfg.MaxUploadBytes)
	assert.Equal(t, "test-api-key", cfg.YandexSpeechKitAPIKey)
}

func TestLoad_ObjectStorageConfig(t *testing.T) {
	os.Setenv("YANDEX_STORAGE_ACCESS_KEY", "access-key")
	os.Setenv("YANDEX_STORAGE_SECRET_KEY", "secret-key")
	os.Setenv("YANDEX_STORAGE_BUCKET", "my-bucket")
	defer func() {
		os.Unsetenv("YANDEX_STORAGE_ACCESS_KEY")
		os.Unsetenv("YANDEX_STORAGE_SECRET_KEY")
		os.Unsetenv("YANDEX_STORAGE_BUCKET")
	}()

	cfg := Load()

	assert.Equal(t, "access-key", cfg.YandexStorageAccessKey)
	assert.Equal(t, "secret-key", cfg.YandexStorageSecretKey)
	assert.Equal(t, "my-bucket", cfg.YandexStorageBucket)
}

func TestHasObjectStorage_True(t *testing.T) {
	cfg := Config{
		YandexStorageAccessKey: "access",
		YandexStorageSecretKey: "secret",
		YandexStorageBucket:    "bucket",
	}

	assert.True(t, cfg.HasObjectStorage())
}

func TestHasObjectStorage_False_MissingAccessKey(t *testing.T) {
	cfg := Config{
		YandexStorageAccessKey: "",
		YandexStorageSecretKey: "secret",
		YandexStorageBucket:    "bucket",
	}

	assert.False(t, cfg.HasObjectStorage())
}

func TestHasObjectStorage_False_MissingSecretKey(t *testing.T) {
	cfg := Config{
		YandexStorageAccessKey: "access",
		YandexStorageSecretKey: "",
		YandexStorageBucket:    "bucket",
	}

	assert.False(t, cfg.HasObjectStorage())
}

func TestHasObjectStorage_False_MissingBucket(t *testing.T) {
	cfg := Config{
		YandexStorageAccessKey: "access",
		YandexStorageSecretKey: "secret",
		YandexStorageBucket:    "",
	}

	assert.False(t, cfg.HasObjectStorage())
}

func TestHasObjectStorage_False_AllEmpty(t *testing.T) {
	cfg := Config{}

	assert.False(t, cfg.HasObjectStorage())
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_VAR", "test-value")
	defer os.Unsetenv("TEST_VAR")

	assert.Equal(t, "test-value", getEnv("TEST_VAR", "default"))
	assert.Equal(t, "default", getEnv("NONEXISTENT_VAR", "default"))
}

func TestGetEnvInt64(t *testing.T) {
	os.Setenv("TEST_INT", "12345")
	defer os.Unsetenv("TEST_INT")

	assert.Equal(t, int64(12345), getEnvInt64("TEST_INT", 0))
	assert.Equal(t, int64(999), getEnvInt64("NONEXISTENT_INT", 999))
}

func TestGetEnvInt64_InvalidValue(t *testing.T) {
	os.Setenv("TEST_INVALID_INT", "not-a-number")
	defer os.Unsetenv("TEST_INVALID_INT")

	// Should return fallback for invalid integer
	assert.Equal(t, int64(100), getEnvInt64("TEST_INVALID_INT", 100))
}

func TestGetEnvInt64_EmptyValue(t *testing.T) {
	os.Setenv("TEST_EMPTY_INT", "")
	defer os.Unsetenv("TEST_EMPTY_INT")

	// Empty string should return fallback
	assert.Equal(t, int64(200), getEnvInt64("TEST_EMPTY_INT", 200))
}
