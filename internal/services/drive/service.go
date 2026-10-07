package drive

import (
	"context"
	"fmt"

	"github.com/erikkubica/gws/internal/auth"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Service wraps the official Google Drive API client.
type Service struct {
	client *drive.Service
}

// NewService creates an authenticated Drive service wrapper.
func NewService(ctx context.Context) (*Service, error) {
	httpClient, err := auth.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("auth client: %w", err)
	}

	srv, err := drive.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create drive service: %w", err)
	}

	return &Service{client: srv}, nil
}
