package session

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSessionID_FromContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), sessionKey, "test-session-123")
	req = req.WithContext(ctx)

	sessionID := GetSessionID(req)

	assert.Equal(t, "test-session-123", sessionID)
}

func TestGetSessionID_NoSession(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	sessionID := GetSessionID(req)

	assert.Equal(t, "", sessionID)
}

func TestGetSessionID_EmptySession(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), sessionKey, "")
	req = req.WithContext(ctx)

	sessionID := GetSessionID(req)

	assert.Equal(t, "", sessionID)
}

func TestGetSessionID_WrongType(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), sessionKey, 12345) // wrong type
	req = req.WithContext(ctx)

	sessionID := GetSessionID(req)

	assert.Equal(t, "", sessionID)
}

func TestGetOrCreateSessionID_NewSession(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	sessionID := getOrCreateSessionID(w, req)

	assert.NotEmpty(t, sessionID)

	// Check cookie was set
	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == CookieName {
			sessionCookie = c
			break
		}
	}

	require.NotNil(t, sessionCookie)
	assert.Equal(t, sessionID, sessionCookie.Value)
	assert.True(t, sessionCookie.HttpOnly)
	assert.Equal(t, "/", sessionCookie.Path)
}

func TestGetOrCreateSessionID_ExistingSession(t *testing.T) {
	existingID := "existing-session-456"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  CookieName,
		Value: existingID,
	})
	w := httptest.NewRecorder()

	sessionID := getOrCreateSessionID(w, req)

	assert.Equal(t, existingID, sessionID)

	// No new cookie should be set
	cookies := w.Result().Cookies()
	assert.Len(t, cookies, 0)
}

func TestGetOrCreateSessionID_EmptyCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  CookieName,
		Value: "",
	})
	w := httptest.NewRecorder()

	sessionID := getOrCreateSessionID(w, req)

	// Should create new session when cookie value is empty
	assert.NotEmpty(t, sessionID)
}

func TestMiddleware(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Expect the upsert query
	mock.ExpectExec("INSERT INTO user_sessions").
		WillReturnResult(sqlmock.NewResult(1, 1))

	var capturedSessionID string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedSessionID = GetSessionID(r)
		w.WriteHeader(http.StatusOK)
	})

	middleware := Middleware(db)
	handler := middleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, capturedSessionID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMiddleware_ExistingSession(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	existingID := "existing-session-789"

	// Expect the upsert query with existing session
	mock.ExpectExec("INSERT INTO user_sessions").
		WillReturnResult(sqlmock.NewResult(1, 1))

	var capturedSessionID string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedSessionID = GetSessionID(r)
		w.WriteHeader(http.StatusOK)
	})

	middleware := Middleware(db)
	handler := middleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  CookieName,
		Value: existingID,
	})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, existingID, capturedSessionID)
}

func TestCookieName(t *testing.T) {
	assert.Equal(t, "session_id", CookieName)
}

func TestUpsertSession(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sessionID := "test-session-upsert"

	mock.ExpectExec("INSERT INTO user_sessions").
		WithArgs(sessionID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = upsertSession(db, sessionID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpsertSession_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("INSERT INTO user_sessions").
		WillReturnError(assert.AnError)

	err = upsertSession(db, "test-session")

	assert.Error(t, err)
}
