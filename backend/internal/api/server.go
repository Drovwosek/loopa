package api

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"

	"loopa/backend/internal/config"
	"loopa/backend/internal/session"
)

type Server struct {
	db     *sql.DB
	config config.Config
}

func NewServer(db *sql.DB, cfg config.Config) *Server {
	return &Server{db: db, config: cfg}
}

func (s *Server) Router() http.Handler {
	router := chi.NewRouter()
	router.Use(session.Middleware(s.db))

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})

	router.Route("/api", func(r chi.Router) {
		r.Post("/uploads", s.handleUpload)
		r.Get("/tasks/{id}", s.handleGetTask)
		r.Get("/tasks/{id}/export", s.handleExport)
		r.Get("/tasks/{id}/segments", s.handleGetSegments)
		r.Put("/tasks/{id}/segments/{segId}", s.handleUpdateSegment)
		r.Put("/tasks/{id}/speakers/{speakerId}", s.handleUpdateSpeaker)
		r.Get("/tasks/{id}/audio", s.handleGetAudio)
		r.Get("/history", s.handleHistory)
		r.Delete("/tasks/{id}", s.handleDeleteTask)

		r.Post("/projects", s.handleCreateProject)
		r.Get("/projects", s.handleListProjects)
		r.Get("/projects/{id}", s.handleGetProject)
		r.Get("/projects/{id}/files", s.handleListProjectFiles)
		r.Delete("/projects/{id}", s.handleDeleteProject)
	})

	return corsHandler.Handler(router)
}
