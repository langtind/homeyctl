package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	ClientID     = "69637f960b43d2064df4fa70"
	ClientSecret = "922a71f421653b83b99d7a98be74d8ab"
	RedirectURI  = "http://localhost:8484/callback"
	AuthURL      = "https://api.athom.com/oauth2/authorise"
	TokenURL     = "https://api.athom.com/oauth2/token"
)

// All available scopes - request full access so we can create any scoped PATs
var AllScopes = []string{
	"homey",
	"homey.device",
	"homey.device.readonly",
	"homey.device.control",
	"homey.flow",
	"homey.flow.readonly",
	"homey.flow.start",
	"homey.zone",
	"homey.zone.readonly",
	"homey.app",
	"homey.app.readonly",
	"homey.app.control",
	"homey.logic",
	"homey.logic.readonly",
	"homey.insights",
	"homey.insights.readonly",
	"homey.energy",
	"homey.energy.readonly",
	"homey.system",
	"homey.system.readonly",
	"homey.user",
	"homey.user.readonly",
	"homey.user.self",
	"homey.notifications",
	"homey.notifications.readonly",
	"homey.presence",
	"homey.presence.readonly",
	"homey.presence.self",
	"homey.alarm",
	"homey.alarm.readonly",
	"homey.dashboard",
	"homey.dashboard.readonly",
	"homey.updates",
	"homey.updates.readonly",
	"homey.mood",
	"homey.mood.readonly",
	"homey.mood.set",
	"homey.reminder",
	"homey.reminder.readonly",
	"homey.geolocation",
	"homey.geolocation.readonly",
	"homey.speech",
}

// TokenResponse represents the OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// User represents an Athom user
type User struct {
	ID        string  `json:"_id"`
	Email     string  `json:"email"`
	FirstName string  `json:"firstname"`
	LastName  string  `json:"lastname"`
	Homeys    []Homey `json:"homeys"`
}

// Homey represents a Homey device
type Homey struct {
	ID             string `json:"_id"`
	Name           string `json:"name"`
	LocalAddress   string `json:"localAddress"`
	LocalURL       string `json:"localUrl"`
	LocalURLSecure string `json:"localUrlSecure"`
	RemoteURL      string `json:"remoteUrl"`
	Token          string // Session token for this specific Homey
}

// Login performs the OAuth login flow and returns a delegation token for the selected Homey
func Login() (*Homey, error) {
	// Create channel to receive auth code
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Start local server
	server := &http.Server{Addr: ":8484"}
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errChan <- fmt.Errorf("no authorization code received")
			fmt.Fprintf(w, "<html><body><h1>Error</h1><p>No authorization code received.</p></body></html>")
			return
		}
		codeChan <- code
		fmt.Fprintf(w, "<html><body><h1>Success!</h1><p>You can close this window and return to the terminal.</p></body></html>")
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Build authorization URL (Athom uses 'scopes' not standard 'scope')
	scopes := strings.Join(AllScopes, " ")
	authURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scopes=%s",
		AuthURL,
		ClientID,
		url.QueryEscape(RedirectURI),
		url.QueryEscape(scopes),
	)

	// Open browser
	fmt.Println("Opening browser for authentication...")
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Could not open browser. Please visit:\n%s\n", authURL)
	}

	// Wait for auth code or error
	var code string
	select {
	case code = <-codeChan:
	case err := <-errChan:
		server.Shutdown(context.Background())
		return nil, err
	case <-time.After(5 * time.Minute):
		server.Shutdown(context.Background())
		return nil, fmt.Errorf("authentication timed out")
	}

	// Shutdown server
	server.Shutdown(context.Background())

	// Exchange code for token
	fmt.Println("Exchanging authorization code for token...")
	tokenResp, err := exchangeCodeForToken(code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Get user info (includes Homeys)
	fmt.Println("Getting user info...")
	user, err := getUser(tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	fmt.Printf("Logged in as: %s %s (%s)\n", user.FirstName, user.LastName, user.Email)
	fmt.Printf("Found %d Homey(s)\n", len(user.Homeys))

	if len(user.Homeys) == 0 {
		return nil, fmt.Errorf("no Homeys found on your account")
	}

	// Debug: print Homey info
	for _, h := range user.Homeys {
		fmt.Printf("  - %s (ID: %s)\n", h.Name, h.ID)
		fmt.Printf("    LocalURL: %s\n", h.LocalURL)
		fmt.Printf("    LocalURLSecure: %s\n", h.LocalURLSecure)
		fmt.Printf("    RemoteURL: %s\n", h.RemoteURL)
	}

	// Select Homey (for now just use the first one, could add selection UI later)
	selectedHomey := user.Homeys[0]
	if len(user.Homeys) > 1 {
		fmt.Println("\nFound multiple Homeys:")
		for i, h := range user.Homeys {
			fmt.Printf("  %d. %s (%s)\n", i+1, h.Name, h.LocalAddress)
		}
		fmt.Printf("\nUsing: %s\n", selectedHomey.Name)
	}

	// Use RemoteURL for OAuth authentication (required for delegation token flow)
	// LocalURL can be used after getting a session token
	homeyURL := selectedHomey.RemoteURL
	localURL := selectedHomey.LocalURL
	if localURL == "" {
		localURL = selectedHomey.LocalURLSecure
	}
	if homeyURL == "" {
		return nil, fmt.Errorf("no URL available for Homey %s", selectedHomey.Name)
	}

	// Try to get delegation token
	fmt.Println("Getting delegation token...")
	delegationToken, err := getDelegationToken(tokenResp.AccessToken, selectedHomey.ID)
	if err != nil {
		fmt.Printf("Note: Delegation token not available (%v)\n", err)
		fmt.Println("Trying to login with access token via remote URL...")

		// Try login with access token to remote URL
		sessionToken, loginErr := loginToHomey(homeyURL, tokenResp.AccessToken)
		if loginErr != nil {
			// Try with local URL as last resort
			if localURL != "" {
				fmt.Println("Trying local URL...")
				sessionToken, loginErr = loginToHomey(localURL, tokenResp.AccessToken)
			}
			if loginErr != nil {
				return nil, fmt.Errorf("could not authenticate with Homey.\n\nThe OAuth login flow requires the delegation API which may not be available for your account.\n\nAlternative: Create an API key manually via the web UI:\n  1. Go to %s\n  2. Navigate to Settings → API Keys\n  3. Create a new API key with the scopes you need\n  4. Run: homeyctl config set-token <your-token>", localURL)
			}
		}
		selectedHomey.Token = sessionToken
		selectedHomey.LocalURL = localURL // Use local URL for subsequent API calls
		return &selectedHomey, nil
	}

	// Login to Homey with delegation token
	fmt.Printf("Logging in to %s...\n", selectedHomey.Name)
	sessionToken, err := loginToHomey(homeyURL, delegationToken)
	if err != nil {
		return nil, fmt.Errorf("failed to login to Homey: %w", err)
	}

	selectedHomey.Token = sessionToken
	selectedHomey.LocalURL = localURL // Use local URL for subsequent API calls
	return &selectedHomey, nil
}

func exchangeCodeForToken(code string) (*TokenResponse, error) {
	data := url.Values{
		"client_id":     {ClientID},
		"client_secret": {ClientSecret},
		"grant_type":    {"authorization_code"},
		"code":          {code},
	}

	resp, err := http.PostForm(TokenURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

func getUser(accessToken string) (*User, error) {
	req, err := http.NewRequest("GET", "https://api.athom.com/user/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get user: %s", string(body))
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func getDelegationToken(accessToken, homeyID string) (string, error) {
	// Athom binder tokenet til én Homey når homeyId er med, og vil avvise ubundne tokens.
	req, err := http.NewRequest("POST", "https://api.athom.com/delegation/token?audience=homey&homeyId="+url.QueryEscape(homeyID), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to get delegation token: %s", string(body))
	}

	// Response is just a string (the token)
	var token string
	if err := json.Unmarshal(body, &token); err != nil {
		// Try as raw string
		token = strings.Trim(string(body), "\"")
	}

	return token, nil
}

func loginToHomey(homeyURL, delegationToken string) (string, error) {
	loginBody := map[string]string{"token": delegationToken}
	jsonBody, err := json.Marshal(loginBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", homeyURL+"/api/manager/users/login", bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to login to Homey: %s", string(body))
	}

	// Response is a JSON string (the session token)
	var sessionToken string
	if err := json.Unmarshal(body, &sessionToken); err != nil {
		return "", fmt.Errorf("failed to parse session token: %w", err)
	}

	return sessionToken, nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return cmd.Start()
}
