package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var invalidChars = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func SaveUploadedFile(baseDir, originalName string, src io.Reader) (string, string, int64, error) {
	if err := EnsureDir(baseDir); err != nil {
		return "", "", 0, err
	}
	cleanName := sanitizeFilename(originalName)
	fileID := uuid.New().String()
	filename := fmt.Sprintf("%s_%s", fileID, cleanName)
	fullPath := filepath.Join(baseDir, filename)

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", "", 0, err
	}
	defer dst.Close()

	written, err := io.Copy(dst, src)
	if err != nil {
		return "", "", 0, err
	}
	return fullPath, fileID, written, nil
}

func sanitizeFilename(name string) string {
	base := filepath.Base(name)
	base = strings.ReplaceAll(base, " ", "_")
	base = invalidChars.ReplaceAllString(base, "")
	if base == "" {
		return "file"
	}
	return base
}
