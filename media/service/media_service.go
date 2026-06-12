package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Juhasen/RaaS/media/models"
	"github.com/Juhasen/RaaS/media/repository"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

// LoadEnv reads a local .env file and populates environment variables.
func LoadEnv(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Use system env variables if file doesn't exist
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
			(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
			val = val[1 : len(val)-1]
		}
		if key != "" {
			_ = os.Setenv(key, val)
		}
	}
	return scanner.Err()
}

// MediaService orchestrates media uploads and metadata management.
type MediaService struct {
	repo         repository.MediaRepository
	s3Client     *s3.Client
	kafkaWriter  *kafka.Writer
	bucketName   string
	publicURL    string
	localStorage string
}

// NewMediaService instantiates a new MediaService.
func NewMediaService(repo repository.MediaRepository, kafkaBrokers []string, kafkaTopic string) *MediaService {
	accountID := os.Getenv("R2_ACCOUNT_ID")
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	bucketName := os.Getenv("R2_BUCKET_NAME")
	if bucketName == "" {
		bucketName = os.Getenv("S3_BUCKET")
		if bucketName == "" {
			bucketName = "raas-media-bucket"
		}
	}
	publicURL := os.Getenv("R2_PUBLIC_URL")

	var s3Client *s3.Client
	if accountID != "" && accessKey != "" && secretKey != "" {
		r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:           fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID),
				SigningRegion: region,
			}, nil
		})

		cfg, err := config.LoadDefaultConfig(context.Background(),
			config.WithEndpointResolverWithOptions(r2Resolver),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		)
		if err == nil {
			s3Client = s3.NewFromConfig(cfg)
			log.Println("Initialized Cloudflare R2 S3 Client successfully")
		} else {
			log.Printf("Failed to initialize R2 configuration: %v", err)
		}
	} else {
		log.Println("Cloudflare R2 credentials not fully provided; running in local fallback directory mode")
	}

	var kw *kafka.Writer
	if len(kafkaBrokers) > 0 {
		kw = &kafka.Writer{
			Addr:     kafka.TCP(kafkaBrokers...),
			Topic:    kafkaTopic,
			Balancer: &kafka.LeastBytes{},
		}
		log.Printf("Initialized Media Kafka writer on topic %s", kafkaTopic)
	}

	localStore := "./uploads"
	_ = os.MkdirAll(localStore, 0755)

	return &MediaService{
		repo:         repo,
		s3Client:     s3Client,
		kafkaWriter:  kw,
		bucketName:   bucketName,
		publicURL:    publicURL,
		localStorage: localStore,
	}
}

// UploadMedia checks the file MIME type, uploads to Cloudflare R2 or local disk, saves metadata, and posts a Kafka event.
func (s *MediaService) UploadMedia(ctx context.Context, listingID string, file *multipart.FileHeader) (*models.Media, error) {
	contentType := file.Header.Get("Content-Type")
	allowedMimeTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	if !allowedMimeTypes[strings.ToLower(contentType)] {
		return nil, fmt.Errorf("invalid file type %s: only images are allowed", contentType)
	}

	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	id := uuid.New().String()
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		parts := strings.Split(contentType, "/")
		if len(parts) == 2 {
			ext = "." + parts[1]
		}
	}
	key := fmt.Sprintf("%s/%s%s", listingID, id, ext)

	var fileURL string

	if s.s3Client != nil {
		_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(s.bucketName),
			Key:         aws.String(key),
			Body:        src,
			ContentType: aws.String(contentType),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to upload object to Cloudflare R2: %w", err)
		}

		if s.publicURL != "" {
			fileURL = fmt.Sprintf("%s/%s", strings.TrimSuffix(s.publicURL, "/"), key)
		} else {
			fileURL = fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/%s", os.Getenv("R2_ACCOUNT_ID"), s.bucketName, key)
		}
	} else {
		destPath := filepath.Join(s.localStorage, listingID)
		if err := os.MkdirAll(destPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create local directory: %w", err)
		}

		filePath := filepath.Join(destPath, id+ext)
		destFile, err := os.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create local file: %w", err)
		}
		defer destFile.Close()

		if _, err := io.Copy(destFile, src); err != nil {
			return nil, fmt.Errorf("failed to copy local file: %w", err)
		}

		fileURL = fmt.Sprintf("http://localhost:8080/uploads/%s", key)
	}

	media := &models.Media{
		ID:        id,
		ListingID: listingID,
		URL:       fileURL,
		Type:      "IMAGE",
		CreatedAt: time.Now(),
	}

	if err := s.repo.Save(ctx, media); err != nil {
		return nil, fmt.Errorf("failed to save media metadata: %w", err)
	}

	if s.kafkaWriter != nil {
		eventMsg := fmt.Sprintf(`{"event":"media.uploaded","media_id":"%s","listing_id":"%s","url":"%s","type":"%s"}`, media.ID, media.ListingID, media.URL, media.Type)
		err := s.kafkaWriter.WriteMessages(ctx,
			kafka.Message{
				Key:   []byte(media.ID),
				Value: []byte(eventMsg),
			},
		)
		if err != nil {
			log.Printf("Failed to emit Kafka media.uploaded event: %v", err)
		}
	}

	return media, nil
}

// FindByID retrieves a media record by ID.
func (s *MediaService) FindByID(ctx context.Context, id string) (*models.Media, error) {
	return s.repo.FindByID(ctx, id)
}

// FindByListingID retrieves all media records associated with listingID.
func (s *MediaService) FindByListingID(ctx context.Context, listingID string) ([]*models.Media, error) {
	return s.repo.FindByListingID(ctx, listingID)
}

// Close closes S3 and Kafka connections gracefully.
func (s *MediaService) Close() {
	if s.kafkaWriter != nil {
		_ = s.kafkaWriter.Close()
	}
}
