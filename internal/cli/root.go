package cli

import (
	"fmt"
	"os"

	"github.com/erikkubica/gws/internal/auth"
	"github.com/spf13/cobra"
)

// Version is the build version of gws (injected via ldflags at build time).
var Version = "1.1.0"

var accountFlag string

var rootCmd = &cobra.Command{
	Use:          "gws",
	Short:        "Unified Google Workspace CLI & Model Context Protocol (MCP) Server",
	Version:      Version,
	SilenceUsage: true,
	Long: `gws is an all-in-one developer tool and MCP server for Google Workspace.
It provides instant terminal access and AI agent integration for Gmail, Calendar, Meet, Drive, Tasks, Sheets, Docs, Chat, and YouTube.`,
}

// Execute is the main entrypoint for the CLI application.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&accountFlag, "account", "a", "", "Google account email to use for this command")
	cobra.OnInitialize(func() {
		if accountFlag != "" {
			auth.SelectedAccount = accountFlag
		} else if envAcc := os.Getenv("GWS_ACCOUNT"); envAcc != "" {
			auth.SelectedAccount = envAcc
		}
	})
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(gcpCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(mailCmd)
	rootCmd.AddCommand(calCmd)
	rootCmd.AddCommand(driveCmd)
	rootCmd.AddCommand(ytCmd)
	rootCmd.AddCommand(tasksCmd)
	rootCmd.AddCommand(sheetsCmd)
	rootCmd.AddCommand(docsCmd)
	rootCmd.AddCommand(meetCmd)
	rootCmd.AddCommand(chatCmd)
}

func printError(err error) {
	fmt.Printf("Error: %v\n", err)
}
