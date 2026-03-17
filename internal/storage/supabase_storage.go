package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
)

type SupabaseStorage struct {
	baseURL string
	bucket  string
	apiKey  string
}

func NewSupabaseStorage(url, bucket, apiKey string) *SupabaseStorage {
	log.Printf("Supabase Storage: URL=%s, Bucket=%s, KeyLen=%d", url, bucket, len(apiKey))
	return &SupabaseStorage{
		baseURL: url,
		bucket:  bucket,
		apiKey:  apiKey,
	}
}

// UploadFile uploads a file to Supabase Storage and returns the public URL
func (s *SupabaseStorage) UploadFile(ctx context.Context, path string, file io.Reader) (string, error) {
	uploadURL := fmt.Sprintf(
		"%s/storage/v1/object/%s/%s",
		s.baseURL,
		s.bucket,
		path,
	)

	log.Printf("Uploading to: %s", uploadURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, file)
	if err != nil {
		return "", fmt.Errorf("failed to create upload request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("x-upsert", "true") // overwrite existing file instead of 409 Conflict

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Supabase error response: %s", string(body))
		return "", fmt.Errorf("upload failed with status: %s", resp.Status)
	}

	publicURL := fmt.Sprintf(
		"%s/storage/v1/object/public/%s/%s",
		s.baseURL,
		s.bucket,
		path,
	)

	return publicURL, nil

}
