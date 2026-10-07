package auth

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestSanitizeAccount(t *testing.T) {
	t.Run("nominal_cases", func(t *testing.T) {
		nominal := []struct {
			input    string
			expected string
		}{
			{"user@example.com", "user@example.com"},
			{"WORK@EXAMPLE.COM", "work@example.com"},
			{"  user@example.com  ", "user@example.com"},
			{"work-account", "work-account"},
		}
		for _, tc := range nominal {
			got, err := SanitizeAccount(tc.input)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.input, err)
			}
			if got != tc.expected {
				t.Errorf("SanitizeAccount(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		}
	})

	t.Run("boundary_thresholds", func(t *testing.T) {
		// Single character boundary
		got, err := SanitizeAccount("a")
		if err != nil || got != "a" {
			t.Errorf("expected 'a', got %q, err: %v", got, err)
		}
	})

	t.Run("malformed_inputs", func(t *testing.T) {
		malformed := []string{
			"",
			"   ",
			"../traversal",
			"user/path",
			"user\\backslash",
			"user:colons",
			"user*wildcard",
			"user?question",
		}
		for _, input := range malformed {
			_, err := SanitizeAccount(input)
			if err == nil {
				t.Errorf("expected error for malformed input %q, got nil", input)
			}
		}
	})
}

func TestMultiAccountLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	tok1 := &oauth2.Token{
		AccessToken:  "token-work",
		RefreshToken: "refresh-work",
		Expiry:       time.Now().Add(1 * time.Hour),
	}
	tok2 := &oauth2.Token{
		AccessToken:  "token-personal",
		RefreshToken: "refresh-personal",
		Expiry:       time.Now().Add(2 * time.Hour),
	}

	// 1. Save first account
	if err := SaveAccountToken("work@example.com", tok1); err != nil {
		t.Fatalf("failed to save tok1: %v", err)
	}

	active, err := GetActiveAccount()
	if err != nil || active != "work@example.com" {
		t.Fatalf("expected active account 'work@example.com', got %q, err: %v", active, err)
	}

	// 2. Save second account
	if err := SaveAccountToken("user@example.com", tok2); err != nil {
		t.Fatalf("failed to save tok2: %v", err)
	}

	// 3. List accounts
	accounts, err := ListAccounts()
	if err != nil {
		t.Fatalf("failed to list accounts: %v", err)
	}
	if len(accounts) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(accounts))
	}

	// 4. Switch active account back to work@example.com
	if err := SetActiveAccount("work@example.com"); err != nil {
		t.Fatalf("failed to set active account: %v", err)
	}
	active, _ = GetActiveAccount()
	if active != "work@example.com" {
		t.Fatalf("expected active 'work@example.com', got %q", active)
	}

	// 5. Test context resolution
	ctxOverride := WithAccount(context.Background(), "user@example.com")
	tokResolved, accResolved, err := ResolveToken(ctxOverride)
	if err != nil {
		t.Fatalf("failed to resolve token with context: %v", err)
	}
	if accResolved != "user@example.com" || tokResolved.AccessToken != "token-personal" {
		t.Errorf("expected token-personal for user@example.com, got %s / %s", accResolved, tokResolved.AccessToken)
	}

	// 6. Delete account and verify active auto-switches
	if err := DeleteAccount("work@example.com"); err != nil {
		t.Fatalf("failed to delete account: %v", err)
	}
	active, _ = GetActiveAccount()
	if active != "user@example.com" {
		t.Fatalf("expected active account to auto-switch to 'user@example.com', got %q", active)
	}
}

func TestLegacyTokenMigration(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	gwsDir := filepath.Join(tmpDir, ".config", "gws")
	if err := os.MkdirAll(gwsDir, 0700); err != nil {
		t.Fatalf("failed to create gws dir: %v", err)
	}

	tok := &oauth2.Token{
		AccessToken:  "legacy-access-token",
		RefreshToken: "legacy-refresh-token",
		Expiry:       time.Now().Add(1 * time.Hour),
	}
	b, _ := json.Marshal(tok)
	legacyPath := filepath.Join(gwsDir, "token.json")
	if err := os.WriteFile(legacyPath, b, 0600); err != nil {
		t.Fatalf("failed to write legacy token: %v", err)
	}

	// Run migration
	if err := MigrateLegacyToken(); err != nil {
		t.Fatalf("migration returned error: %v", err)
	}

	// Verify token can still be resolved
	tokLoaded, _, err := ResolveToken(context.Background())
	if err != nil {
		t.Fatalf("failed to resolve token post-migration: %v", err)
	}
	if tokLoaded.AccessToken != "legacy-access-token" {
		t.Errorf("expected 'legacy-access-token', got %q", tokLoaded.AccessToken)
	}
}
