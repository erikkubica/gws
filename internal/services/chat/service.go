package chat

import (
	"context"
	"fmt"
	"net/http"

	"github.com/erikkubica/gws/internal/auth"
	"google.golang.org/api/chat/v1"
	"google.golang.org/api/option"
)

// Service wraps the official Google Chat API client.
type Service struct {
	client     *chat.Service
	httpClient *http.Client
}

// NewService creates an authenticated Google Chat service wrapper.
func NewService(ctx context.Context) (*Service, error) {
	httpClient, err := auth.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("auth client: %w", err)
	}

	srv, err := chat.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create chat service: %w", err)
	}

	return &Service{
		client:     srv,
		httpClient: httpClient,
	}, nil
}
