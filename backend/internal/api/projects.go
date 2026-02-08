package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"loopa/backend/internal/session"
)

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	sessionID := session.GetSessionID(r)
	projectID := uuid.New().String()
	now := time.Now().UTC()

	_, err := s.db.Exec(
		`INSERT INTO projects (id, name, description, status, user_session_id, created_at)
		 VALUES (?, ?, ?, 'active', ?, ?)`,
		projectID, req.Name, req.Description, sessionID, now,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create project")
		return
	}

	resp := ProjectResponse{
		ID:          projectID,
		Name:        req.Name,
		Description: req.Description,
		Status:      "active",
		CreatedAt:   now.Format(time.RFC3339),
		FileCount:   0,
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	sessionID := session.GetSessionID(r)

	rows, err := s.db.Query(
		`SELECT p.id, p.name, p.description, p.status, p.created_at,
		        (SELECT COUNT(*) FROM files f WHERE f.project_id = p.id) as file_count
		 FROM projects p
		 WHERE p.user_session_id = ?
		 ORDER BY p.created_at DESC`,
		sessionID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load projects")
		return
	}
	defer rows.Close()

	items := []ProjectResponse{}
	for rows.Next() {
		var item ProjectResponse
		var desc sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.Name, &desc, &item.Status, &createdAt, &item.FileCount); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to parse projects")
			return
		}
		if desc.Valid {
			item.Description = &desc.String
		}
		item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	sessionID := session.GetSessionID(r)

	var item ProjectResponse
	var desc sql.NullString
	var createdAt time.Time

	err := s.db.QueryRow(
		`SELECT p.id, p.name, p.description, p.status, p.created_at,
		        (SELECT COUNT(*) FROM files f WHERE f.project_id = p.id) as file_count
		 FROM projects p
		 WHERE p.id = ? AND p.user_session_id = ?`,
		projectID, sessionID,
	).Scan(&item.ID, &item.Name, &desc, &item.Status, &createdAt, &item.FileCount)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load project")
		return
	}

	if desc.Valid {
		item.Description = &desc.String
	}
	item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	writeJSON(w, http.StatusOK, item)
}

// handleListProjectFiles возвращает файлы проекта с их задачами.
func (s *Server) handleListProjectFiles(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	sessionID := session.GetSessionID(r)

	// Проверяем что проект принадлежит пользователю
	var exists int
	err := s.db.QueryRow(
		`SELECT 1 FROM projects WHERE id = ? AND user_session_id = ?`,
		projectID, sessionID,
	).Scan(&exists)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check project")
		return
	}

	rows, err := s.db.Query(
		`SELECT f.id, f.original_name, f.uploaded_at,
		        t.id, t.status
		 FROM files f
		 LEFT JOIN transcription_tasks t ON t.file_id = f.id
		 WHERE f.project_id = ?
		 ORDER BY f.uploaded_at DESC`,
		projectID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load files")
		return
	}
	defer rows.Close()

	type ProjectFileItem struct {
		FileID       string  `json:"fileId"`
		OriginalName string  `json:"originalName"`
		UploadedAt   string  `json:"uploadedAt"`
		TaskID       *string `json:"taskId,omitempty"`
		Status       *string `json:"status,omitempty"`
	}

	items := []ProjectFileItem{}
	for rows.Next() {
		var item ProjectFileItem
		var uploadedAt time.Time
		var taskID, status sql.NullString
		if err := rows.Scan(&item.FileID, &item.OriginalName, &uploadedAt, &taskID, &status); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to parse files")
			return
		}
		item.UploadedAt = uploadedAt.UTC().Format(time.RFC3339)
		if taskID.Valid {
			item.TaskID = &taskID.String
		}
		if status.Valid {
			item.Status = &status.String
		}
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	sessionID := session.GetSessionID(r)

	res, err := s.db.Exec(
		`DELETE FROM projects WHERE id = ? AND user_session_id = ?`,
		projectID, sessionID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete project")
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
