package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/youtube/v3"
)

var Scopes = []string{
	gmail.GmailModifyScope,
	calendar.CalendarScope,
	drive.DriveScope,
	youtube.YoutubeReadonlyScope,
	"https://www.googleapis.com/auth/tasks",
	"https://www.googleapis.com/auth/spreadsheets",
}

// LoadOAuthConfig reads credentials.json and builds an oauth2.Config.
func LoadOAuthConfig(redirectURL string) (*oauth2.Config, error) {
	credFile, err := CredentialsPath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(credFile)
	if err != nil {
		return nil, fmt.Errorf("read credentials file (%s): %w", credFile, err)
	}
	cfg, err := google.ConfigFromJSON(b, Scopes...)
	if err != nil {
		return nil, fmt.Errorf("parse credentials JSON: %w", err)
	}
	if redirectURL != "" {
		cfg.RedirectURL = redirectURL
	}
	return cfg, nil
}

// SaveToken persists the OAuth2 token into token.json.
func SaveToken(token *oauth2.Token) error {
	tokPath, err := TokenPath()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(tokPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("open token file: %w", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

// LoadToken reads token.json from disk.
func LoadToken() (*oauth2.Token, error) {
	tokPath, err := TokenPath()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(tokPath)
	if err != nil {
		return nil, fmt.Errorf("open token file: %w", err)
	}
	defer f.Close()
	tok := &oauth2.Token{}
	if err := json.NewDecoder(f).Decode(tok); err != nil {
		return nil, fmt.Errorf("decode token: %w", err)
	}
	return tok, nil
}

// GetClient returns an authenticated HTTP client backed by auto-refreshing token.
func GetClient(ctx context.Context) (*http.Client, error) {
	cfg, err := LoadOAuthConfig("")
	if err != nil {
		return nil, err
	}
	tok, err := LoadToken()
	if err != nil {
		return nil, fmt.Errorf("no active session found: please run 'gws auth login' first: %w", err)
	}
	return cfg.Client(ctx, tok), nil
}
