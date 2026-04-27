package cos

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
	"gx1727.com/xin/framework/pkg/storage"
)

type CosStorage struct {
	client  *cos.Client
	baseURL string
}

type Config struct {
	URL       string // https://<bucket>.cos.<region>.myqcloud.com
	SecretID  string
	SecretKey string
	BaseURL   string // Custom domain, e.g., https://img.gx1727.com
}

func NewCosStorage(cfg Config) (storage.Storage, error) {
	u, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid cos url: %w", err)
	}

	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Timeout: 100 * time.Second,
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.SecretID,
			SecretKey: cfg.SecretKey,
		},
	})

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = cfg.URL
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &CosStorage{
		client:  client,
		baseURL: baseURL,
	}, nil
}

func (s *CosStorage) Put(ctx context.Context, key string, file io.Reader) (string, error) {
	_, err := s.client.Object.Put(ctx, key, file, nil)
	if err != nil {
		return "", fmt.Errorf("cos put object failed: %w", err)
	}

	return s.GetURL(ctx, key)
}

func (s *CosStorage) Delete(ctx context.Context, key string) error {
	_, err := s.client.Object.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("cos delete object failed: %w", err)
	}
	return nil
}

func (s *CosStorage) GetURL(ctx context.Context, key string) (string, error) {
	// For public read bucket, just return the concatenated URL
	// If it's a private bucket, we might need to generate a presigned URL here
	return fmt.Sprintf("%s/%s", s.baseURL, key), nil
}
