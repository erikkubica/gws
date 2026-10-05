package gmail

import (
	"context"
	"fmt"

	"github.com/erikkubica/gmcp/internal/auth"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Service wraps the official Google Gmail API client.
type Service struct {
	client *gmail.Service
}

// NewService creates an authenticated Gmail service wrapper.
func NewService(ctx context.Context) (*Service, error) {
	httpClient, err := auth.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("auth client: %w", err)
	}

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create gmail service: %w", err)
	}

	return &Service{client: srv}, nil
}
