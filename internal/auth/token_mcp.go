package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MCPTokenPath returns the file path for the stored MCP token.
func MCPTokenPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "mcp_token"), nil
}

// GenerateMCPToken creates a cryptographically secure 256-bit token.
func GenerateMCPToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return "gws_mcp_" + hex.EncodeToString(bytes), nil
}

// SaveMCPToken stores the token in ~/.config/gws/mcp_token with 0600 mode.
func SaveMCPToken(token string) error {
	path, err := MCPTokenPath()
	if err != nil {
		return err
	}
	cleanToken := strings.TrimSpace(token)
	if err := os.WriteFile(path, []byte(cleanToken), 0600); err != nil {
		return fmt.Errorf("write mcp token to %s: %w", path, err)
	}
	return nil
}

// LoadMCPToken retrieves the MCP token from GWS_MCP_TOKEN env or file.
func LoadMCPToken() (string, error) {
	if envToken := strings.TrimSpace(os.Getenv("GWS_MCP_TOKEN")); envToken != "" {
		return envToken, nil
	}
	path, err := MCPTokenPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// RevokeMCPToken deletes the stored MCP token file.
func RevokeMCPToken() error {
	path, err := MCPTokenPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove mcp token: %w", err)
	}
	return nil
}
