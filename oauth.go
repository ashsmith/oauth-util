package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/browser"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	IdToken      string `json:"id_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
}

type OAuthFlow struct {
	server    *http.Server
	authCode  chan string
	authError chan string
	port      string
}

func NewOAuthFlow() *OAuthFlow {
	return &OAuthFlow{
		authCode:  make(chan string, 1),
		authError: make(chan string, 1),
	}
}

func (o *OAuthFlow) StartFlow(appConfig AppConfig, port string) (*TokenResponse, error) {
	o.port = port

	// Start local server
	if err := o.startCallbackServer(); err != nil {
		return nil, fmt.Errorf("failed to start callback server: %v", err)
	}
	defer o.cleanup()

	// Build authorization URL
	authURL := o.buildAuthURL(appConfig)

	// Open browser
	if err := browser.OpenURL(authURL); err != nil {
		return nil, fmt.Errorf("failed to open browser: %v", err)
	}

	// Wait for authorization code
	select {
	case code := <-o.authCode:
		// Exchange code for tokens
		tokens, err := o.exchangeCodeForTokens(code, appConfig)
		if err != nil {
			return nil, fmt.Errorf("token exchange failed: %v", err)
		}
		return tokens, nil
	case err := <-o.authError:
		return nil, fmt.Errorf("OAuth error: %s", err)
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("timeout waiting for authorization")
	}
}

func (o *OAuthFlow) startCallbackServer() error {
	r := mux.NewRouter()
	r.HandleFunc("/", o.handleCallback)

	o.server = &http.Server{
		Addr:    ":" + o.port,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		if err := o.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			o.authError <- fmt.Sprintf("server error: %v", err)
		}
	}()

	// Give server a moment to start
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (o *OAuthFlow) handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	error := r.URL.Query().Get("error")

	if error != "" {
		http.Error(w, fmt.Sprintf("Authentication Error: %s", error), http.StatusBadRequest)
		o.authError <- error
		return
	}

	if code == "" {
		http.Error(w, "No authorization code received", http.StatusBadRequest)
		o.authError <- "no authorization code"
		return
	}

	// Send success response
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
		<html>
			<body>
				<h1>Authentication Successful!</h1>
				<p>You can close this window and return to the terminal.</p>
			</body>
		</html>
	`))

	// Send code to main flow
	o.authCode <- code
}

func (o *OAuthFlow) buildAuthURL(appConfig AppConfig) string {
	redirectURI := fmt.Sprintf("http://localhost:%s/", o.port)

	// Parse domain to get base URL
	domainURL, _ := url.Parse(appConfig.Domain)
	authEndpoint := fmt.Sprintf("%s://%s/oauth2/authorize", domainURL.Scheme, domainURL.Host)

	params := url.Values{}
	params.Set("client_id", appConfig.ClientID)
	params.Set("response_type", "code")
	params.Set("scope", appConfig.Scope)
	params.Set("redirect_uri", redirectURI)

	return fmt.Sprintf("%s?%s", authEndpoint, params.Encode())
}

func (o *OAuthFlow) exchangeCodeForTokens(code string, appConfig AppConfig) (*TokenResponse, error) {
	redirectURI := fmt.Sprintf("http://localhost:%s/", o.port)

	// Parse domain to get base URL
	domainURL, _ := url.Parse(appConfig.Domain)
	tokenEndpoint := fmt.Sprintf("%s://%s/oauth2/token", domainURL.Scheme, domainURL.Host)

	// Prepare form data
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", appConfig.ClientID)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	// Create HTTP request
	req, err := http.NewRequest("POST", tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			if desc, ok := errorResp["error_description"].(string); ok {
				return nil, fmt.Errorf("token exchange failed: %s", desc)
			}
			if errMsg, ok := errorResp["error"].(string); ok {
				return nil, fmt.Errorf("token exchange failed: %s", errMsg)
			}
		}
		return nil, fmt.Errorf("token exchange failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var tokens TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %v", err)
	}

	return &tokens, nil
}

func (o *OAuthFlow) cleanup() {
	if o.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		o.server.Shutdown(ctx)
	}
}
