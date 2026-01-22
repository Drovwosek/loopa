package models

import "time"

type File struct {
	ID            string
	OriginalName  string
	StoragePath   string
	FileSize      int64
	MimeType      string
	UploadedAt    time.Time
	UserSessionID string
}

type Task struct {
	ID             string
	FileID         string
	Status         string
	Provider       string
	Language       *string
	TranscriptText *string
	RawResponse    *string
	ErrorMessage   *string
	CreatedAt      time.Time
	StartedAt      *time.Time
	CompletedAt    *time.Time
}
