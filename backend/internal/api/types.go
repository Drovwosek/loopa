package api

type TaskResponse struct {
	ID             string            `json:"id"`
	Status         string            `json:"status"`
	OriginalName   string            `json:"originalName"`
	TranscriptText *string           `json:"transcriptText,omitempty"`
	ErrorMessage   *string           `json:"errorMessage,omitempty"`
	CreatedAt      string            `json:"createdAt"`
	CompletedAt    *string           `json:"completedAt,omitempty"`
	Segments       []SegmentResponse `json:"segments,omitempty"`
	NumSpeakers    int               `json:"numSpeakers,omitempty"`
}

type HistoryItem struct {
	ID           string `json:"id"`
	OriginalName string `json:"originalName"`
	Status       string `json:"status"`
	UploadedAt   string `json:"uploadedAt"`
}

type ProjectResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"createdAt"`
	FileCount   int     `json:"fileCount"`
}

type CreateProjectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type SegmentResponse struct {
	ID          string  `json:"id"`
	SpeakerID   *string `json:"speakerId,omitempty"`
	SpeakerName *string `json:"speakerName,omitempty"`
	StartTime   int     `json:"startTime"`
	EndTime     int     `json:"endTime"`
	Text        string  `json:"text"`
	HasFillers  bool    `json:"hasFillers"`
	IsCorrected bool    `json:"isCorrected"`
}

type UpdateSegmentRequest struct {
	Text string `json:"text"`
}

type UpdateSpeakerRequest struct {
	Name string `json:"name"`
}
