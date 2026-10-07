package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// AppConfig represents global user configuration for gws.
type AppConfig struct {
	ActiveAccount string `json:"active_account"`
}

// AccountInfo summarizes an authenticated account's status.
type AccountInfo struct {
	Email  string    `json:"email"`
	Active bool      `json:"active"`
	Valid  bool      `json:"valid"`
	Expiry time.Time `json:"expiry"`
}

// SanitizeAccount validates and normalizes an account identifier.
func SanitizeAccount(email string) (string, error) {
	clean := strings.TrimSpace(strings.ToLower(email))
	if clean == "" {
		return "", fmt.Errorf("account identifier cannot be empty")
	}
	if strings.ContainsAny(clean, "/\\:*?\"<>|") || strings.Contains(clean, "..") {
		return "", fmt.Errorf("invalid account identifier: %s", email)
	}
	return clean, nil
}

// LoadAppConfig reads config.json from disk.
func LoadAppConfig() (*AppConfig, error) {
	p, err := AppConfigPath()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &AppConfig{}, nil
		}
		return nil, fmt.Errorf("open app config: %w", err)
	}
	defer f.Close()
	var cfg AppConfig
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode app config: %w", err)
	}
	return &cfg, nil
}

// SaveAppConfig persists config.json to disk.
func SaveAppConfig(cfg *AppConfig) error {
	p, err := AppConfigPath()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("open app config: %w", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(cfg)
}

// GetActiveAccount returns the email of the active account.
func GetActiveAccount() (string, error) {
	cfg, err := LoadAppConfig()
	if err != nil {
		return "", err
	}
	return cfg.ActiveAccount, nil
}

// SetActiveAccount updates the active account in config.json.
func SetActiveAccount(email string) error {
	sanitized, err := SanitizeAccount(email)
	if err != nil {
		return err
	}
	tokPath, err := AccountTokenPath(sanitized)
	if err != nil {
		return err
	}
	if _, err := os.Stat(tokPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("account %q is not authenticated: run 'gws auth login' first", email)
		}
		return err
	}
	return SaveAppConfig(&AppConfig{ActiveAccount: sanitized})
}

// AccountDir returns the base directory for a given account.
func AccountDir(email string) (string, error) {
	sanitized, err := SanitizeAccount(email)
	if err != nil {
		return "", err
	}
	base, err := AccountsDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, sanitized)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("create account directory %s: %w", dir, err)
	}
	return dir, nil
}

// AccountTokenPath returns the path to token.json for a specific account.
func AccountTokenPath(email string) (string, error) {
	dir, err := AccountDir(email)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "token.json"), nil
}

// SaveAccountToken writes the OAuth2 token into the account's directory and sets it as active.
func SaveAccountToken(email string, token *oauth2.Token) error {
	tokPath, err := AccountTokenPath(email)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(tokPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("open account token file: %w", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(token); err != nil {
		return fmt.Errorf("encode account token: %w", err)
	}
	sanitized, _ := SanitizeAccount(email)
	return SaveAppConfig(&AppConfig{ActiveAccount: sanitized})
}

// LoadAccountToken reads token.json for a specific account.
func LoadAccountToken(email string) (*oauth2.Token, error) {
	tokPath, err := AccountTokenPath(email)
	if err != nil {
		return nil, err
	}
	return readTokenFromFile(tokPath)
}

func readTokenFromFile(path string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	if err := json.NewDecoder(f).Decode(tok); err != nil {
		return nil, fmt.Errorf("decode token: %w", err)
	}
	return tok, nil
}

// ListAccounts enumerates all authenticated accounts with active and validity status.
func ListAccounts() ([]AccountInfo, error) {
	_ = MigrateLegacyToken()
	base, err := AccountsDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, err
	}
	active, _ := GetActiveAccount()
	var list []AccountInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		info, err := loadAccountInfo(entry.Name(), active)
		if err == nil {
			list = append(list, info)
		}
	}
	return list, nil
}

func loadAccountInfo(email, activeEmail string) (AccountInfo, error) {
	tok, err := LoadAccountToken(email)
	if err != nil {
		return AccountInfo{}, err
	}
	return AccountInfo{
		Email:  email,
		Active: strings.EqualFold(email, activeEmail),
		Valid:  tok.Valid(),
		Expiry: tok.Expiry,
	}, nil
}

// DeleteAccount removes an authenticated account and updates the active account if needed.
func DeleteAccount(email string) error {
	sanitized, err := SanitizeAccount(email)
	if err != nil {
		return err
	}
	dir, err := AccountDir(sanitized)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("remove account directory: %w", err)
	}
	active, _ := GetActiveAccount()
	if strings.EqualFold(active, sanitized) {
		return selectNextActiveAccount()
	}
	return nil
}

func selectNextActiveAccount() error {
	accounts, err := ListAccounts()
	if err != nil || len(accounts) == 0 {
		return SaveAppConfig(&AppConfig{ActiveAccount: ""})
	}
	return SaveAppConfig(&AppConfig{ActiveAccount: accounts[0].Email})
}

// MigrateLegacyToken automatically moves a legacy single token into accounts/.
func MigrateLegacyToken() error {
	legacyPath, err := TokenPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(legacyPath); err != nil {
		return nil
	}
	base, err := AccountsDir()
	if err != nil {
		return err
	}
	entries, _ := os.ReadDir(base)
	if len(entries) > 0 {
		return nil
	}
	tok, err := readTokenFromFile(legacyPath)
	if err != nil {
		return nil
	}
	return migrateTokenToAccounts(tok, legacyPath)
}

func migrateTokenToAccounts(tok *oauth2.Token, legacyPath string) error {
	email := "default"
	cfg, err := LoadOAuthConfig("")
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if fetched, err := FetchUserEmail(ctx, cfg.Client(ctx, tok)); err == nil && fetched != "" {
			email = fetched
		}
	}
	if err := SaveAccountToken(email, tok); err != nil {
		return err
	}
	_ = os.Rename(legacyPath, legacyPath+".migrated")
	return nil
}
