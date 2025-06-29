package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/ohler55/ojg/jp"
	"github.com/spf13/cobra"
)

var (
	clientID   string
	domain     string
	scope      string
	port       string
	appName    string
	jsonOutput bool
	jsonPath   string
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure OAuth2 app settings",
	Run: func(cmd *cobra.Command, args []string) {
		name, appConfig, err := interactiveSetup()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		saveApp(name, appConfig)
		color.Green("✅ App configuration saved successfully!")
	},
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Start OAuth flow to get JWT token",
	Run: func(cmd *cobra.Command, args []string) {
		var appConfig AppConfig
		var currentAppName string
		var err error

		// Determine which app config to use
		if appName != "" {
			// Use specified app
			config, exists := getApp(appName)
			if !exists {
				fmt.Fprintf(os.Stderr, "Error: App '%s' not found. Run 'oauth-util configure' to set up apps.\n", appName)
				os.Exit(1)
			}
			appConfig = config
			currentAppName = appName
		} else if clientID != "" && domain != "" {
			// Use command line parameters
			appConfig = AppConfig{
				ClientID: clientID,
				Domain:   domain,
				Scope:    scope,
			}
			// Don't save tokens for command-line configs
		} else {
			// Use default app
			defaultAppName := getDefaultApp()
			if defaultAppName == "" {
				fmt.Fprintf(os.Stderr, "Error: No default app set. Use --app to specify an app or run 'oauth-util configure'.\n")
				os.Exit(1)
			}
			config, exists := getApp(defaultAppName)
			if !exists {
				fmt.Fprintf(os.Stderr, "Error: Default app not found. Run 'oauth-util configure' to set up apps.\n")
				os.Exit(1)
			}
			appConfig = config
			currentAppName = defaultAppName
		}

		// Start OAuth flow
		oauth := NewOAuthFlow()
		tokens, err := oauth.StartFlow(appConfig, port)
		if err != nil {
			if jsonOutput {
				errorResp := map[string]string{"error": err.Error()}
				json.NewEncoder(os.Stderr).Encode(errorResp)
			} else {
				fmt.Fprintf(os.Stderr, "❌ Error during OAuth flow: %v\n", err)
			}
			os.Exit(1)
		}

		// Save tokens if using a saved app configuration
		if currentAppName != "" {
			if err := saveTokensToApp(currentAppName, tokens); err != nil {
				if jsonOutput {
					errorResp := map[string]string{"error": fmt.Sprintf("Failed to save tokens: %v", err)}
					json.NewEncoder(os.Stderr).Encode(errorResp)
				} else {
					fmt.Fprintf(os.Stderr, "⚠️  Warning: Failed to save tokens: %v\n", err)
				}
			} else if !jsonOutput {
				color.Green("✅ Tokens saved to configuration for app '%s'", currentAppName)
			}
		}

		// Output tokens
		if jsonPath != "" {
			// Apply JSONPath filtering
			result, err := applyJSONPath(tokens, jsonPath)
			if err != nil {
				if jsonOutput {
					errorResp := map[string]string{"error": err.Error()}
					json.NewEncoder(os.Stderr).Encode(errorResp)
				} else {
					fmt.Fprintf(os.Stderr, "❌ JSONPath error: %v\n", err)
				}
				os.Exit(1)
			}

			if jsonOutput {
				json.NewEncoder(os.Stdout).Encode(result)
			} else {
				// For non-JSON output, try to format nicely
				if str, ok := result.(string); ok {
					fmt.Println(str)
				} else {
					output, _ := json.MarshalIndent(result, "", "  ")
					fmt.Println(string(output))
				}
			}
		} else if jsonOutput {
			json.NewEncoder(os.Stdout).Encode(tokens)
		} else {
			output, _ := json.MarshalIndent(tokens, "", "  ")
			fmt.Println(string(output))
		}
	},
}

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Get JWT token using saved app configuration",
	Run: func(cmd *cobra.Command, args []string) {
		var appConfig AppConfig
		var currentAppName string
		var err error

		// Determine which app config to use
		if appName != "" {
			// Use specified app
			config, exists := getApp(appName)
			if !exists {
				if jsonOutput {
					errorResp := map[string]string{"error": fmt.Sprintf("App '%s' not found.", appName)}
					json.NewEncoder(os.Stderr).Encode(errorResp)
				} else {
					fmt.Fprintf(os.Stderr, "❌ Error: App '%s' not found.\n", appName)
				}
				os.Exit(1)
			}
			appConfig = config
			currentAppName = appName
		} else {
			// Use default app
			defaultAppName := getDefaultApp()
			if defaultAppName == "" {
				if jsonOutput {
					errorResp := map[string]string{"error": "No default app set. Use --app to specify an app or run 'oauth-util configure'."}
					json.NewEncoder(os.Stderr).Encode(errorResp)
				} else {
					fmt.Fprintf(os.Stderr, "❌ Error: No default app set. Use --app to specify an app or run 'oauth-util configure'.\n")
				}
				os.Exit(1)
			}
			config, exists := getApp(defaultAppName)
			if !exists {
				if jsonOutput {
					errorResp := map[string]string{"error": "Default app not found. Run 'oauth-util configure' to set up apps."}
					json.NewEncoder(os.Stderr).Encode(errorResp)
				} else {
					fmt.Fprintf(os.Stderr, "❌ Error: Default app not found. Run 'oauth-util configure' to set up apps.\n")
				}
				os.Exit(1)
			}
			appConfig = config
			currentAppName = defaultAppName
		}

		// First, check if we have a valid stored token
		if storedToken, err := getStoredToken(currentAppName); err == nil {
			if jsonPath != "" {
				// Apply JSONPath filtering
				result, err := applyJSONPath(storedToken, jsonPath)
				if err != nil {
					if jsonOutput {
						errorResp := map[string]string{"error": err.Error()}
						json.NewEncoder(os.Stderr).Encode(errorResp)
					} else {
						fmt.Fprintf(os.Stderr, "❌ JSONPath error: %v\n", err)
					}
					os.Exit(1)
				}

				if jsonOutput {
					json.NewEncoder(os.Stdout).Encode(result)
				} else {
					// For non-JSON output, try to format nicely
					if str, ok := result.(string); ok {
						fmt.Println(str)
					} else {
						output, _ := json.MarshalIndent(result, "", "  ")
						fmt.Println(string(output))
					}
				}
			} else if jsonOutput {
				json.NewEncoder(os.Stdout).Encode(storedToken)
			} else {
				output, _ := json.MarshalIndent(storedToken, "", "  ")
				fmt.Println(string(output))
			}
			return
		} else if !jsonOutput {
			fmt.Printf("ℹ️  No valid stored token found: %v\n", err)
			fmt.Println("🔄 Starting new OAuth flow...")
		}

		// Start OAuth flow
		oauth := NewOAuthFlow()
		tokens, err := oauth.StartFlow(appConfig, port)
		if err != nil {
			if jsonOutput {
				errorResp := map[string]string{"error": err.Error()}
				json.NewEncoder(os.Stderr).Encode(errorResp)
			} else {
				fmt.Fprintf(os.Stderr, "❌ Error during OAuth flow: %v\n", err)
			}
			os.Exit(1)
		}

		// Save tokens to configuration
		if err := saveTokensToApp(currentAppName, tokens); err != nil {
			if jsonOutput {
				errorResp := map[string]string{"error": fmt.Sprintf("Failed to save tokens: %v", err)}
				json.NewEncoder(os.Stderr).Encode(errorResp)
			} else {
				fmt.Fprintf(os.Stderr, "⚠️  Warning: Failed to save tokens: %v\n", err)
			}
		} else if !jsonOutput {
			color.Green("✅ Tokens saved to configuration for app '%s'", currentAppName)
		}

		// Output tokens
		if jsonPath != "" {
			// Apply JSONPath filtering
			result, err := applyJSONPath(tokens, jsonPath)
			if err != nil {
				if jsonOutput {
					errorResp := map[string]string{"error": err.Error()}
					json.NewEncoder(os.Stderr).Encode(errorResp)
				} else {
					fmt.Fprintf(os.Stderr, "❌ JSONPath error: %v\n", err)
				}
				os.Exit(1)
			}

			if jsonOutput {
				json.NewEncoder(os.Stdout).Encode(result)
			} else {
				// For non-JSON output, try to format nicely
				if str, ok := result.(string); ok {
					fmt.Println(str)
				} else {
					output, _ := json.MarshalIndent(result, "", "  ")
					fmt.Println(string(output))
				}
			}
		} else if jsonOutput {
			json.NewEncoder(os.Stdout).Encode(tokens)
		} else {
			output, _ := json.MarshalIndent(tokens, "", "  ")
			fmt.Println(string(output))
		}
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured apps",
	Run: func(cmd *cobra.Command, args []string) {
		listApps()
	},
}

var setDefaultCmd = &cobra.Command{
	Use:   "set-default [appName]",
	Short: "Set default app",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		_, exists := getApp(appName)
		if !exists {
			fmt.Fprintf(os.Stderr, "❌ Error: App '%s' not found.\n", appName)
			os.Exit(1)
		}

		setDefaultApp(appName)
		color.Green("✅ '%s' set as default app.", appName)
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [appName]",
	Short: "Delete an app configuration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		_, exists := getApp(appName)
		if !exists {
			fmt.Fprintf(os.Stderr, "❌ Error: App '%s' not found.\n", appName)
			os.Exit(1)
		}

		deleteApp(appName)
		color.Green("✅ App '%s' deleted successfully.", appName)
	},
}

var clearTokensCmd = &cobra.Command{
	Use:   "clear-tokens [appName]",
	Short: "Clear stored tokens for an app",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		app, exists := getApp(appName)
		if !exists {
			fmt.Fprintf(os.Stderr, "❌ Error: App '%s' not found.\n", appName)
			os.Exit(1)
		}

		// Clear token information
		app.AccessToken = ""
		app.RefreshToken = ""
		app.IdToken = ""
		app.TokenType = ""
		app.ExpiresAt = ""
		app.ExpiresIn = 0

		// Save updated config
		config.Apps[appName] = app
		if err := saveConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error saving config: %v\n", err)
			os.Exit(1)
		}

		color.Green("✅ Tokens cleared for app '%s'", appName)
	},
}

// applyJSONPath applies JSONPath filtering to a token response
func applyJSONPath(tokens *TokenResponse, jsonPathExpr string) (interface{}, error) {
	// Convert tokens to JSON
	tokenJSON, err := json.Marshal(tokens)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tokens: %v", err)
	}

	// Parse JSONPath expression
	expr, err := jp.ParseString(jsonPathExpr)
	if err != nil {
		return nil, fmt.Errorf("invalid JSONPath expression: %v", err)
	}

	// Parse JSON data
	var data interface{}
	if err := json.Unmarshal(tokenJSON, &data); err != nil {
		return nil, fmt.Errorf("failed to parse token data: %v", err)
	}

	// Apply JSONPath
	result := expr.Get(data)

	// Handle single result vs multiple results
	if len(result) == 0 {
		return nil, fmt.Errorf("JSONPath expression returned no results")
	} else if len(result) == 1 {
		return result[0], nil
	} else {
		return result, nil
	}
}

func init() {
	// Login command flags
	loginCmd.Flags().StringVarP(&clientID, "client-id", "c", "", "OAuth2 Client ID")
	loginCmd.Flags().StringVarP(&domain, "domain", "d", "", "OAuth2 Domain (full URL)")
	loginCmd.Flags().StringVarP(&scope, "scope", "s", "openid email profile", "OAuth2 Scope")
	loginCmd.Flags().StringVarP(&port, "port", "p", "3000", "Local server port")
	loginCmd.Flags().StringVarP(&appName, "app", "a", "", "Use saved app configuration")
	loginCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output only JSON data (for piping to jq)")

	// Token command flags
	tokenCmd.Flags().StringVarP(&port, "port", "p", "3000", "Local server port")
	tokenCmd.Flags().StringVarP(&appName, "app", "a", "", "Use specific app (defaults to default app)")
	tokenCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output only JSON data (for piping to jq)")
	tokenCmd.Flags().StringVar(&jsonPath, "jsonpath", "", "JSONPath expression to filter token response")
}
