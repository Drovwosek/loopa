package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFile_Structure(t *testing.T) {
	now := time.Now()
	file := File{
		ID:            "file-123",
		OriginalName:  "recording.mp3",
		StoragePath:   "/uploads/abc123_recording.mp3",
		FileSize:      1024000,
		MimeType:      "audio/mpeg",
		UploadedAt:    now,
		UserSessionID: "session-456",
	}

	assert.Equal(t, "file-123", file.ID)
	assert.Equal(t, "recording.mp3", file.OriginalName)
	assert.Equal(t, "/uploads/abc123_recording.mp3", file.StoragePath)
	assert.Equal(t, int64(1024000), file.FileSize)
	assert.Equal(t, "audio/mpeg", file.MimeType)
	assert.Equal(t, now, file.UploadedAt)
	assert.Equal(t, "session-456", file.UserSessionID)
}

func TestTask_Structure(t *testing.T) {
	now := time.Now()
	startedAt := now.Add(time.Minute)
	completedAt := now.Add(2 * time.Minute)
	lang := "ru-RU"
	transcript := "Hello world"
	rawResp := `{"result": "Hello world"}`

	task := Task{
		ID:             "task-123",
		FileID:         "file-456",
		Status:         "готово",
		Provider:       "yandex_speechkit",
		Language:       &lang,
		TranscriptText: &transcript,
		RawResponse:    &rawResp,
		ErrorMessage:   nil,
		CreatedAt:      now,
		StartedAt:      &startedAt,
		CompletedAt:    &completedAt,
	}

	assert.Equal(t, "task-123", task.ID)
	assert.Equal(t, "file-456", task.FileID)
	assert.Equal(t, "готово", task.Status)
	assert.Equal(t, "yandex_speechkit", task.Provider)
	assert.Equal(t, "ru-RU", *task.Language)
	assert.Equal(t, "Hello world", *task.TranscriptText)
	assert.Equal(t, `{"result": "Hello world"}`, *task.RawResponse)
	assert.Nil(t, task.ErrorMessage)
	assert.Equal(t, now, task.CreatedAt)
	assert.Equal(t, startedAt, *task.StartedAt)
	assert.Equal(t, completedAt, *task.CompletedAt)
}

func TestTask_WithError(t *testing.T) {
	now := time.Now()
	errMsg := "Speech recognition failed"

	task := Task{
		ID:           "task-error",
		FileID:       "file-789",
		Status:       "ошибка",
		Provider:     "yandex_speechkit",
		ErrorMessage: &errMsg,
		CreatedAt:    now,
	}

	assert.Equal(t, "ошибка", task.Status)
	require.NotNil(t, task.ErrorMessage)
	assert.Equal(t, "Speech recognition failed", *task.ErrorMessage)
	assert.Nil(t, task.TranscriptText)
}

func TestTask_Pending(t *testing.T) {
	now := time.Now()

	task := Task{
		ID:        "task-pending",
		FileID:    "file-111",
		Status:    "ожидает",
		Provider:  "mock",
		CreatedAt: now,
	}

	assert.Equal(t, "ожидает", task.Status)
	assert.Nil(t, task.Language)
	assert.Nil(t, task.TranscriptText)
	assert.Nil(t, task.RawResponse)
	assert.Nil(t, task.ErrorMessage)
	assert.Nil(t, task.StartedAt)
	assert.Nil(t, task.CompletedAt)
}

func TestTask_InProgress(t *testing.T) {
	now := time.Now()
	startedAt := now.Add(time.Minute)

	task := Task{
		ID:        "task-processing",
		FileID:    "file-222",
		Status:    "в процессе",
		Provider:  "yandex_speechkit",
		CreatedAt: now,
		StartedAt: &startedAt,
	}

	assert.Equal(t, "в процессе", task.Status)
	assert.NotNil(t, task.StartedAt)
	assert.Nil(t, task.CompletedAt)
}
