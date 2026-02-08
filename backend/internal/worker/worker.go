package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"loopa/backend/internal/media"
	"loopa/backend/internal/mlclient"
	"loopa/backend/internal/speechkit"
	"loopa/backend/internal/storage"
)

const (
	// Максимальная длительность для синхронного API (секунды)
	maxSyncDuration = 30.0
	// Длительность одной части при разбиении (секунды)
	chunkDuration = 29
)

type TaskRow struct {
	ID          string
	StoragePath string
}

// S3Config содержит настройки Yandex Object Storage для async API.
type S3Config struct {
	AccessKey string
	SecretKey string
	Bucket    string
}

type Worker struct {
	db           *sql.DB
	speechKit    *speechkit.Client
	mlClient     *mlclient.Client
	s3Client     *storage.S3Client
	uploadDir    string
	provider     string // "whisper" или "speechkit"
	pollInterval time.Duration
}

// New создаёт worker.
func New(db *sql.DB, provider string, apiKey string, folderId string, uploadDir string, mlServiceURL string, s3cfg *S3Config) *Worker {
	var ml *mlclient.Client
	if mlServiceURL != "" {
		ml = mlclient.New(mlServiceURL)
	}

	var sk *speechkit.Client
	if provider == "speechkit" {
		sk = speechkit.NewClient(apiKey, folderId)
	}

	var s3c *storage.S3Client
	if s3cfg != nil {
		var err error
		s3c, err = storage.NewS3Client(s3cfg.AccessKey, s3cfg.SecretKey, s3cfg.Bucket)
		if err != nil {
			log.Printf("S3 client init failed, falling back to chunked mode: %v", err)
		} else {
			log.Println("S3 client initialized — async SpeechKit API enabled for long audio")
		}
	}

	return &Worker{
		db:           db,
		speechKit:    sk,
		mlClient:     ml,
		s3Client:     s3c,
		uploadDir:    uploadDir,
		provider:     provider,
		pollInterval: 2 * time.Second,
	}
}

func (w *Worker) Run(stop <-chan struct{}) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			if err := w.processBatch(); err != nil {
				log.Printf("worker error: %v", err)
			}
		}
	}
}

func (w *Worker) processBatch() error {
	rows, err := w.db.Query(
		`SELECT t.id, f.storage_path
		 FROM transcription_tasks t
		 JOIN files f ON f.id = t.file_id
		 WHERE t.status = 'ожидает'
		 ORDER BY t.created_at
		 LIMIT 5`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	var tasks []TaskRow
	for rows.Next() {
		var t TaskRow
		if err := rows.Scan(&t.ID, &t.StoragePath); err != nil {
			return err
		}
		tasks = append(tasks, t)
	}

	for _, task := range tasks {
		if err := w.processTask(task); err != nil {
			log.Printf("task %s failed: %v", task.ID, err)
		}
	}
	return nil
}

func (w *Worker) processTask(task TaskRow) error {
	startTime := time.Now()
	now := startTime.UTC()
	res, err := w.db.Exec(
		`UPDATE transcription_tasks
		 SET status = 'в процессе', started_at = ?
		 WHERE id = ? AND status = 'ожидает'`,
		now, task.ID,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil || affected == 0 {
		return nil
	}

	if w.provider == "whisper" {
		return w.processTaskWhisper(task, startTime)
	}
	return w.processTaskSpeechKit(task, startTime)
}

// processTaskWhisper — pipeline через Faster-Whisper (ML-сервис /transcribe-full).
func (w *Worker) processTaskWhisper(task TaskRow, startTime time.Time) error {
	inputPath := task.StoragePath

	if w.mlClient == nil {
		return w.failTask(task.ID, "ML-сервис не настроен для Whisper провайдера")
	}

	log.Printf("task %s: starting Whisper transcription", task.ID)

	resp, err := w.mlClient.TranscribeFull(inputPath, "", nil, true)
	if err != nil {
		return w.failTask(task.ID, "Ошибка транскрибации: "+err.Error())
	}

	log.Printf("task %s: transcription done — %d segments, %d speakers, lang=%s (%.1fs)",
		task.ID, len(resp.Segments), resp.NumSpeakers, resp.Language, resp.ProcessingTimeSeconds)

	// Сохраняем данные о спикерах
	speakerJSON, _ := json.Marshal(resp)
	w.db.Exec(
		`UPDATE transcription_tasks SET speaker_data = ? WHERE id = ?`,
		string(speakerJSON), task.ID,
	)

	// Сохраняем сегменты с точным word-level alignment
	now := time.Now().UTC()
	for _, seg := range resp.Segments {
		segID := uuid.New().String()
		startMs := int(seg.Start * 1000)
		endMs := int(seg.End * 1000)

		w.db.Exec(
			`INSERT INTO transcription_segments
			 (id, task_id, speaker_id, start_time, end_time, text, has_fillers, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			segID, task.ID, seg.Speaker, startMs, endMs, seg.Text, seg.HasFillers, now,
		)
	}

	processingTime := int(time.Since(startTime).Seconds())

	_, err = w.db.Exec(
		`UPDATE transcription_tasks
		 SET status = 'готово', transcript_text = ?, provider = 'faster_whisper',
		     processing_time = ?, completed_at = ?
		 WHERE id = ?`,
		resp.FullText, processingTime, time.Now().UTC(), task.ID,
	)
	return err
}

// processTaskSpeechKit — pipeline через Yandex SpeechKit (legacy fallback).
func (w *Worker) processTaskSpeechKit(task TaskRow, startTime time.Time) error {
	inputPath := task.StoragePath

	// Определяем длительность аудио
	duration, err := media.GetDuration(inputPath)
	if err != nil {
		log.Printf("task %s: failed to get duration, using async mode: %v", task.ID, err)
		duration = maxSyncDuration + 1
	}

	// Конвертируем аудио в OGG Opus для SpeechKit
	oggPath, err := media.ExtractAudio(inputPath, w.uploadDir)
	if err != nil {
		return w.failTask(task.ID, "Ошибка конвертации аудио: "+err.Error())
	}
	defer os.Remove(oggPath)

	// Транскрибация через SpeechKit
	var text string
	if duration <= maxSyncDuration {
		text, err = w.speechKit.RecognizeFile(oggPath, "ru-RU")
	} else if w.s3Client != nil {
		text, err = w.recognizeLongAudioAsync(task.ID, oggPath)
	} else {
		text, err = w.recognizeLongAudio(task.ID, oggPath)
	}

	if err != nil {
		return w.failTask(task.ID, "Ошибка распознавания: "+err.Error())
	}

	// Диаризация через ML-сервис (если доступен)
	if w.mlClient != nil {
		w.diarizeAndSaveSegments(task.ID, oggPath, text)
	}

	processingTime := int(time.Since(startTime).Seconds())

	_, err = w.db.Exec(
		`UPDATE transcription_tasks
		 SET status = 'готово', transcript_text = ?, provider = 'yandex_speechkit',
		     processing_time = ?, completed_at = ?
		 WHERE id = ?`,
		text, processingTime, time.Now().UTC(), task.ID,
	)
	return err
}

// diarizeAndSaveSegments выполняет диаризацию и сохраняет сегменты (для SpeechKit pipeline).
func (w *Worker) diarizeAndSaveSegments(taskID, audioPath, transcriptText string) {
	log.Printf("task %s: starting diarization", taskID)

	diarization, err := w.mlClient.Diarize(audioPath)
	if err != nil {
		log.Printf("task %s: diarization failed (non-fatal): %v", taskID, err)
		w.saveSingleSegment(taskID, transcriptText)
		return
	}

	log.Printf("task %s: diarization found %d speakers, %d segments",
		taskID, diarization.NumSpeakers, len(diarization.Segments))

	speakerJSON, _ := json.Marshal(diarization)
	w.db.Exec(
		`UPDATE transcription_tasks SET speaker_data = ? WHERE id = ?`,
		string(speakerJSON), taskID,
	)

	var textProcessed *mlclient.TextProcessResponse
	if w.mlClient != nil {
		textProcessed, err = w.mlClient.ProcessText(transcriptText, true, false)
		if err != nil {
			log.Printf("task %s: text processing failed (non-fatal): %v", taskID, err)
		}
	}

	now := time.Now().UTC()
	words := strings.Fields(transcriptText)
	totalWords := len(words)
	numSegments := len(diarization.Segments)

	for i, seg := range diarization.Segments {
		segID := uuid.New().String()
		startMs := int(seg.Start * 1000)
		endMs := int(seg.End * 1000)

		// Распределяем слова пропорционально по сегментам
		segStart := i * totalWords / numSegments
		segEnd := (i + 1) * totalWords / numSegments
		if segEnd > totalWords {
			segEnd = totalWords
		}
		segText := ""
		if segStart < segEnd {
			segText = strings.Join(words[segStart:segEnd], " ")
		}

		hasFillers := false
		if textProcessed != nil && len(textProcessed.Segments) > 0 {
			hasFillers = textProcessed.Segments[0].HasFillers
		}

		w.db.Exec(
			`INSERT INTO transcription_segments
			 (id, task_id, speaker_id, start_time, end_time, text, has_fillers, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			segID, taskID, seg.Speaker, startMs, endMs, segText, hasFillers, now,
		)
	}
}

// saveSingleSegment сохраняет весь текст как один сегмент (fallback без диаризации).
func (w *Worker) saveSingleSegment(taskID, text string) {
	now := time.Now().UTC()
	segID := uuid.New().String()

	hasFillers := false
	if w.mlClient != nil {
		resp, err := w.mlClient.ProcessText(text, true, false)
		if err == nil && resp.TotalFillers > 0 {
			hasFillers = true
		}
	}

	w.db.Exec(
		`INSERT INTO transcription_segments
		 (id, task_id, start_time, end_time, text, has_fillers, created_at)
		 VALUES (?, ?, 0, 0, ?, ?, ?)`,
		segID, taskID, text, hasFillers, now,
	)
}

// recognizeLongAudioAsync загружает файл в S3 и использует async SpeechKit API.
func (w *Worker) recognizeLongAudioAsync(taskID, oggPath string) (string, error) {
	log.Printf("task %s: uploading to S3 for async recognition", taskID)

	key := storage.GenerateKey("audio", fmt.Sprintf("%s_%s", taskID, filepath.Base(oggPath)))
	ctx := context.Background()

	s3URI, err := w.s3Client.Upload(ctx, oggPath, key)
	if err != nil {
		return "", fmt.Errorf("S3 upload failed: %w", err)
	}

	defer func() {
		if delErr := w.s3Client.Delete(ctx, key); delErr != nil {
			log.Printf("task %s: failed to delete S3 object: %v", taskID, delErr)
		}
	}()

	log.Printf("task %s: starting async recognition (URI: %s)", taskID, s3URI)

	text, err := w.speechKit.RecognizeLongAudio(s3URI, "ru-RU")
	if err != nil {
		return "", fmt.Errorf("async recognition failed: %w", err)
	}

	return text, nil
}

func (w *Worker) recognizeLongAudio(taskID, inputPath string) (string, error) {
	log.Printf("task %s: splitting long audio into chunks", taskID)

	chunks, err := media.SplitAudio(inputPath, w.uploadDir, chunkDuration)
	if err != nil {
		return "", err
	}

	defer func() {
		for _, chunk := range chunks {
			os.Remove(chunk)
		}
	}()

	log.Printf("task %s: processing %d chunks", taskID, len(chunks))

	var results []string
	for i, chunk := range chunks {
		log.Printf("task %s: recognizing chunk %d/%d", taskID, i+1, len(chunks))
		text, err := w.speechKit.RecognizeFile(chunk, "ru-RU")
		if err != nil {
			return "", err
		}
		if text != "" {
			results = append(results, text)
		}
	}

	return strings.Join(results, " "), nil
}

func (w *Worker) failTask(taskID string, errMsg string) error {
	log.Printf("task %s error: %s", taskID, errMsg)
	_, err := w.db.Exec(
		`UPDATE transcription_tasks
		 SET status = 'ошибка', error_message = ?, completed_at = ?
		 WHERE id = ?`,
		errMsg, time.Now().UTC(), taskID,
	)
	return err
}
