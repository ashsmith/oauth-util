package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version will be set during build time via ldflags
var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "oauth-util",
	Short: "CLI tool to get JWT tokens via OAuth2 flow",
	Long: `A fast and efficient CLI tool to obtain JWT tokens via OAuth2 flow
with support for multiple providers.`,
	Version: version,
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
