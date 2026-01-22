package speechkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	recognizeURL     = "https://stt.api.cloud.yandex.net/speech/v1/stt:recognize"
	longRunningURL   = "https://transcribe.api.cloud.yandex.net/speech/stt/v2/longRunningRecognize"
	operationsURL    = "https://operation.api.cloud.yandex.net/operations"
)

type Client struct {
	apiKey     string
	folderId   string
	httpClient *http.Client
}

type RecognizeResponse struct {
	Result string `json:"result"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Async API types
type LongRunningRequest struct {
	Config   RecognitionConfig `json:"config"`
	Audio    AudioSource       `json:"audio"`
	FolderId string            `json:"folderId"`
}

type RecognitionConfig struct {
	Specification RecognitionSpec `json:"specification"`
}

type RecognitionSpec struct {
	LanguageCode       string `json:"languageCode"`
	Model              string `json:"model,omitempty"`
	AudioEncoding      string `json:"audioEncoding"`
	SampleRateHertz    int    `json:"sampleRateHertz"`
	AudioChannelCount  int    `json:"audioChannelCount"`
}

type AudioSource struct {
	URI string `json:"uri"`
}

type Operation struct {
	ID          string           `json:"id"`
	Done        bool             `json:"done"`
	Response    *OperationResult `json:"response,omitempty"`
	Error       *OperationError  `json:"error,omitempty"`
}

type OperationResult struct {
	Chunks []TranscriptChunk `json:"chunks"`
}

type TranscriptChunk struct {
	Alternatives []Alternative `json:"alternatives"`
}

type Alternative struct {
	Text string `json:"text"`
}

type OperationError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewClient(apiKey string, folderId string) *Client {
	return &Client{
		apiKey:   apiKey,
		folderId: folderId,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// RecognizeFile распознаёт аудиофайл и возвращает текст.
// Файл должен быть в формате OGG Opus (конвертируйте через ffmpeg).
// Ограничение: до 30 секунд аудио для синхронного API.
func (c *Client) RecognizeFile(filePath string, lang string) (string, error) {
	audioData, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read audio file: %w", err)
	}

	return c.Recognize(audioData, lang)
}

// Recognize отправляет аудиоданные на распознавание.
// audioData — OGG Opus данные.
// lang — код языка (ru-RU, en-US и т.д.), пустая строка для автоопределения.
func (c *Client) Recognize(audioData []byte, lang string) (string, error) {
	url := recognizeURL + "?format=oggopus"
	if lang != "" {
		url += "&lang=" + lang
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(audioData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Api-Key "+c.apiKey)
	req.Header.Set("Content-Type", "application/octet-stream")
	if c.folderId != "" {
		req.Header.Set("x-folder-id", c.folderId)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Message != "" {
			return "", fmt.Errorf("speechkit error (code %d): %s", errResp.Code, errResp.Message)
		}
		return "", fmt.Errorf("speechkit error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result RecognizeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Result, nil
}

// RecognizeLongAudio запускает асинхронное распознавание для длинных файлов.
// fileURI — URL файла в Yandex Object Storage (https://storage.yandexcloud.net/bucket/key).
// Возвращает текст транскрипта после завершения операции.
func (c *Client) RecognizeLongAudio(fileURI string, lang string) (string, error) {
	// Запускаем операцию
	opID, err := c.startLongRunningRecognition(fileURI, lang)
	if err != nil {
		return "", err
	}

	// Ждём завершения с polling
	return c.waitForOperation(opID)
}

func (c *Client) startLongRunningRecognition(fileURI string, lang string) (string, error) {
	reqBody := LongRunningRequest{
		Config: RecognitionConfig{
			Specification: RecognitionSpec{
				LanguageCode:      lang,
				Model:             "general",
				AudioEncoding:     "OGG_OPUS",
				SampleRateHertz:   48000,
				AudioChannelCount: 1,
			},
		},
		Audio: AudioSource{
			URI: fileURI,
		},
		FolderId: c.folderId,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, longRunningURL, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Api-Key "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	if c.folderId != "" {
		req.Header.Set("x-folder-id", c.folderId)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Message != "" {
			return "", fmt.Errorf("speechkit error: %s", errResp.Message)
		}
		return "", fmt.Errorf("speechkit error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var op Operation
	if err := json.Unmarshal(body, &op); err != nil {
		return "", fmt.Errorf("failed to parse operation: %w", err)
	}

	return op.ID, nil
}

func (c *Client) waitForOperation(opID string) (string, error) {
	url := fmt.Sprintf("%s/%s", operationsURL, opID)

	// Polling с экспоненциальной задержкой: 1s, 2s, 4s, 8s... max 30s
	delay := time.Second
	maxDelay := 30 * time.Second
	maxAttempts := 120 // ~30 минут максимум

	for i := 0; i < maxAttempts; i++ {
		time.Sleep(delay)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "Api-Key "+c.apiKey)
		if c.folderId != "" {
			req.Header.Set("x-folder-id", c.folderId)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			// Retry on network errors
			delay = min(delay*2, maxDelay)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			delay = min(delay*2, maxDelay)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("operation check failed: status %d", resp.StatusCode)
		}

		var op Operation
		if err := json.Unmarshal(body, &op); err != nil {
			return "", fmt.Errorf("failed to parse operation: %w", err)
		}

		if op.Done {
			if op.Error != nil {
				return "", fmt.Errorf("recognition failed: %s", op.Error.Message)
			}
			return extractText(op.Response), nil
		}

		// Увеличиваем задержку
		delay = min(delay*2, maxDelay)
	}

	return "", fmt.Errorf("operation timed out after %d attempts", maxAttempts)
}

func extractText(result *OperationResult) string {
	if result == nil {
		return ""
	}

	var text string
	for _, chunk := range result.Chunks {
		if len(chunk.Alternatives) > 0 {
			if text != "" {
				text += " "
			}
			text += chunk.Alternatives[0].Text
		}
	}
	return text
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
