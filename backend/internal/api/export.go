package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"

	"loopa/backend/internal/exporter"
	"loopa/backend/internal/session"
)

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	format := strings.ToLower(r.URL.Query().Get("format"))
	if format != "txt" && format != "docx" {
		writeError(w, http.StatusBadRequest, "invalid format")
		return
	}

	sessionID := session.GetSessionID(r)
	var (
		status       string
		originalName string
		transcript   sql.NullString
	)
	err := s.db.QueryRow(
		`SELECT t.status, f.original_name, t.transcript_text
		 FROM transcription_tasks t
		 JOIN files f ON f.id = t.file_id
		 WHERE t.id = ? AND f.user_session_id = ?`,
		taskID, sessionID,
	).Scan(&status, &originalName, &transcript)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load task")
		return
	}
	if status != "готово" || !transcript.Valid {
		writeError(w, http.StatusConflict, "transcript not ready")
		return
	}

	filename := sanitizeDownloadName(originalName) + "." + format
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	if format == "txt" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(transcript.String))
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	if err := exporter.WriteDocx(w, transcript.String); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate docx")
		return
	}
}

func sanitizeDownloadName(name string) string {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	base = strings.ReplaceAll(base, " ", "_")
	base = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			return r
		}
		return '_'
	}, base)
	if base == "" {
		return "transcript"
	}
	return base
}
