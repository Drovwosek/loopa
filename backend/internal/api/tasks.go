package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"loopa/backend/internal/session"
)

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		writeError(w, http.StatusBadRequest, "task id required")
		return
	}

	sessionID := session.GetSessionID(r)
	var (
		status       string
		originalName string
		transcript   sql.NullString
		errorMsg     sql.NullString
		createdAt    time.Time
		completedAt  sql.NullTime
	)

	err := s.db.QueryRow(
		`SELECT t.status, f.original_name, t.transcript_text, t.error_message, t.created_at, t.completed_at
		 FROM transcription_tasks t
		 JOIN files f ON f.id = t.file_id
		 WHERE t.id = ? AND f.user_session_id = ?`,
		taskID, sessionID,
	).Scan(&status, &originalName, &transcript, &errorMsg, &createdAt, &completedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load task")
		return
	}

	resp := TaskResponse{
		ID:           taskID,
		Status:       status,
		OriginalName: originalName,
		CreatedAt:    createdAt.UTC().Format(time.RFC3339),
	}
	if transcript.Valid {
		resp.TranscriptText = &transcript.String
	}
	if errorMsg.Valid {
		resp.ErrorMessage = &errorMsg.String
	}
	if completedAt.Valid {
		value := completedAt.Time.UTC().Format(time.RFC3339)
		resp.CompletedAt = &value
	}

	writeJSON(w, http.StatusOK, resp)
}
