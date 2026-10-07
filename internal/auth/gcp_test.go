package auth

import (
	"testing"
)

func TestParseGCPCredentials(t *testing.T) {
	t.Run("nominal_installed_format", func(t *testing.T) {
		json := []byte(`{
			"installed": {
				"client_id": "123456.apps.googleusercontent.com",
				"client_secret": "GOCSPX-secret123",
				"project_id": "test-project"
			}
		}`)
		creds, err := ParseGCPCredentials(json)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if creds.ClientID != "123456.apps.googleusercontent.com" {
			t.Errorf("expected client_id %s, got %s", "123456.apps.googleusercontent.com", creds.ClientID)
		}
		if creds.ClientSecret != "GOCSPX-secret123" {
			t.Errorf("expected client_secret %s, got %s", "GOCSPX-secret123", creds.ClientSecret)
		}
		if creds.ProjectID != "test-project" {
			t.Errorf("expected project_id %s, got %s", "test-project", creds.ProjectID)
		}
	})

	t.Run("nominal_web_format", func(t *testing.T) {
		json := []byte(`{
			"web": {
				"client_id": "web-client.apps.googleusercontent.com",
				"client_secret": "GOCSPX-websecret"
			}
		}`)
		creds, err := ParseGCPCredentials(json)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if creds.ClientID != "web-client.apps.googleusercontent.com" {
			t.Errorf("expected client_id %s, got %s", "web-client.apps.googleusercontent.com", creds.ClientID)
		}
	})

	t.Run("malformed_missing_client_secret", func(t *testing.T) {
		json := []byte(`{"installed": {"client_id": "id-only"}}`)
		_, err := ParseGCPCredentials(json)
		if err == nil {
			t.Fatal("expected error for missing client_secret, got nil")
		}
	})

	t.Run("malformed_missing_client_id", func(t *testing.T) {
		json := []byte(`{"installed": {"client_secret": "secret-only"}}`)
		_, err := ParseGCPCredentials(json)
		if err == nil {
			t.Fatal("expected error for missing client_id, got nil")
		}
	})

	t.Run("malformed_empty_json", func(t *testing.T) {
		_, err := ParseGCPCredentials([]byte(`{}`))
		if err == nil {
			t.Fatal("expected error for empty json, got nil")
		}
	})

	t.Run("malformed_invalid_syntax", func(t *testing.T) {
		_, err := ParseGCPCredentials([]byte(`not a json`))
		if err == nil {
			t.Fatal("expected error for invalid syntax, got nil")
		}
	})
}

func TestMaskSecret(t *testing.T) {
	t.Run("nominal_secret", func(t *testing.T) {
		masked := MaskSecret("GOCSPX-abcdefghijkl1234")
		if masked != "GOCSPX-...1234" {
			t.Errorf("unexpected masked output: %s", masked)
		}
	})

	t.Run("short_secret_boundary", func(t *testing.T) {
		masked := MaskSecret("short")
		if masked != "********" {
			t.Errorf("unexpected masked output: %s", masked)
		}
	})

	t.Run("empty_string_boundary", func(t *testing.T) {
		masked := MaskSecret("")
		if masked != "********" {
			t.Errorf("unexpected masked output: %s", masked)
		}
	})
}
