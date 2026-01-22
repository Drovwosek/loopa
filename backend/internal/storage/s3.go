package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const yandexStorageEndpoint = "https://storage.yandexcloud.net"

type S3Client struct {
	client *s3.Client
	bucket string
}

// NewS3Client создаёт клиент для Yandex Object Storage.
func NewS3Client(accessKey, secretKey, bucket string) (*S3Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
		config.WithRegion("ru-central1"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(yandexStorageEndpoint)
	})

	return &S3Client{
		client: client,
		bucket: bucket,
	}, nil
}

// Upload загружает файл в Object Storage и возвращает S3 URI.
func (c *S3Client) Upload(ctx context.Context, localPath, key string) (string, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	_, err = c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload: %w", err)
	}

	// Возвращаем URI в формате для SpeechKit
	uri := fmt.Sprintf("https://storage.yandexcloud.net/%s/%s", c.bucket, key)
	return uri, nil
}

// Delete удаляет файл из Object Storage.
func (c *S3Client) Delete(ctx context.Context, key string) error {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	return err
}

// GenerateKey генерирует уникальный ключ для файла.
func GenerateKey(prefix, filename string) string {
	return filepath.Join(prefix, filename)
}
