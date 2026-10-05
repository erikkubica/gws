package tasks

import (
	"context"
	"fmt"

	"github.com/erikkubica/gmcp/internal/auth"
	"google.golang.org/api/option"
	"google.golang.org/api/tasks/v1"
)

// Service wraps the official Google Tasks API client.
type Service struct {
	client *tasks.Service
}

// NewService creates an authenticated Tasks service wrapper.
func NewService(ctx context.Context) (*Service, error) {
	httpClient, err := auth.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("auth client: %w", err)
	}

	srv, err := tasks.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create tasks service: %w", err)
	}

	return &Service{client: srv}, nil
}
