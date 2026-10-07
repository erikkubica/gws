package cli

import (
	"fmt"
	"os"

	"github.com/erikkubica/gws/internal/auth"
	"github.com/spf13/cobra"
)

var gcpCmd = &cobra.Command{
	Use:   "gcp",
	Short: "Manage Google Cloud Platform (GCP) OAuth application credentials",
}

var gcpImportCmd = &cobra.Command{
	Use:   "import <path/to/credentials.json>",
	Short: "Import Google Cloud OAuth client credentials from a JSON file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read file %s: %w", filePath, err)
		}
		creds, err := auth.ParseGCPCredentials(data)
		if err != nil {
			return fmt.Errorf("parse credentials: %w", err)
		}
		if err := auth.SaveGCPCredentials(creds); err != nil {
			return err
		}
		path, _ := auth.CredentialsPath()
		fmt.Printf("✔ Successfully imported GCP OAuth credentials.\n")
		fmt.Printf("  Client ID:   %s\n", auth.MaskClientID(creds.ClientID))
		fmt.Printf("  Saved to:    %s (mode 0600)\n", path)
		return nil
	},
}

var (
	flagClientID     string
	flagClientSecret string
	flagProjectID    string
)

var gcpSetCmd = &cobra.Command{
	Use:   "set [client_id] [client_secret]",
	Short: "Configure GCP OAuth client ID and secret directly",
	RunE: func(cmd *cobra.Command, args []string) error {
		clientID, clientSecret := resolveClientArgs(args)
		if clientID == "" || clientSecret == "" {
			return fmt.Errorf("both client_id and client_secret are required (positional or via flags)")
		}
		creds := &auth.GCPCredentials{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			ProjectID:    flagProjectID,
		}
		if err := auth.SaveGCPCredentials(creds); err != nil {
			return err
		}
		path, _ := auth.CredentialsPath()
		fmt.Printf("✔ Successfully configured GCP OAuth credentials.\n")
		fmt.Printf("  Client ID:   %s\n", auth.MaskClientID(clientID))
		fmt.Printf("  Saved to:    %s (mode 0600)\n", path)
		return nil
	},
}

func resolveClientArgs(args []string) (string, string) {
	id := flagClientID
	secret := flagClientSecret
	if len(args) >= 1 && id == "" {
		id = args[0]
	}
	if len(args) >= 2 && secret == "" {
		secret = args[1]
	}
	return id, secret
}

var gcpStatusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"show"},
	Short:   "Display configured GCP OAuth application details (aliases: show)",
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := auth.LoadGCPCredentials()
		if err != nil {
			fmt.Println("No GCP OAuth credentials configured.")
			fmt.Println("Run 'gws gcp import <file>' or 'gws gcp set <client_id> <client_secret>' to set up.")
			return nil
		}
		path, _ := auth.CredentialsPath()
		fmt.Println("GCP OAuth Application Configuration:")
		fmt.Printf("  Client ID:     %s\n", creds.ClientID)
		fmt.Printf("  Client Secret: %s\n", auth.MaskSecret(creds.ClientSecret))
		if creds.ProjectID != "" {
			fmt.Printf("  Project ID:    %s\n", creds.ProjectID)
		}
		fmt.Printf("  File:          %s (mode 0600)\n", path)
		return nil
	},
}

func init() {
	gcpSetCmd.Flags().StringVar(&flagClientID, "client-id", "", "Google OAuth Client ID")
	gcpSetCmd.Flags().StringVar(&flagClientSecret, "client-secret", "", "Google OAuth Client Secret")
	gcpSetCmd.Flags().StringVar(&flagProjectID, "project-id", "", "Google Cloud Project ID")

	gcpCmd.AddCommand(gcpImportCmd)
	gcpCmd.AddCommand(gcpSetCmd)
	gcpCmd.AddCommand(gcpStatusCmd)
}
