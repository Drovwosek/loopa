package api

type TaskResponse struct {
	ID             string  `json:"id"`
	Status         string  `json:"status"`
	OriginalName   string  `json:"originalName"`
	TranscriptText *string `json:"transcriptText,omitempty"`
	ErrorMessage   *string `json:"errorMessage,omitempty"`
	CreatedAt      string  `json:"createdAt"`
	CompletedAt    *string `json:"completedAt,omitempty"`
}

type HistoryItem struct {
	ID           string `json:"id"`
	OriginalName string `json:"originalName"`
	Status       string `json:"status"`
	UploadedAt   string `json:"uploadedAt"`
}
