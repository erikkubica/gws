package youtube

import (
	"context"
	"fmt"

	"github.com/erikkubica/gmcp/internal/auth"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// Service wraps the official Google YouTube v3 API client.
type Service struct {
	client *youtube.Service
}

// NewService creates an authenticated YouTube service wrapper.
func NewService(ctx context.Context) (*Service, error) {
	httpClient, err := auth.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("auth client: %w", err)
	}

	srv, err := youtube.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create youtube service: %w", err)
	}

	return &Service{client: srv}, nil
}
