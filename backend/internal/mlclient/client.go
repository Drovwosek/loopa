package mlclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type DiarizationSegment struct {
	Speaker  string  `json:"speaker"`
	Start    float64 `json:"start"`
	End      float64 `json:"end"`
	Duration float64 `json:"duration"`
}

type DiarizationResponse struct {
	Segments    []DiarizationSegment `json:"segments"`
	NumSpeakers int                  `json:"num_speakers"`
}

type TextSegment struct {
	Text         string   `json:"text"`
	HasFillers   bool     `json:"has_fillers"`
	CleanedText  string   `json:"cleaned_text"`
	FillersFound []string `json:"fillers_found"`
}

type TextProcessResponse struct {
	Segments     []TextSegment `json:"segments"`
	TotalFillers int           `json:"total_fillers"`
}

type TextProcessRequest struct {
	Text          string `json:"text"`
	DetectFillers bool   `json:"detect_fillers"`
	RemoveFillers bool   `json:"remove_fillers"`
}

type WordTimestamp struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

type TranscribeSegment struct {
	Speaker      string          `json:"speaker"`
	Start        float64         `json:"start"`
	End          float64         `json:"end"`
	Text         string          `json:"text"`
	Words        []WordTimestamp  `json:"words"`
	HasFillers   bool            `json:"has_fillers"`
	FillersFound []string        `json:"fillers_found"`
}

type TranscribeFullResponse struct {
	Language              string              `json:"language"`
	FullText              string              `json:"full_text"`
	Segments              []TranscribeSegment `json:"segments"`
	NumSpeakers           int                 `json:"num_speakers"`
	ProcessingTimeSeconds float64             `json:"processing_time_seconds"`
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Minute,
		},
	}
}

// Diarize отправляет аудиофайл на диаризацию.
func (c *Client) Diarize(audioPath string) (*DiarizationResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, err := os.Open(audioPath)
	if err != nil {
		return nil, fmt.Errorf("open audio file: %w", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("audio", filepath.Base(audioPath))
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("copy audio data: %w", err)
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/diarize", body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("diarize request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("diarize error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var result DiarizationResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse diarize response: %w", err)
	}

	return &result, nil
}

// ProcessText отправляет текст на обработку (определение паразитов).
func (c *Client) ProcessText(text string, detectFillers, removeFillers bool) (*TextProcessResponse, error) {
	reqBody := TextProcessRequest{
		Text:          text,
		DetectFillers: detectFillers,
		RemoveFillers: removeFillers,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/process-text", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("process-text request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("process-text error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var result TextProcessResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &result, nil
}

// TranscribeFull отправляет аудиофайл на полный pipeline: транскрибация + диаризация + alignment.
func (c *Client) TranscribeFull(audioPath string, language string, numSpeakers *int, detectFillers bool) (*TranscribeFullResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, err := os.Open(audioPath)
	if err != nil {
		return nil, fmt.Errorf("open audio file: %w", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("audio", filepath.Base(audioPath))
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("copy audio data: %w", err)
	}
	writer.Close()

	// Формируем URL с query-параметрами
	url := c.baseURL + "/transcribe-full?detect_fillers=" + fmt.Sprintf("%t", detectFillers)
	if language != "" {
		url += "&language=" + language
	}
	if numSpeakers != nil {
		url += "&num_speakers=" + fmt.Sprintf("%d", *numSpeakers)
	}

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("transcribe-full request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("transcribe-full error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var result TranscribeFullResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse transcribe-full response: %w", err)
	}

	return &result, nil
}

// Health проверяет доступность ML-сервиса.
func (c *Client) Health() error {
	resp, err := c.httpClient.Get(c.baseURL + "/health")
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ml service unhealthy: status %d", resp.StatusCode)
	}
	return nil
}
