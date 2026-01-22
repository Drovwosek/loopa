package api

import (
	"net/http"
	"time"

	"loopa/backend/internal/session"
)

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	sessionID := session.GetSessionID(r)
	rows, err := s.db.Query(
		`SELECT t.id, f.original_name, t.status, f.uploaded_at
		 FROM transcription_tasks t
		 JOIN files f ON f.id = t.file_id
		 WHERE f.user_session_id = ?
		 ORDER BY f.uploaded_at DESC
		 LIMIT 10`,
		sessionID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load history")
		return
	}
	defer rows.Close()

	items := []HistoryItem{}
	for rows.Next() {
		var item HistoryItem
		var uploaded time.Time
		if err := rows.Scan(&item.ID, &item.OriginalName, &item.Status, &uploaded); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to parse history")
			return
		}
		item.UploadedAt = uploaded.UTC().Format(time.RFC3339)
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, items)
}
