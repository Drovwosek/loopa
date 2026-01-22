package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureDir(t *testing.T) {
	tempDir := t.TempDir()
	newDir := filepath.Join(tempDir, "subdir", "nested")

	err := EnsureDir(newDir)
	assert.NoError(t, err)

	info, err := os.Stat(newDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestEnsureDir_ExistingDir(t *testing.T) {
	tempDir := t.TempDir()

	// Call twice, should not error
	err := EnsureDir(tempDir)
	assert.NoError(t, err)

	err = EnsureDir(tempDir)
	assert.NoError(t, err)
}

func TestSaveUploadedFile(t *testing.T) {
	tempDir := t.TempDir()
	content := []byte("test file content")
	reader := bytes.NewReader(content)

	path, fileID, size, err := SaveUploadedFile(tempDir, "test.mp3", reader)

	require.NoError(t, err)
	assert.NotEmpty(t, fileID)
	assert.Equal(t, int64(len(content)), size)
	assert.True(t, strings.HasSuffix(path, "test.mp3"))

	// Verify file was created
	savedContent, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, content, savedContent)
}

func TestSaveUploadedFile_WithSpacesInName(t *testing.T) {
	tempDir := t.TempDir()
	content := []byte("content")
	reader := bytes.NewReader(content)

	path, _, _, err := SaveUploadedFile(tempDir, "file with spaces.mp3", reader)

	require.NoError(t, err)
	assert.Contains(t, path, "file_with_spaces.mp3")
}

func TestSaveUploadedFile_LargeFile(t *testing.T) {
	tempDir := t.TempDir()
	// Create 1MB of data
	content := make([]byte, 1024*1024)
	for i := range content {
		content[i] = byte(i % 256)
	}
	reader := bytes.NewReader(content)

	path, _, size, err := SaveUploadedFile(tempDir, "large.wav", reader)

	require.NoError(t, err)
	assert.Equal(t, int64(len(content)), size)

	savedContent, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, content, savedContent)
}

func TestSaveUploadedFile_EmptyContent(t *testing.T) {
	tempDir := t.TempDir()
	reader := bytes.NewReader([]byte{})

	path, _, size, err := SaveUploadedFile(tempDir, "empty.mp3", reader)

	require.NoError(t, err)
	assert.Equal(t, int64(0), size)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, int64(0), info.Size())
}

func TestSaveUploadedFile_CreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	newDir := filepath.Join(tempDir, "uploads", "new")
	content := []byte("test")
	reader := bytes.NewReader(content)

	path, _, _, err := SaveUploadedFile(newDir, "test.mp3", reader)

	require.NoError(t, err)
	assert.FileExists(t, path)
}

func TestSaveUploadedFile_UniqueIDs(t *testing.T) {
	tempDir := t.TempDir()

	var ids []string
	for i := 0; i < 10; i++ {
		reader := bytes.NewReader([]byte("content"))
		_, fileID, _, err := SaveUploadedFile(tempDir, "test.mp3", reader)
		require.NoError(t, err)
		ids = append(ids, fileID)
	}

	// All IDs should be unique
	idSet := make(map[string]bool)
	for _, id := range ids {
		assert.False(t, idSet[id], "Duplicate ID found")
		idSet[id] = true
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple.mp3", "simple.mp3"},
		{"file with spaces.wav", "file_with_spaces.wav"},
		{"file!@#$%.mp4", "file.mp4"},
		{"../../../etc/passwd", "passwd"},
		{"/absolute/path/file.mp3", "file.mp3"},
		{"valid-file_name.wav", "valid-file_name.wav"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeFilename_EmptyResult(t *testing.T) {
	// When sanitization results in empty string, should return "file"
	result := sanitizeFilename("!@#$%")
	assert.Equal(t, "file", result)
}

func TestSanitizeFilename_Unicode(t *testing.T) {
	// Unicode characters get stripped
	result := sanitizeFilename("файл.mp3")
	// Result depends on implementation - just verify it's not empty and is valid
	assert.NotEmpty(t, result)
}

func TestSaveUploadedFile_InvalidDirectory(t *testing.T) {
	// Skip on Windows as path handling is different
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	// Try to save to a directory that doesn't exist and can't be created
	reader := bytes.NewReader([]byte("content"))
	_, _, _, err := SaveUploadedFile("/root/nonexistent/path/that/cannot/exist", "test.mp3", reader)

	assert.Error(t, err)
}

func TestSaveUploadedFile_PathTraversal(t *testing.T) {
	tempDir := t.TempDir()
	content := []byte("malicious content")
	reader := bytes.NewReader(content)

	// Try path traversal attack
	path, _, _, err := SaveUploadedFile(tempDir, "../../../etc/passwd", reader)

	require.NoError(t, err)
	// File should be saved in tempDir, not in /etc
	assert.True(t, strings.HasPrefix(path, tempDir))
	assert.Contains(t, path, "passwd")
}
