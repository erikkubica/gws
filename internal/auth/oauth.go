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
	"https://www.googleapis.com/auth/userinfo.email",
}

// SelectedAccount stores an account override set via CLI flags or programmatic calls.
var SelectedAccount string

type accountCtxKey struct{}

// WithAccount returns a context with the targeted account attached.
func WithAccount(ctx context.Context, account string) context.Context {
	return context.WithValue(ctx, accountCtxKey{}, account)
}

// AccountFromContext extracts the account override from context.
func AccountFromContext(ctx context.Context) string {
	if val, ok := ctx.Value(accountCtxKey{}).(string); ok && val != "" {
		return val
	}
	return ""
}

// LoadOAuthConfig reads credentials.json and builds an oauth2.Config.
func LoadOAuthConfig(redirectURL string) (*oauth2.Config, error) {
	credFile, err := CredentialsPath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(credFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("GCP OAuth credentials not found: run 'gws gcp import <file>' or 'gws gcp set <client_id> <client_secret>' first")
		}
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

// FetchUserEmail queries Google Userinfo or Gmail Profile to obtain the authenticated email.
func FetchUserEmail(ctx context.Context, client *http.Client) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err == nil {
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var info struct {
				Email string `json:"email"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&info); err == nil && info.Email != "" {
				return info.Email, nil
			}
		}
	}
	return fetchEmailFromGmailProfile(ctx, client)
}

func fetchEmailFromGmailProfile(ctx context.Context, client *http.Client) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://gmail.googleapis.com/gmail/v1/users/me/profile", nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch gmail profile: http %d", resp.StatusCode)
	}
	var prof struct {
		EmailAddress string `json:"emailAddress"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&prof); err != nil {
		return "", err
	}
	return prof.EmailAddress, nil
}

// ResolveToken resolves the active token and its associated account name.
func ResolveToken(ctx context.Context) (*oauth2.Token, string, error) {
	_ = MigrateLegacyToken()
	target := determineTargetAccount(ctx)
	if target != "" {
		tok, err := LoadAccountToken(target)
		if err == nil {
			return tok, target, nil
		}
	}
	tok, err := loadLegacyToken()
	if err == nil {
		return tok, "default", nil
	}
	return nil, "", fmt.Errorf("not authenticated: run 'gws auth login'")
}

func determineTargetAccount(ctx context.Context) string {
	if acc := AccountFromContext(ctx); acc != "" {
		return acc
	}
	if SelectedAccount != "" {
		return SelectedAccount
	}
	acc, err := GetActiveAccount()
	if err == nil && acc != "" {
		return acc
	}
	return ""
}

func loadLegacyToken() (*oauth2.Token, error) {
	tokPath, err := TokenPath()
	if err != nil {
		return nil, err
	}
	return readTokenFromFile(tokPath)
}

// SaveToken persists the OAuth2 token into the active account or legacy token.json.
func SaveToken(token *oauth2.Token) error {
	act, err := GetActiveAccount()
	if err == nil && act != "" {
		return SaveAccountToken(act, token)
	}
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

// LoadToken reads the active OAuth2 token from disk.
func LoadToken() (*oauth2.Token, error) {
	tok, _, err := ResolveToken(context.Background())
	return tok, err
}

// GetClient returns an authenticated HTTP client backed by auto-refreshing token.
func GetClient(ctx context.Context) (*http.Client, error) {
	cfg, err := LoadOAuthConfig("")
	if err != nil {
		return nil, err
	}
	tok, _, err := ResolveToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("not authenticated: run 'gws auth login'")
	}
	return cfg.Client(ctx, tok), nil
}
