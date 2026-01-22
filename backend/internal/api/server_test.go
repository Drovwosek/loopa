package api

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"loopa/backend/internal/config"
)

func setupTestServer(t *testing.T) (*Server, sqlmock.Sqlmock, *sql.DB) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	cfg := config.Config{
		UploadDir:      t.TempDir(),
		MaxUploadBytes: 1024 * 1024,
	}
	server := NewServer(db, cfg)
	return server, mock, db
}

func TestNewServer(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	cfg := config.Config{
		UploadDir:      "/tmp/test",
		MaxUploadBytes: 1024,
	}

	server := NewServer(db, cfg)

	assert.NotNil(t, server)
	assert.Equal(t, db, server.db)
	assert.Equal(t, cfg, server.config)
}

func TestRouter(t *testing.T) {
	server, _, db := setupTestServer(t)
	defer db.Close()

	router := server.Router()
	assert.NotNil(t, router)

	// Test CORS middleware is applied
	req := httptest.NewRequest(http.MethodOptions, "/api/uploads", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// CORS should return proper headers
	assert.Equal(t, "http://localhost:5173", w.Header().Get("Access-Control-Allow-Origin"))
}
