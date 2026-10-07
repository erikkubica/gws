package auth

import (
	"strings"
	"testing"
)

func TestGenerateMCPToken(t *testing.T) {
	t.Run("nominal_token_generation", func(t *testing.T) {
		token, err := GenerateMCPToken()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(token, "gws_mcp_") {
			t.Errorf("token should start with gws_mcp_, got %s", token)
		}
		// 8 prefix chars + 64 hex chars = 72 chars
		if len(token) != 72 {
			t.Errorf("expected token length 72, got %d (%s)", len(token), token)
		}
	})

	t.Run("uniqueness_boundary", func(t *testing.T) {
		t1, _ := GenerateMCPToken()
		t2, _ := GenerateMCPToken()
		if t1 == t2 {
			t.Fatal("two consecutive tokens must not be identical")
		}
	})
}
