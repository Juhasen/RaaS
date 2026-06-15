package service

import (
	"context"
	"mime/multipart"
	"net/textproto"
	"os"
	"testing"

	"github.com/Juhasen/RaaS/media/models"
)

func TestLoadEnv(t *testing.T) {
	// Create a temporary .env file
	tempFile, err := os.CreateTemp("", "test_env_*.env")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	content := `
# A comment line
TEST_KEY_1=value1
  TEST_KEY_2 = "value 2"  
TEST_KEY_3 = 'value_3'
`
	if _, err := tempFile.WriteString(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	_ = tempFile.Close()

	// Load the env variables
	if err := LoadEnv(tempFile.Name()); err != nil {
		t.Fatalf("LoadEnv failed: %v", err)
	}

	// Verify values
	if os.Getenv("TEST_KEY_1") != "value1" {
		t.Errorf("expected TEST_KEY_1 to be 'value1', got '%s'", os.Getenv("TEST_KEY_1"))
	}
	if os.Getenv("TEST_KEY_2") != "value 2" {
		t.Errorf("expected TEST_KEY_2 to be 'value 2', got '%s'", os.Getenv("TEST_KEY_2"))
	}
	if os.Getenv("TEST_KEY_3") != "value_3" {
		t.Errorf("expected TEST_KEY_3 to be 'value_3', got '%s'", os.Getenv("TEST_KEY_3"))
	}

	// Cleanup env vars
	_ = os.Unsetenv("TEST_KEY_1")
	_ = os.Unsetenv("TEST_KEY_2")
	_ = os.Unsetenv("TEST_KEY_3")
}

type mockMediaRepository struct{}

func (m *mockMediaRepository) Save(ctx context.Context, media *models.Media) error {
	return nil
}
func (m *mockMediaRepository) FindByID(ctx context.Context, id string) (*models.Media, error) {
	return nil, nil
}
func (m *mockMediaRepository) FindByListingID(ctx context.Context, listingID string) ([]*models.Media, error) {
	return nil, nil
}

func TestUploadMedia_Validation(t *testing.T) {
	mockRepo := &mockMediaRepository{}
	// Instantiate service without brokers (disabling Kafka client initialization)
	srv := NewMediaService(mockRepo, &Config{}, nil, "")

	// 1. Test invalid type
	header := &multipart.FileHeader{
		Filename: "test.pdf",
		Size:     123,
		Header:   make(textproto.MIMEHeader),
	}
	header.Header.Set("Content-Type", "application/pdf")

	_, err := srv.UploadMedia(context.Background(), "listing-123", header)
	if err == nil {
		t.Error("expected error for PDF upload, got nil")
	}

	// 2. Test valid type (will fail during local file create because header is a mock without underlying file content, but we test the MIME check)
	header.Filename = "test.png"
	header.Header.Set("Content-Type", "image/png")
	_, err = srv.UploadMedia(context.Background(), "listing-123", header)
	// It should fail on file.Open() because src is nil, which confirms it passed the MIME type validation!
	if err == nil {
		t.Error("expected error on file open, got nil")
	}
	if err.Error() == "invalid file type image/png: only images are allowed" {
		t.Error("PNG was rejected by MIME validation")
	}
}
