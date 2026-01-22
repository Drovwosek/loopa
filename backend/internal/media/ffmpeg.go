package media

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// ExtractAudio извлекает аудио из медиафайла и конвертирует в OGG Opus.
// Формат OGG Opus оптимален для Yandex SpeechKit.
func ExtractAudio(inputPath, outputDir string) (string, error) {
	outputName := fmt.Sprintf("%s.ogg", uuid.New().String())
	outputPath := filepath.Join(outputDir, outputName)

	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-i", inputPath,
		"-vn",
		"-acodec", "libopus",
		"-ar", "48000",
		"-ac", "1",
		"-b:a", "64k",
		outputPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("ffmpeg failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return outputPath, nil
}

// GetDuration возвращает длительность медиафайла в секундах.
func GetDuration(inputPath string) (float64, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		inputPath,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}

	duration, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}

// SplitAudio разбивает аудио на части указанной длительности (секунды).
// Возвращает список путей к частям.
func SplitAudio(inputPath, outputDir string, chunkDuration int) ([]string, error) {
	duration, err := GetDuration(inputPath)
	if err != nil {
		return nil, err
	}

	var chunks []string
	for start := 0.0; start < duration; start += float64(chunkDuration) {
		chunkName := fmt.Sprintf("%s_chunk_%d.ogg", uuid.New().String(), int(start))
		chunkPath := filepath.Join(outputDir, chunkName)

		cmd := exec.Command(
			"ffmpeg",
			"-y",
			"-i", inputPath,
			"-ss", fmt.Sprintf("%.2f", start),
			"-t", fmt.Sprintf("%d", chunkDuration),
			"-vn",
			"-acodec", "libopus",
			"-ar", "48000",
			"-ac", "1",
			"-b:a", "64k",
			chunkPath,
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			// Cleanup created chunks on error
			for _, c := range chunks {
				os.Remove(c)
			}
			return nil, fmt.Errorf("ffmpeg split failed: %w: %s", err, strings.TrimSpace(string(out)))
		}
		chunks = append(chunks, chunkPath)
	}

	return chunks, nil
}
