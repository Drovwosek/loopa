package worker

import (
	"database/sql"
	"log"
	"os"
	"strings"
	"time"

	"loopa/backend/internal/media"
	"loopa/backend/internal/speechkit"
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

type Worker struct {
	db           *sql.DB
	speechKit    *speechkit.Client
	uploadDir    string
	pollInterval time.Duration
}

// New создаёт worker.
func New(db *sql.DB, apiKey string, folderId string, uploadDir string) *Worker {
	return &Worker{
		db:           db,
		speechKit:    speechkit.NewClient(apiKey, folderId),
		uploadDir:    uploadDir,
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
	now := time.Now().UTC()
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

	// StoragePath уже содержит полный путь к файлу
	inputPath := task.StoragePath

	// Определяем длительность аудио
	duration, err := media.GetDuration(inputPath)
	if err != nil {
		log.Printf("task %s: failed to get duration, using async mode: %v", task.ID, err)
		duration = maxSyncDuration + 1 // Используем async если не можем определить
	}

	// Конвертируем аудио в OGG Opus для SpeechKit
	oggPath, err := media.ExtractAudio(inputPath, w.uploadDir)
	if err != nil {
		return w.failTask(task.ID, "Ошибка конвертации аудио: "+err.Error())
	}
	defer os.Remove(oggPath)

	var text string
	if duration <= maxSyncDuration {
		// Короткое аудио — один запрос
		text, err = w.speechKit.RecognizeFile(oggPath, "ru-RU")
	} else {
		// Длинное аудио — разбиваем на части
		text, err = w.recognizeLongAudio(task.ID, oggPath)
	}

	if err != nil {
		return w.failTask(task.ID, "Ошибка распознавания: "+err.Error())
	}

	_, err = w.db.Exec(
		`UPDATE transcription_tasks
		 SET status = 'готово', transcript_text = ?, provider = 'yandex_speechkit', completed_at = ?
		 WHERE id = ?`,
		text, time.Now().UTC(), task.ID,
	)
	return err
}

func (w *Worker) recognizeLongAudio(taskID, inputPath string) (string, error) {
	log.Printf("task %s: splitting long audio into chunks", taskID)

	// Разбиваем на части по 29 секунд
	chunks, err := media.SplitAudio(inputPath, w.uploadDir, chunkDuration)
	if err != nil {
		return "", err
	}

	// Удаляем временные файлы после обработки
	defer func() {
		for _, chunk := range chunks {
			os.Remove(chunk)
		}
	}()

	log.Printf("task %s: processing %d chunks", taskID, len(chunks))

	// Распознаём каждую часть
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
