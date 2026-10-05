package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gmcp",
	Short: "Unified Google Workspace CLI & Model Context Protocol (MCP) Server",
	Long: `gmcp is an all-in-one developer tool and MCP server for Google Workspace.
It provides instant terminal access and AI agent integration for Gmail, Calendar, Drive, and YouTube.`,
}

// Execute is the main entrypoint for the CLI application.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(mailCmd)
	rootCmd.AddCommand(calCmd)
	rootCmd.AddCommand(driveCmd)
	rootCmd.AddCommand(ytCmd)
	rootCmd.AddCommand(tasksCmd)
	rootCmd.AddCommand(sheetsCmd)
	rootCmd.AddCommand(docsCmd)
	rootCmd.AddCommand(meetCmd)
}

func printError(err error) {
	fmt.Printf("Error: %v\n", err)
}
