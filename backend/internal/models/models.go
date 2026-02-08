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
	ProjectID     *string
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
	ProcessingTime *int
	SpeakerData    *string
	CreatedAt      time.Time
	StartedAt      *time.Time
	CompletedAt    *time.Time
}

type Project struct {
	ID            string
	Name          string
	Description   *string
	Status        string
	UserSessionID string
	CreatedAt     time.Time
	UpdatedAt     *time.Time
}

type TranscriptionSegment struct {
	ID          string
	TaskID      string
	SpeakerID   *string
	SpeakerName *string
	StartTime   int
	EndTime     int
	Text        string
	HasFillers  bool
	IsCorrected bool
	CreatedAt   time.Time
}
