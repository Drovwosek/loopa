package api

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"loopa/backend/internal/session"
)

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		writeError(w, http.StatusBadRequest, "task id required")
		return
	}
	sessionID := session.GetSessionID(r)

	tx, err := s.db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start transaction")
		return
	}
	defer tx.Rollback()

	var storagePath string
	err = tx.QueryRow(
		`SELECT f.storage_path
		 FROM transcription_tasks t
		 JOIN files f ON f.id = t.file_id
		 WHERE t.id = ? AND f.user_session_id = ?`,
		taskID, sessionID,
	).Scan(&storagePath)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load task")
		return
	}

	if _, err := tx.Exec("DELETE FROM transcription_tasks WHERE id = ?", taskID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete task")
		return
	}

	if _, err := tx.Exec(
		`DELETE f FROM files f
		 LEFT JOIN transcription_tasks t ON t.file_id = f.id
		 WHERE f.storage_path = ? AND t.id IS NULL`,
		storagePath,
	); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete file metadata")
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to commit delete")
		return
	}

	_ = os.Remove(storagePath)
	w.WriteHeader(http.StatusNoContent)
}
