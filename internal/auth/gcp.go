package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// GCPCredentials represents Google Cloud OAuth application client credentials.
type GCPCredentials struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	ProjectID    string   `json:"project_id,omitempty"`
	AuthURI      string   `json:"auth_uri,omitempty"`
	TokenURI     string   `json:"token_uri,omitempty"`
	RedirectURIs []string `json:"redirect_uris,omitempty"`
}

type gcpRawConfig struct {
	Installed *GCPCredentials `json:"installed"`
	Web       *GCPCredentials `json:"web"`
	ClientID  string          `json:"client_id"`
	Secret    string          `json:"client_secret"`
	ProjectID string          `json:"project_id"`
}

// ParseGCPCredentials parses credentials from JSON data.
func ParseGCPCredentials(data []byte) (*GCPCredentials, error) {
	var raw gcpRawConfig
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid json format: %w", err)
	}
	creds := resolveRawCredentials(&raw)
	if creds.ClientID == "" || creds.ClientSecret == "" {
		return nil, fmt.Errorf("credentials must contain both client_id and client_secret")
	}
	fillDefaults(creds)
	return creds, nil
}

func resolveRawCredentials(raw *gcpRawConfig) *GCPCredentials {
	if raw.Installed != nil {
		return raw.Installed
	}
	if raw.Web != nil {
		return raw.Web
	}
	return &GCPCredentials{
		ClientID:     raw.ClientID,
		ClientSecret: raw.Secret,
		ProjectID:    raw.ProjectID,
	}
}

func fillDefaults(c *GCPCredentials) {
	if c.AuthURI == "" {
		c.AuthURI = "https://accounts.google.com/o/oauth2/auth"
	}
	if c.TokenURI == "" {
		c.TokenURI = "https://oauth2.googleapis.com/token"
	}
	if len(c.RedirectURIs) == 0 {
		c.RedirectURIs = []string{"http://localhost"}
	}
}

// SaveGCPCredentials writes normalized credentials to credentials.json with 0600 mode.
func SaveGCPCredentials(creds *GCPCredentials) error {
	path, err := CredentialsPath()
	if err != nil {
		return err
	}
	fillDefaults(creds)
	payload := map[string]*GCPCredentials{
		"installed": creds,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write credentials to %s: %w", path, err)
	}
	return nil
}

// LoadGCPCredentials loads configured GCP credentials.
func LoadGCPCredentials() (*GCPCredentials, error) {
	path, err := CredentialsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseGCPCredentials(data)
}

// MaskSecret masks a secret string showing only the prefix and suffix.
func MaskSecret(secret string) string {
	if len(secret) <= 8 {
		return "********"
	}
	prefix := secret[:7]
	suffix := secret[len(secret)-4:]
	return fmt.Sprintf("%s...%s", prefix, suffix)
}

// MaskClientID masks a client ID showing prefix and domain.
func MaskClientID(clientID string) string {
	parts := strings.Split(clientID, ".")
	if len(parts) > 1 {
		prefix := clientID
		if len(prefix) > 12 {
			prefix = clientID[:12] + "..."
		}
		return fmt.Sprintf("%s.%s", prefix, parts[len(parts)-1])
	}
	if len(clientID) > 16 {
		return clientID[:12] + "..."
	}
	return clientID
}
