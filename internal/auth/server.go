package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"golang.org/x/oauth2"
)

// openBrowser attempts to open the default system browser to the target URL.
func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return fmt.Errorf("unsupported platform for auto-open")
	}
}

// writeAuthResponse sends a pleasant HTML confirmation to the user browser.
func writeAuthResponse(w http.ResponseWriter, title, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Connection", "close")
	fmt.Fprintf(w, `<!DOCTYPE html><html><body style="font-family:sans-serif;text-align:center;padding:50px;">
		<h1 style="color:#10b981;">%s</h1>
		<p>%s</p>
	</body></html>`, title, message)
}

// LoginFlow starts a local webserver and conducts the browser OAuth flow.
func LoginFlow(ctx context.Context) error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("bind loopback listener: %w", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	redirectURL := fmt.Sprintf("http://localhost:%d", port)

	cfg, err := LoadOAuthConfig(redirectURL)
	if err != nil {
		return err
	}

	authURL := cfg.AuthCodeURL("state-gmcp", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Printf("Opening browser for Google Authentication...\nIf browser does not open, visit:\n%s\n\n", authURL)
	_ = openBrowser(authURL)

	return handleCallback(ctx, listener, cfg)
}

// handleCallback awaits the HTTP authorization response and exchanges code for token.
func handleCallback(ctx context.Context, l net.Listener, cfg *oauth2.Config) error {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if code := r.URL.Query().Get("code"); code != "" {
				writeAuthResponse(w, "Authentication Successful!", "You can close this tab and return to the terminal.")
				codeCh <- code
				return
			}
			errStr := r.URL.Query().Get("error")
			writeAuthResponse(w, "Authentication Failed", "Error: "+errStr)
			errCh <- fmt.Errorf("oauth failed: %s", errStr)
		}),
	}

	go func() {
		if err := server.Serve(l); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case code := <-codeCh:
		tok, err := cfg.Exchange(ctx, code)
		if err != nil {
			return fmt.Errorf("exchange auth code: %w", err)
		}
		if err := SaveToken(tok); err != nil {
			return fmt.Errorf("save token: %w", err)
		}
		go func() {
			time.Sleep(100 * time.Millisecond)
			_ = server.Close()
		}()
		return nil
	case err := <-errCh:
		return err
	case <-time.After(3 * time.Minute):
		return fmt.Errorf("authentication timed out waiting for browser callback")
	}
}
