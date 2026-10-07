package sheets

import (
	"context"
	"fmt"

	"github.com/erikkubica/gws/internal/auth"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Service wraps the official Google Sheets API client.
type Service struct {
	client *sheets.Service
}

// NewService creates an authenticated Sheets service wrapper.
func NewService(ctx context.Context) (*Service, error) {
	httpClient, err := auth.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("auth client: %w", err)
	}

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create sheets service: %w", err)
	}

	return &Service{client: srv}, nil
}
