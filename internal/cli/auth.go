package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/erikkubica/gws/internal/auth"
	"github.com/spf13/cobra"
)

var logoutAllFlag bool

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Google OAuth2 authentication, accounts, and credentials",
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in with your Google account via browser",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		email, err := auth.LoginFlow(ctx)
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}
		fmt.Printf("Successfully authenticated as %s!\n", email)
		fmt.Printf("Active account set to: %s\n", email)
		return nil
	},
}

var listAccountsCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"accounts", "ls"},
	Short:   "List all authenticated Google accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		accounts, err := auth.ListAccounts()
		if err != nil {
			return fmt.Errorf("list accounts: %w", err)
		}
		if len(accounts) == 0 {
			fmt.Println("No accounts authenticated yet. Run 'gws auth login' to add an account.")
			return nil
		}
		fmt.Println("Authenticated Accounts:")
		for _, acc := range accounts {
			activeMarker := "  "
			if acc.Active {
				activeMarker = "* "
			}
			status := "valid"
			if !acc.Valid {
				status = "token expired, auto-refreshes on use"
			}
			fmt.Printf("%s%s (%s)\n", activeMarker, acc.Email, status)
		}
		return nil
	},
}

var switchAccountCmd = &cobra.Command{
	Use:   "switch [email]",
	Short: "Switch the active default Google account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		if err := auth.SetActiveAccount(target); err != nil {
			return err
		}
		fmt.Printf("Switched active account to: %s\n", target)
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
	accounts, err := auth.ListAccounts()
	if err != nil || len(accounts) == 0 {
		fmt.Println("User Account:    Not authenticated (run 'gws auth login')")
		return
	}
	active, _ := auth.GetActiveAccount()
	fmt.Printf("Active Account:  %s\n", active)
	fmt.Printf("Total Accounts:  %d authenticated (run 'gws auth list' to inspect)\n", len(accounts))
}

var logoutCmd = &cobra.Command{
	Use:   "logout [email]",
	Short: "Log out and remove saved authentication tokens",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if logoutAllFlag {
			return logoutAllAccounts()
		}
		target := ""
		if len(args) > 0 {
			target = args[0]
		}
		return logoutSingleAccount(target)
	},
}

func logoutAllAccounts() error {
	accounts, err := auth.ListAccounts()
	if err != nil {
		return err
	}
	for _, acc := range accounts {
		_ = auth.DeleteAccount(acc.Email)
	}
	tokPath, _ := auth.TokenPath()
	_ = os.Remove(tokPath)
	_ = auth.SaveAppConfig(&auth.AppConfig{ActiveAccount: ""})
	fmt.Println("Logged out of all accounts.")
	return nil
}

func logoutSingleAccount(target string) error {
	if target == "" {
		act, err := auth.GetActiveAccount()
		if err != nil || act == "" {
			fmt.Println("No active account to log out.")
			return nil
		}
		target = act
	}
	if err := auth.DeleteAccount(target); err != nil {
		return err
	}
	fmt.Printf("Logged out %s successfully.\n", target)
	if next, _ := auth.GetActiveAccount(); next != "" {
		fmt.Printf("Active account switched to: %s\n", next)
	}
	return nil
}

func init() {
	logoutCmd.Flags().BoolVarP(&logoutAllFlag, "all", "", false, "Log out and clear all accounts")

	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(listAccountsCmd)
	authCmd.AddCommand(switchAccountCmd)
	authCmd.AddCommand(statusCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(gcpImportCmd)
}
