package api

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"loopa/backend/internal/media"
	"loopa/backend/internal/session"
	"loopa/backend/internal/storage"
)

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, s.config.MaxUploadBytes)

	reader, err := r.MultipartReader()
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart data")
		return
	}

	var (
		originalName string
		mimeType     string
		storagePath  string
		fileID       string
		fileSize     int64
		projectID    string
	)

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid multipart stream")
			return
		}

		// Читаем project_id из формы
		if part.FormName() == "projectId" {
			data, _ := io.ReadAll(part)
			_ = part.Close()
			projectID = strings.TrimSpace(string(data))
			continue
		}

		if part.FormName() != "file" || part.FileName() == "" {
			_ = part.Close()
			continue
		}

		originalName = part.FileName()
		mimeType = part.Header.Get("Content-Type")
		if mimeType == "" {
			mimeType = mime.TypeByExtension(strings.ToLower(filepath.Ext(originalName)))
		}
		if !isAllowedFile(originalName, mimeType) {
			_ = part.Close()
			writeError(w, http.StatusBadRequest, "unsupported file type")
			return
		}
		path, id, size, err := storage.SaveUploadedFile(s.config.UploadDir, originalName, part)
		_ = part.Close()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save upload")
			return
		}
		storagePath = path
		fileID = id
		fileSize = size
		break
	}

	if storagePath == "" {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}

	if isVideoFile(originalName, mimeType) {
		originalPath := storagePath
		audioPath, err := media.ExtractAudio(originalPath, s.config.UploadDir)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to extract audio")
			return
		}
		_ = os.Remove(originalPath)
		storagePath = audioPath
	}

	sessionID := session.GetSessionID(r)
	if sessionID == "" {
		writeError(w, http.StatusInternalServerError, "session not initialized")
		return
	}

	now := time.Now().UTC()

	// project_id — NULL если не указан
	var projectIDParam interface{}
	if projectID != "" {
		projectIDParam = projectID
	}

	_, err = s.db.Exec(
		`INSERT INTO files (id, original_name, storage_path, file_size, mime_type, uploaded_at, user_session_id, project_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		fileID, originalName, storagePath, fileSize, mimeType, now, sessionID, projectIDParam,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store file metadata")
		return
	}

	taskID := uuid.New().String()
	_, err = s.db.Exec(
		`INSERT INTO transcription_tasks (id, file_id, status, provider, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		taskID, fileID, "ожидает", "mock", now,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create task")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"taskId": taskID})
}

func isVideoFile(name, mimeType string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == ".mp4" || ext == ".mov" {
		return true
	}
	return strings.HasPrefix(mimeType, "video/")
}

func isAllowedFile(name, mimeType string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".mp3", ".wav", ".mp4", ".mov":
		return true
	}
	return strings.HasPrefix(mimeType, "audio/") || strings.HasPrefix(mimeType, "video/")
}
