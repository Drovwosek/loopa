package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"loopa/backend/internal/session"
)

func (s *Server) handleGetSegments(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	sessionID := session.GetSessionID(r)

	// Проверяем, что задача принадлежит пользователю
	var exists int
	err := s.db.QueryRow(
		`SELECT 1 FROM transcription_tasks t
		 JOIN files f ON f.id = t.file_id
		 WHERE t.id = ? AND f.user_session_id = ?`,
		taskID, sessionID,
	).Scan(&exists)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to verify task")
		return
	}

	rows, err := s.db.Query(
		`SELECT id, speaker_id, speaker_name, start_time, end_time, text, has_fillers, is_corrected
		 FROM transcription_segments
		 WHERE task_id = ?
		 ORDER BY start_time`,
		taskID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load segments")
		return
	}
	defer rows.Close()

	segments := []SegmentResponse{}
	for rows.Next() {
		var seg SegmentResponse
		var speakerID, speakerName sql.NullString
		if err := rows.Scan(
			&seg.ID, &speakerID, &speakerName,
			&seg.StartTime, &seg.EndTime, &seg.Text,
			&seg.HasFillers, &seg.IsCorrected,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to parse segments")
			return
		}
		if speakerID.Valid {
			seg.SpeakerID = &speakerID.String
		}
		if speakerName.Valid {
			seg.SpeakerName = &speakerName.String
		}
		segments = append(segments, seg)
	}

	writeJSON(w, http.StatusOK, segments)
}

func (s *Server) handleUpdateSegment(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	segmentID := chi.URLParam(r, "segId")
	sessionID := session.GetSessionID(r)

	// Проверяем владельца
	var exists int
	err := s.db.QueryRow(
		`SELECT 1 FROM transcription_tasks t
		 JOIN files f ON f.id = t.file_id
		 WHERE t.id = ? AND f.user_session_id = ?`,
		taskID, sessionID,
	).Scan(&exists)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to verify task")
		return
	}

	var req UpdateSegmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	res, err := s.db.Exec(
		`UPDATE transcription_segments
		 SET text = ?, is_corrected = 1
		 WHERE id = ? AND task_id = ?`,
		req.Text, segmentID, taskID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update segment")
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		writeError(w, http.StatusNotFound, "segment not found")
		return
	}

	// Обновляем общий текст транскрипции
	s.rebuildTranscriptText(taskID)

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *Server) handleUpdateSpeaker(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	speakerID := chi.URLParam(r, "speakerId")
	sessionID := session.GetSessionID(r)

	// Проверяем владельца
	var exists int
	err := s.db.QueryRow(
		`SELECT 1 FROM transcription_tasks t
		 JOIN files f ON f.id = t.file_id
		 WHERE t.id = ? AND f.user_session_id = ?`,
		taskID, sessionID,
	).Scan(&exists)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to verify task")
		return
	}

	var req UpdateSpeakerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	_, err = s.db.Exec(
		`UPDATE transcription_segments
		 SET speaker_name = ?
		 WHERE task_id = ? AND speaker_id = ?`,
		req.Name, taskID, speakerID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update speaker")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *Server) handleGetAudio(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	sessionID := session.GetSessionID(r)

	var storagePath string
	err := s.db.QueryRow(
		`SELECT f.storage_path FROM transcription_tasks t
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

	w.Header().Set("Content-Type", "audio/ogg")
	w.Header().Set("Accept-Ranges", "bytes")
	http.ServeFile(w, r, storagePath)
}

// rebuildTranscriptText пересобирает полный текст транскрипции из сегментов.
func (s *Server) rebuildTranscriptText(taskID string) {
	rows, err := s.db.Query(
		`SELECT text FROM transcription_segments
		 WHERE task_id = ?
		 ORDER BY start_time`,
		taskID,
	)
	if err != nil {
		return
	}
	defer rows.Close()

	fullText := ""
	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			return
		}
		if fullText != "" {
			fullText += " "
		}
		fullText += text
	}

	s.db.Exec(
		`UPDATE transcription_tasks SET transcript_text = ? WHERE id = ?`,
		fullText, taskID,
	)

	// Обновляем время
	s.db.Exec(
		`UPDATE transcription_tasks SET completed_at = ? WHERE id = ?`,
		time.Now().UTC(), taskID,
	)
}
