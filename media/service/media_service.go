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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"

	"github.com/Juhasen/RaaS/media/models"
	"github.com/Juhasen/RaaS/media/repository"
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

	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil // If it's a directory (e.g. empty mounted folder in k8s), ignore it
	}

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
func NewMediaService(repo repository.MediaRepository, cfg *Config, kafkaBrokers []string, kafkaTopic string) *MediaService {
	accountID := cfg.R2AccountID
	accessKey := cfg.R2AccessKeyID
	secretKey := cfg.R2SecretAccessKey
	bucketName := cfg.R2BucketName
	publicURL := cfg.R2PublicURL

	var s3Client *s3.Client
	if accountID != "" && accessKey != "" && secretKey != "" {
		awsCfg, err := config.LoadDefaultConfig(context.Background(),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
			config.WithRegion("auto"),
		)
		if err == nil {
			s3Client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
				o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID))
			})
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
	log.Printf("Starting media upload for listing %s: filename=%s content_type=%s size=%d", listingID, file.Filename, file.Header.Get("Content-Type"), file.Size)

	contentType := file.Header.Get("Content-Type")
	allowedMimeTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	if !allowedMimeTypes[strings.ToLower(contentType)] {
		log.Printf("Rejected media upload for listing %s due to invalid content type %s", listingID, contentType)
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
	log.Printf("Prepared media storage key for listing %s: %s", listingID, key)

	var fileURL string

	if s.s3Client != nil {
		log.Printf("Uploading media for listing %s to Cloudflare R2 bucket %s", listingID, s.bucketName)
		_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(s.bucketName),
			Key:         aws.String(key),
			Body:        src,
			ContentType: aws.String(contentType),
		})
		if err != nil {
			log.Printf("Failed to upload media object for listing %s to Cloudflare R2: %v", listingID, err)
			return nil, fmt.Errorf("failed to upload object to Cloudflare R2: %w", err)
		}

		if s.publicURL != "" {
			fileURL = fmt.Sprintf("%s/%s", strings.TrimSuffix(s.publicURL, "/"), key)
		} else {
			fileURL = fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/%s", os.Getenv("R2_ACCOUNT_ID"), s.bucketName, key)
		}
		log.Printf("Stored media for listing %s in R2 with url %s", listingID, fileURL)
	} else {
		log.Printf("Uploading media for listing %s to local storage under %s", listingID, s.localStorage)
		destPath := filepath.Join(s.localStorage, listingID)
		if err := os.MkdirAll(destPath, 0755); err != nil {
			log.Printf("Failed to create local media directory for listing %s: %v", listingID, err)
			return nil, fmt.Errorf("failed to create local directory: %w", err)
		}

		filePath := filepath.Join(destPath, id+ext)
		destFile, err := os.Create(filePath)
		if err != nil {
			log.Printf("Failed to create local media file for listing %s: %v", listingID, err)
			return nil, fmt.Errorf("failed to create local file: %w", err)
		}
		defer destFile.Close()

		if _, err := io.Copy(destFile, src); err != nil {
			log.Printf("Failed to write local media file for listing %s: %v", listingID, err)
			return nil, fmt.Errorf("failed to copy local file: %w", err)
		}

		fileURL = fmt.Sprintf("http://localhost:8080/uploads/%s", key)
		log.Printf("Stored media for listing %s locally at %s", listingID, fileURL)
	}

	media := &models.Media{
		ID:        id,
		ListingID: listingID,
		URL:       fileURL,
		Type:      "IMAGE",
		CreatedAt: time.Now(),
	}

	if err := s.repo.Save(ctx, media); err != nil {
		log.Printf("Failed to save media metadata for listing %s (media_id=%s): %v", listingID, media.ID, err)
		return nil, fmt.Errorf("failed to save media metadata: %w", err)
	}
	log.Printf("Saved media metadata for listing %s with media_id=%s", listingID, media.ID)

	if s.kafkaWriter != nil {
		eventMsg := fmt.Sprintf(`{"event":"media.uploaded","media_id":"%s","listing_id":"%s","url":"%s","type":"%s"}`, media.ID, media.ListingID, media.URL, media.Type)
		log.Printf("Publishing media.uploaded event for listing %s with media_id=%s", listingID, media.ID)
		err := s.kafkaWriter.WriteMessages(ctx,
			kafka.Message{
				Key:   []byte(media.ID),
				Value: []byte(eventMsg),
			},
		)
		if err != nil {
			log.Printf("Failed to emit Kafka media.uploaded event: %v", err)
		} else {
			log.Printf("Published media.uploaded event for listing %s with media_id=%s", listingID, media.ID)
		}
	} else {
		log.Printf("Kafka writer is disabled; media.uploaded event not published for listing %s with media_id=%s", listingID, media.ID)
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

// DeleteMediaByListingID removes the local directory or R2 objects associated with listingID, and deletes database records.
func (s *MediaService) DeleteMediaByListingID(ctx context.Context, listingID string) error {
	// 1. Fetch metadata records first
	mediaList, err := s.repo.FindByListingID(ctx, listingID)
	if err != nil {
		return fmt.Errorf("failed to fetch media list: %w", err)
	}

	// 2. Delete actual files
	if s.s3Client != nil {
		// Delete objects from Cloudflare R2
		for _, m := range mediaList {
			ext := filepath.Ext(m.URL)
			key := fmt.Sprintf("%s/%s%s", listingID, m.ID, ext)

			_, err = s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(s.bucketName),
				Key:    aws.String(key),
			})
			if err != nil {
				log.Printf("Warning: failed to delete S3 object %s: %v", key, err)
			}
		}
	} else {
		// Delete local directory
		destPath := filepath.Join(s.localStorage, listingID)
		if err := os.RemoveAll(destPath); err != nil {
			log.Printf("Warning: failed to delete local media directory %s: %v", destPath, err)
		}
	}

	// 3. Delete metadata records from DB
	if err := s.repo.DeleteByListingID(ctx, listingID); err != nil {
		return fmt.Errorf("failed to delete media database records: %w", err)
	}

	log.Printf("Successfully cleaned up media for listing %s", listingID)
	return nil
}

// Close closes S3 and Kafka connections gracefully.
func (s *MediaService) Close() {
	if s.kafkaWriter != nil {
		_ = s.kafkaWriter.Close()
	}
}
