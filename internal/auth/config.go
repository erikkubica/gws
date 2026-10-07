package auth

import (
	"fmt"
	"os"
	"path/filepath"
)

// ConfigDir returns the default directory where gws stores credentials and tokens.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}
	dir := filepath.Join(home, ".config", "gws")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("create config directory %s: %w", dir, err)
	}
	return dir, nil
}

// CredentialsPath returns the path to credentials.json.
func CredentialsPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(dir, "credentials.json")
	if _, err := os.Stat(p); err == nil {
		return p, nil
	}
	legacy := filepath.Join(filepath.Dir(dir), "gmcp", "credentials.json")
	if _, err := os.Stat(legacy); err == nil {
		return legacy, nil
	}
	return p, nil
}

// TokenPath returns the path to token.json.
func TokenPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(dir, "token.json")
	if _, err := os.Stat(p); err == nil {
		return p, nil
	}
	legacy := filepath.Join(filepath.Dir(dir), "gmcp", "token.json")
	if _, err := os.Stat(legacy); err == nil {
		return legacy, nil
	}
	return p, nil
}

// AccountsDir returns the directory where per-account data and tokens are stored.
func AccountsDir() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	accDir := filepath.Join(dir, "accounts")
	if err := os.MkdirAll(accDir, 0700); err != nil {
		return "", fmt.Errorf("create accounts directory %s: %w", accDir, err)
	}
	return accDir, nil
}

// AppConfigPath returns the path to gws configuration file (config.json).
func AppConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}
