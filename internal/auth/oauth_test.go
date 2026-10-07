package auth

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

type mockTokenSource struct {
	mu     sync.Mutex
	tokens []*oauth2.Token
	idx    int
}

func (m *mockTokenSource) Token() (*oauth2.Token, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.idx >= len(m.tokens) {
		return m.tokens[len(m.tokens)-1], nil
	}
	t := m.tokens[m.idx]
	m.idx++
	return t, nil
}

func TestPersistingTokenSource_RefreshPersistsAccountToken(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	account := "persist-test@example.com"
	tok1 := &oauth2.Token{
		AccessToken:  "initial-token",
		RefreshToken: "refresh-token-1",
		Expiry:       time.Now().Add(1 * time.Hour),
	}
	if err := SaveAccountToken(account, tok1); err != nil {
		t.Fatalf("failed to seed initial token: %v", err)
	}

	tok2 := &oauth2.Token{
		AccessToken:  "refreshed-token",
		RefreshToken: "refresh-token-2",
		Expiry:       time.Now().Add(2 * time.Hour),
	}

	mockSource := &mockTokenSource{
		tokens: []*oauth2.Token{tok1, tok2},
	}

	pts := &persistingTokenSource{
		source:  mockSource,
		account: account,
		last:    tok1,
	}

	// First call returns initial token; last matches so no file write
	gotTok, err := pts.Token()
	if err != nil {
		t.Fatalf("unexpected error on first Token(): %v", err)
	}
	if gotTok.AccessToken != "initial-token" {
		t.Fatalf("expected initial-token, got %s", gotTok.AccessToken)
	}

	// Second call returns refreshed token; pts should persist it
	refreshedTok, err := pts.Token()
	if err != nil {
		t.Fatalf("unexpected error on refreshed Token(): %v", err)
	}
	if refreshedTok.AccessToken != "refreshed-token" {
		t.Fatalf("expected refreshed-token, got %s", refreshedTok.AccessToken)
	}

	// Verify disk file has refreshed token
	loaded, err := LoadAccountToken(account)
	if err != nil {
		t.Fatalf("failed to load token from disk: %v", err)
	}
	if loaded.AccessToken != "refreshed-token" {
		t.Fatalf("expected disk token to be 'refreshed-token', got %q", loaded.AccessToken)
	}
}

func TestPersistingTokenSource_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	account := "concurrent@example.com"
	initialTok := &oauth2.Token{
		AccessToken: "concurrent-initial",
		Expiry:      time.Now().Add(1 * time.Hour),
	}
	if err := SaveAccountToken(account, initialTok); err != nil {
		t.Fatalf("failed to seed token: %v", err)
	}

	mockSource := &mockTokenSource{
		tokens: []*oauth2.Token{
			initialTok,
			{AccessToken: "concurrent-refreshed", Expiry: time.Now().Add(2 * time.Hour)},
		},
	}

	pts := &persistingTokenSource{
		source:  mockSource,
		account: account,
		last:    initialTok,
	}

	var wg sync.WaitGroup
	workers := 10
	errChan := make(chan error, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				tok, err := pts.Token()
				if err != nil {
					errChan <- fmt.Errorf("worker %d: %w", workerID, err)
					return
				}
				if tok == nil || tok.AccessToken == "" {
					errChan <- fmt.Errorf("worker %d got empty token", workerID)
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Fatal(err)
	}
}

func TestPersistingTokenSource_LegacyTokenPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	gwsDir := filepath.Join(tmpDir, ".config", "gws")
	if err := os.MkdirAll(gwsDir, 0700); err != nil {
		t.Fatalf("failed to create gws dir: %v", err)
	}

	tok1 := &oauth2.Token{AccessToken: "leg-1", Expiry: time.Now().Add(1 * time.Hour)}
	tok2 := &oauth2.Token{AccessToken: "leg-2", Expiry: time.Now().Add(2 * time.Hour)}

	pts := &persistingTokenSource{
		source:  &mockTokenSource{tokens: []*oauth2.Token{tok1, tok2}},
		account: "default",
		last:    tok1,
	}

	tok, err := pts.Token()
	if err != nil || tok.AccessToken != "leg-1" {
		t.Fatalf("unexpected tok: %v, err: %v", tok, err)
	}

	// Refreshed token persistence
	refreshed, err := pts.Token()
	if err != nil || refreshed.AccessToken != "leg-2" {
		t.Fatalf("unexpected refreshed tok: %v, err: %v", refreshed, err)
	}
}
