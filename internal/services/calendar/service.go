package calendar

import (
	"context"
	"fmt"

	"github.com/erikkubica/gmcp/internal/auth"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Service wraps the official Google Calendar API client.
type Service struct {
	client *calendar.Service
}

// NewService creates an authenticated Calendar service wrapper.
func NewService(ctx context.Context) (*Service, error) {
	httpClient, err := auth.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("auth client: %w", err)
	}

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create calendar service: %w", err)
	}

	return &Service{client: srv}, nil
}
