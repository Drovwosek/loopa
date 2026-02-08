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

type exportSegment struct {
	SpeakerName sql.NullString
	StartTime   int
	EndTime     int
	Text        string
}

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

	// Загружаем сегменты (если есть)
	segments := s.loadExportSegments(taskID)

	// Формируем текст с учётом спикеров и таймкодов
	exportText := buildExportText(segments, transcript.String)

	filename := sanitizeDownloadName(originalName) + "." + format
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	if format == "txt" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(exportText))
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	if err := exporter.WriteDocx(w, exportText); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate docx")
		return
	}
}

func (s *Server) loadExportSegments(taskID string) []exportSegment {
	rows, err := s.db.Query(
		`SELECT speaker_name, start_time, end_time, text
		 FROM transcription_segments
		 WHERE task_id = ?
		 ORDER BY start_time`,
		taskID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var segments []exportSegment
	for rows.Next() {
		var seg exportSegment
		if err := rows.Scan(&seg.SpeakerName, &seg.StartTime, &seg.EndTime, &seg.Text); err != nil {
			return nil
		}
		segments = append(segments, seg)
	}
	return segments
}

func buildExportText(segments []exportSegment, fallbackText string) string {
	if len(segments) == 0 {
		return fallbackText
	}

	var sb strings.Builder
	for _, seg := range segments {
		speaker := "Спикер"
		if seg.SpeakerName.Valid && seg.SpeakerName.String != "" {
			speaker = seg.SpeakerName.String
		}

		timeLabel := formatTimeRange(seg.StartTime, seg.EndTime)
		sb.WriteString(fmt.Sprintf("[%s] %s:\n%s\n\n", timeLabel, speaker, seg.Text))
	}
	return strings.TrimSpace(sb.String())
}

func formatTimeRange(startMs, endMs int) string {
	return fmt.Sprintf("%s — %s", formatMs(startMs), formatMs(endMs))
}

func formatMs(ms int) string {
	totalSec := ms / 1000
	m := totalSec / 60
	s := totalSec % 60
	return fmt.Sprintf("%d:%02d", m, s)
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
