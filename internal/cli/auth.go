package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/erikkubica/gws/internal/auth"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Google OAuth2 authentication and credentials",
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in with your Google account via browser",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		if err := auth.LoginFlow(ctx); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}
		fmt.Println("Successfully authenticated! Token saved.")
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		printGCPAuthStatus()
		printUserAuthStatus()
		return nil
	},
}

func printGCPAuthStatus() {
	creds, err := auth.LoadGCPCredentials()
	if err != nil {
		fmt.Println("GCP Application: Not configured (run 'gws gcp import <file>' or 'gws gcp set <id> <secret>')")
		return
	}
	fmt.Printf("GCP Application: Configured (Client ID: %s)\n", auth.MaskClientID(creds.ClientID))
}

func printUserAuthStatus() {
	tok, err := auth.LoadToken()
	if err != nil {
		fmt.Println("User Account:    Not authenticated (run 'gws auth login')")
		return
	}
	if !tok.Valid() {
		fmt.Println("User Account:    Authenticated (access token expired, auto-refreshes on next command)")
		fmt.Printf("Token Expiry:    %v\n", tok.Expiry)
		return
	}
	fmt.Println("User Account:    Authenticated (Active)")
	fmt.Printf("Token Expiry:    %v\n", tok.Expiry)
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear saved authentication tokens",
	RunE: func(cmd *cobra.Command, args []string) error {
		tokPath, err := auth.TokenPath()
		if err != nil {
			return err
		}
		_ = os.Remove(tokPath)
		fmt.Println("Logged out successfully.")
		return nil
	},
}

func init() {
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(statusCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(gcpImportCmd)
}
