package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

type AppConfig struct {
	ClientID     string `json:"client_id" mapstructure:"client_id"`
	Domain       string `json:"domain" mapstructure:"domain"`
	Scope        string `json:"scope" mapstructure:"scope"`
	AccessToken  string `json:"access_token,omitempty" mapstructure:"access_token"`
	IdToken      string `json:"id_token,omitempty" mapstructure:"id_token"`
	RefreshToken string `json:"refresh_token,omitempty" mapstructure:"refresh_token"`
	TokenType    string `json:"token_type,omitempty" mapstructure:"token_type"`
	ExpiresIn    int    `json:"expires_in,omitempty" mapstructure:"expires_in"`
	ExpiresAt    string `json:"expires_at,omitempty" mapstructure:"expires_at"`
}

type Config struct {
	Apps       map[string]AppConfig `json:"apps" mapstructure:"apps"`
	DefaultApp string               `json:"default_app" mapstructure:"default_app"`
}

var config Config

func init() {
	// Create config directory if it doesn't exist
	configDir := filepath.Join(os.Getenv("HOME"), ".config")
	os.MkdirAll(configDir, 0755)

	// Set the config file path explicitly
	configPath := filepath.Join(configDir, "oauth-util.json")
	viper.SetConfigFile(configPath)

	// Load existing config or create new one
	if err := viper.ReadInConfig(); err != nil {
		// Check if it's a config file not found error
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file doesn't exist, create default config
			config = Config{
				Apps:       make(map[string]AppConfig),
				DefaultApp: "",
			}
			// Don't print error for first-time users
		} else {
			// For other errors, check if it's a "no such file or directory" error
			if os.IsNotExist(err) {
				// Config file doesn't exist, create default config
				config = Config{
					Apps:       make(map[string]AppConfig),
					DefaultApp: "",
				}
				// Don't print error for first-time users
			} else {
				// This is a real error, print it
				fmt.Printf("Error reading config: %v\n", err)
				os.Exit(1)
			}
		}
	} else {
		// Load existing config
		if err := viper.Unmarshal(&config); err != nil {
			fmt.Printf("Error unmarshaling config: %v\n", err)
			os.Exit(1)
		}

		// Ensure the map is not nil after unmarshaling
		if config.Apps == nil {
			config.Apps = make(map[string]AppConfig)
		}
	}
}

func saveConfig() error {
	viper.Set("apps", config.Apps)
	viper.Set("default_app", config.DefaultApp)
	return viper.WriteConfig()
}

func getApp(name string) (AppConfig, bool) {
	app, exists := config.Apps[name]
	return app, exists
}

func saveApp(name string, appConfig AppConfig) {
	config.Apps[name] = appConfig
	if err := saveConfig(); err != nil {
		fmt.Printf("Error saving app: %v\n", err)
		os.Exit(1)
	}
}

func deleteApp(name string) {
	delete(config.Apps, name)
	if config.DefaultApp == name {
		config.DefaultApp = ""
	}
	saveConfig()
}

func setDefaultApp(name string) {
	config.DefaultApp = name
	saveConfig()
}

func getDefaultApp() string {
	return config.DefaultApp
}

func interactiveSetup() (string, AppConfig, error) {
	prompt := promptui.Prompt{
		Label: "App name (for easy reference)",
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("app name is required")
			}
			return nil
		},
	}
	name, err := prompt.Run()
	if err != nil {
		return "", AppConfig{}, err
	}

	prompt = promptui.Prompt{
		Label: "OAuth2 Client ID",
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("client ID is required")
			}
			return nil
		},
	}
	clientID, err := prompt.Run()
	if err != nil {
		return "", AppConfig{}, err
	}

	prompt = promptui.Prompt{
		Label: "OAuth2 Domain (full URL, e.g., https://accounts.google.com)",
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("domain is required")
			}
			// Basic URL validation
			if !isValidURL(input) {
				return fmt.Errorf("please enter a valid URL")
			}
			return nil
		},
	}
	domain, err := prompt.Run()
	if err != nil {
		return "", AppConfig{}, err
	}

	prompt = promptui.Prompt{
		Label:   "OAuth2 Scope (default: openid email profile)",
		Default: "openid email profile",
	}
	scope, err := prompt.Run()
	if err != nil {
		return "", AppConfig{}, err
	}

	confirm := promptui.Prompt{
		Label:     "Set this as the default app",
		IsConfirm: true,
		Default:   "y",
	}
	setAsDefault, err := confirm.Run()
	if err != nil && err != promptui.ErrAbort {
		return "", AppConfig{}, err
	}

	appConfig := AppConfig{
		ClientID: clientID,
		Domain:   domain,
		Scope:    scope,
	}

	if setAsDefault == "y" || setAsDefault == "Y" {
		setDefaultApp(name)
	}

	return name, appConfig, nil
}

func listApps() {
	if len(config.Apps) == 0 {
		fmt.Println("📝 No apps configured yet.")
		return
	}

	fmt.Println("📱 Configured Apps:")
	for name, app := range config.Apps {
		isDefault := name == config.DefaultApp
		if isDefault {
			fmt.Printf("  %s (default)\n", name)
		} else {
			fmt.Printf("  %s\n", name)
		}
		fmt.Printf("    Domain: %s\n", app.Domain)
		fmt.Printf("    Client ID: %s\n", app.ClientID)
		fmt.Printf("    Scope: %s\n", app.Scope)

		// Show token status
		if app.AccessToken != "" {
			if isTokenValid(name) {
				// Parse and format the expiration time for display
				if expiresAt, err := time.Parse(time.RFC3339, app.ExpiresAt); err == nil {
					fmt.Printf("    Token: ✅ Valid (expires: %s)\n", expiresAt.Format("2006-01-02 15:04:05"))
				} else {
					fmt.Printf("    Token: ✅ Valid (expires: %s)\n", app.ExpiresAt)
				}
			} else {
				// Parse and format the expiration time for display
				if expiresAt, err := time.Parse(time.RFC3339, app.ExpiresAt); err == nil {
					fmt.Printf("    Token: ❌ Expired (expired: %s)\n", expiresAt.Format("2006-01-02 15:04:05"))
				} else {
					fmt.Printf("    Token: ❌ Expired (expired: %s)\n", app.ExpiresAt)
				}
			}
		} else {
			fmt.Printf("    Token: ⚠️  No token stored\n")
		}
		fmt.Println()
	}
}

func saveTokensToApp(appName string, tokens *TokenResponse) error {
	app, exists := config.Apps[appName]
	if !exists {
		return fmt.Errorf("app '%s' not found", appName)
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)

	// Update app config with token information
	app.AccessToken = tokens.AccessToken
	app.IdToken = tokens.IdToken
	app.RefreshToken = tokens.RefreshToken
	app.TokenType = tokens.TokenType
	app.ExpiresAt = expiresAt.Format(time.RFC3339)
	app.ExpiresIn = tokens.ExpiresIn

	// Save updated config
	config.Apps[appName] = app
	return saveConfig()
}

func isTokenValid(appName string) bool {
	app, exists := config.Apps[appName]
	if !exists {
		return false
	}

	// Check if we have a token
	if app.AccessToken == "" || app.ExpiresAt == "" {
		return false
	}

	// Parse the expiration time
	expiresAt, err := time.Parse(time.RFC3339, app.ExpiresAt)
	if err != nil {
		// Try alternative format
		expiresAt, err = time.Parse("2006-01-02 15:04:05", app.ExpiresAt)
		if err != nil {
			return false
		}
	}

	// Check if token is not expired
	return time.Now().Before(expiresAt)
}

func getStoredToken(appName string) (*TokenResponse, error) {
	app, exists := config.Apps[appName]
	if !exists {
		return nil, fmt.Errorf("app '%s' not found", appName)
	}

	if app.AccessToken == "" {
		return nil, fmt.Errorf("no token stored for app '%s'", appName)
	}

	if !isTokenValid(appName) {
		return nil, fmt.Errorf("token for app '%s' has expired", appName)
	}

	return &TokenResponse{
		AccessToken:  app.AccessToken,
		IdToken:      app.IdToken,
		RefreshToken: app.RefreshToken,
		TokenType:    app.TokenType,
		ExpiresIn:    app.ExpiresIn,
	}, nil
}
