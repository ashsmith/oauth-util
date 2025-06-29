package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "oauth-util",
	Short: "CLI tool to get JWT tokens via OAuth2 flow",
	Long: `A fast and efficient CLI tool to obtain JWT tokens via OAuth2 flow
with support for multiple providers.`,
	Version: "1.0.0",
}

func init() {
	// Add commands
	rootCmd.AddCommand(configureCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(tokenCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(setDefaultCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(clearTokensCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
