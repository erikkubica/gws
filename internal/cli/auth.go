package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/erikkubica/gmcp/internal/auth"
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
		tok, err := auth.LoadToken()
		if err != nil {
			fmt.Println("Not authenticated. Run 'gmcp auth login' to authenticate.")
			return nil
		}
		fmt.Println("Authenticated.")
		fmt.Printf("Token Expiry: %v (Valid: %v)\n", tok.Expiry, tok.Valid())
		return nil
	},
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
}
