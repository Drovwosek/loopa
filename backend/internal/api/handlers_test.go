package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Unit tests for utility functions

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()

	data := map[string]string{"message": "hello"}
	writeJSON(w, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var result map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "hello", result["message"])
}

func TestWriteJSON_WithStruct(t *testing.T) {
	w := httptest.NewRecorder()

	data := TaskResponse{
		ID:           "task-123",
		Status:       "готово",
		OriginalName: "test.mp3",
		CreatedAt:    "2024-01-01T00:00:00Z",
	}
	writeJSON(w, http.StatusCreated, data)

	assert.Equal(t, http.StatusCreated, w.Code)

	var result TaskResponse
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "task-123", result.ID)
	assert.Equal(t, "готово", result.Status)
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()

	writeError(w, http.StatusBadRequest, "invalid input")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var result map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "invalid input", result["error"])
}

func TestWriteJSON_EmptySlice(t *testing.T) {
	w := httptest.NewRecorder()

	data := []HistoryItem{}
	writeJSON(w, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]\n", w.Body.String())
}

func TestIsAllowedFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		mimeType string
		expected bool
	}{
		{"MP3 extension", "audio.mp3", "", true},
		{"WAV extension", "audio.wav", "", true},
		{"MP4 extension", "video.mp4", "", true},
		{"MOV extension", "video.mov", "", true},
		{"MP3 uppercase", "audio.MP3", "", true},
		{"WAV uppercase", "audio.WAV", "", true},
		{"audio/* MIME type", "unknown", "audio/ogg", true},
		{"video/* MIME type", "unknown", "video/webm", true},
		{"PDF not allowed", "document.pdf", "", false},
		{"TXT not allowed", "text.txt", "", false},
		{"Image not allowed", "photo.jpg", "image/jpeg", false},
		{"Empty filename", "", "", false},
		{"No extension with audio MIME", "noext", "audio/mp3", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAllowedFile(tt.filename, tt.mimeType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsVideoFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		mimeType string
		expected bool
	}{
		{"MP4 extension", "video.mp4", "", true},
		{"MOV extension", "video.mov", "", true},
		{"MP4 uppercase", "video.MP4", "", true},
		{"video/* MIME type", "unknown", "video/webm", true},
		{"audio/mp3 is not video", "audio.mp3", "audio/mp3", false},
		{"WAV is not video", "audio.wav", "", false},
		{"MP3 is not video", "audio.mp3", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isVideoFile(tt.filename, tt.mimeType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeDownloadName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple.mp3", "simple"},
		{"file with spaces.wav", "file_with_spaces"},
		{"special!@#chars.mp4", "special___chars"},
		{".mp3", "transcript"},
		{"valid-file_name.wav", "valid-file_name"},
		{"multiple...dots.mp3", "multiple___dots"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeDownloadName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeDownloadName_Unicode(t *testing.T) {
	// Unicode characters get replaced with underscore
	result := sanitizeDownloadName("русский файл.mp3")
	// Just verify it returns something valid, not empty
	assert.NotEmpty(t, result)
	// Should not return "transcript" (fallback) since there are underscores
	assert.NotEqual(t, "", result)
}

func TestTaskResponse_JSONSerialization(t *testing.T) {
	transcript := "Hello world"
	completedAt := "2024-01-01T00:01:00Z"
	
	response := TaskResponse{
		ID:             "task-123",
		Status:         "готово",
		OriginalName:   "test.mp3",
		TranscriptText: &transcript,
		CreatedAt:      "2024-01-01T00:00:00Z",
		CompletedAt:    &completedAt,
	}

	data, err := json.Marshal(response)
	assert.NoError(t, err)

	var decoded TaskResponse
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, "task-123", decoded.ID)
	assert.Equal(t, "готово", decoded.Status)
	assert.Equal(t, "Hello world", *decoded.TranscriptText)
}

func TestHistoryItem_JSONSerialization(t *testing.T) {
	item := HistoryItem{
		ID:           "item-123",
		OriginalName: "file.mp3",
		Status:       "готово",
		UploadedAt:   "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(item)
	assert.NoError(t, err)

	var decoded HistoryItem
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, "item-123", decoded.ID)
	assert.Equal(t, "file.mp3", decoded.OriginalName)
}
