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
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})

	router.Route("/api", func(r chi.Router) {
		r.Post("/uploads", s.handleUpload)
		r.Get("/tasks/{id}", s.handleGetTask)
		r.Get("/tasks/{id}/export", s.handleExport)
		r.Get("/history", s.handleHistory)
		r.Delete("/tasks/{id}", s.handleDeleteTask)
	})

	return corsHandler.Handler(router)
}
