package docs

import (
	"context"
	"fmt"

	"github.com/erikkubica/gmcp/internal/auth"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
)

// Service wraps the official Google Docs API client.
type Service struct {
	client *docs.Service
}

// NewService creates an authenticated Docs service wrapper.
func NewService(ctx context.Context) (*Service, error) {
	httpClient, err := auth.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("auth client: %w", err)
	}

	srv, err := docs.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create docs service: %w", err)
	}

	return &Service{client: srv}, nil
}
