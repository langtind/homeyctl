package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/langtind/homeyctl/internal/client"
	"github.com/langtind/homeyctl/internal/config"
	"github.com/langtind/homeyctl/internal/oauth"
	"github.com/spf13/cobra"
)

// PAT represents a Personal Access Token from the API
type PAT struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`
	CreatedAt string   `json:"createdAt"`
}

// PATCreateResponse is the response from creating a PAT
type PATCreateResponse struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`
	Token     string   `json:"token"`
	CreatedAt string   `json:"createdAt"`
}

// Available scopes in Homey (from constants.mts)
var availableScopes = []string{
	"homey",                    // Full access (everything)
	"homey.alarm",              // Alarm full access
	"homey.alarm.readonly",     // Alarm read-only
	"homey.app",                // App full access
	"homey.app.readonly",       // App read-only
	"homey.app.control",        // App control
	"homey.dashboard",          // Dashboard full access
	"homey.dashboard.readonly", // Dashboard read-only
	"homey.energy",             // Energy full access
	"homey.energy.readonly",    // Energy read-only
	"homey.system",             // System full access
	"homey.system.readonly",    // System read-only
	"homey.user",               // User full access
	"homey.user.readonly",      // User read-only
	"homey.user.self",          // User self management
	"homey.updates",            // Updates full access
	"homey.updates.readonly",   // Updates read-only
	"homey.geolocation",        // Geolocation full access
	"homey.geolocation.readonly",
	"homey.device",          // Device full access
	"homey.device.readonly", // Device read-only
	"homey.device.control",  // Device control
	"homey.flow",            // Flow full access
	"homey.flow.readonly",   // Flow read-only
	"homey.flow.start",      // Flow trigger/start
	"homey.insights",        // Insights full access
	"homey.insights.readonly",
	"homey.logic", // Logic/variables full access
	"homey.logic.readonly",
	"homey.mood", // Mood full access
	"homey.mood.readonly",
	"homey.mood.set",
	"homey.notifications", // Notifications full access
	"homey.notifications.readonly",
	"homey.reminder", // Reminder full access
	"homey.reminder.readonly",
	"homey.presence", // Presence full access
	"homey.presence.readonly",
	"homey.presence.self",
	"homey.speech", // Speech
	"homey.zone",   // Zone full access
	"homey.zone.readonly",
}

// Scope presets for common use cases
// Note: These must match scopes configured on the OAuth app at developer.athom.com
var scopePresets = map[string][]string{
	"readonly": {
		"homey.device.readonly",
		"homey.flow.readonly",
		"homey.zone.readonly",
		"homey.app.readonly",
		"homey.insights.readonly",
		"homey.notifications.readonly",
		"homey.presence.readonly",
	},
	"control": {
		"homey.device.readonly",
		"homey.device.control",
		"homey.flow.readonly",
		"homey.flow.start",
		"homey.zone.readonly",
		"homey.app.readonly",
		"homey.insights.readonly",
		"homey.notifications.readonly",
		"homey.presence.readonly",
	},
	"full": {
		"homey", // Full access to everything
	},
}

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage Personal Access Tokens (PAT)",
	Long: `Create and manage scoped Personal Access Tokens for AI bots and integrations.

PATs allow you to create tokens with limited permissions, so you can safely
give access to third-party tools without exposing full control of your Homey.

Note: Creating PATs requires an owner account with OAuth or password login.
PAT tokens cannot be used to create new PATs.`,
}

var tokenListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Personal Access Tokens",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := apiClient.ListPATs()
		if err != nil {
			if strings.Contains(err.Error(), "Invalid Session Type") {
				return fmt.Errorf("cannot list PATs: you must be logged in with OAuth or password (PAT tokens cannot manage other PATs)")
			}
			return fmt.Errorf("failed to list tokens: %w", err)
		}

		if isTableFormat() {
			var pats []PAT
			if err := json.Unmarshal(data, &pats); err != nil {
				return fmt.Errorf("failed to parse tokens: %w", err)
			}

			if len(pats) == 0 {
				fmt.Println("No tokens found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tSCOPES\tCREATED")
			fmt.Fprintln(w, "--\t----\t------\t-------")
			for _, p := range pats {
				scopes := formatScopes(p.Scopes)
				created := formatTime(p.CreatedAt)
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.ID, p.Name, scopes, created)
			}
			w.Flush()
			return nil
		}

		outputJSON(data)
		return nil
	},
}

var tokenCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new Personal Access Token",
	Long: `Create a new scoped Personal Access Token.

This command will automatically authenticate via OAuth if needed.
By default, the created token is saved to your config for immediate use.

Use --preset for common scope combinations:
  readonly  - Read-only access to devices, flows, zones, etc.
  control   - Read + control devices and trigger flows
  full      - Full access (same as owner)

Or use --scopes for specific scopes:
  --scopes homey.device.readonly,homey.flow.readonly

Use --no-save to create a token without saving it (for external use).

Examples:
  homeyctl token create "AI Bot" --preset readonly
  homeyctl token create "Home Assistant" --preset control
  homeyctl token create "External" --preset readonly --no-save`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		preset, _ := cmd.Flags().GetString("preset")
		scopesStr, _ := cmd.Flags().GetString("scopes")
		noSave, _ := cmd.Flags().GetBool("no-save")

		var scopes []string

		if preset != "" && scopesStr != "" {
			return fmt.Errorf("cannot use both --preset and --scopes")
		}

		if preset != "" {
			presetScopes, ok := scopePresets[preset]
			if !ok {
				return fmt.Errorf("unknown preset: %s (available: readonly, control, full)", preset)
			}
			scopes = presetScopes
		} else if scopesStr != "" {
			scopes = strings.Split(scopesStr, ",")
			for i := range scopes {
				scopes[i] = strings.TrimSpace(scopes[i])
			}
		} else {
			return fmt.Errorf("must specify --preset or --scopes")
		}

		// Try to use existing config, or do OAuth login
		existingCfg, _ := config.Load()

		needsOAuth := existingCfg == nil || existingCfg.Token == ""

		if !needsOAuth {
			// Check if current token can manage PATs by listing them
			tempClient := client.New(existingCfg)
			_, err := tempClient.ListPATs()
			if err != nil {
				errStr := err.Error()
				if strings.Contains(errStr, "Invalid Session Type") {
					// Current token is a PAT, need OAuth
					needsOAuth = true
				}
				// Other errors might just be network issues, we'll catch them later
			}
		}

		// Need OAuth login
		if needsOAuth {
			fmt.Println("OAuth authentication required to create tokens...")
			homey, err := oauth.Login()
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			// Determine which URL to use
			homeyURL := homey.LocalURLSecure
			if homeyURL == "" {
				homeyURL = homey.LocalURL
			}
			if homeyURL == "" {
				homeyURL = homey.RemoteURL
			}

			// Create a temporary client with the OAuth session
			parsedURL, err := url.Parse(homeyURL)
			if err != nil {
				return fmt.Errorf("failed to parse Homey URL: %w", err)
			}

			host := parsedURL.Hostname()
			port := 443
			if parsedURL.Port() != "" {
				fmt.Sscanf(parsedURL.Port(), "%d", &port)
			} else if parsedURL.Scheme == "http" {
				port = 80
			}

			existingCfg = &config.Config{
				Host:  host,
				Port:  port,
				Token: homey.Token,
				TLS:   parsedURL.Scheme == "https",
			}
		}

		// Create the PAT
		tempClient := client.New(existingCfg)
		data, err := tempClient.CreatePAT(name, scopes)
		if err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, "Invalid Session Type") {
				return fmt.Errorf("cannot create PAT: authentication failed. Please try again")
			}
			if strings.Contains(errStr, "Must Be Owner") {
				return fmt.Errorf("cannot create PAT: only the owner account can create tokens")
			}
			if strings.Contains(errStr, "Missing Scopes") {
				return fmt.Errorf("cannot create PAT: the requested scopes are not available.\nTry a different preset or check available scopes with: homeyctl token scopes")
			}
			return fmt.Errorf("failed to create token: %w", err)
		}

		var resp PATCreateResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		fmt.Println()
		fmt.Println("Token created successfully!")
		fmt.Printf("Name:   %s\n", resp.Name)
		fmt.Printf("Scopes: %s\n", strings.Join(resp.Scopes, ", "))

		if noSave {
			// Just print the token
			fmt.Println()
			fmt.Printf("Token: %s\n", resp.Token)
			fmt.Println()
			fmt.Println("IMPORTANT: Save this token now - it cannot be retrieved later.")
		} else {
			// Save the PAT to config
			newCfg := &config.Config{
				Host:   existingCfg.Host,
				Port:   existingCfg.Port,
				Token:  resp.Token,
				Format: "json",
				TLS:    existingCfg.TLS,
			}

			if err := config.Save(newCfg); err != nil {
				// Still show the token if save fails
				fmt.Println()
				fmt.Printf("Token: %s\n", resp.Token)
				fmt.Println()
				return fmt.Errorf("token created but failed to save config: %w\nSave it manually with: homeyctl config set-token <token>", err)
			}

			fmt.Println()
			fmt.Println("Token saved to config. You're ready to use homeyctl!")
			fmt.Printf("Try: homeyctl devices list\n")
		}

		return nil
	},
}

var tokenDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a Personal Access Token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		if err := apiClient.DeletePAT(id); err != nil {
			return fmt.Errorf("failed to delete token: %w", err)
		}

		fmt.Printf("Token deleted: %s\n", id)
		return nil
	},
}

var tokenScopesCmd = &cobra.Command{
	Use:   "scopes",
	Short: "List available scopes",
	Long: `List all available scopes that can be used when creating tokens.

Scopes follow a hierarchy:
  - homey.device includes homey.device.readonly and homey.device.control
  - homey.flow includes homey.flow.readonly and homey.flow.start
  - homey (full access) includes all scopes`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available scopes:")
		fmt.Println()
		fmt.Println("PRESETS:")
		fmt.Println("  readonly  - Read-only access to all resources")
		fmt.Println("  control   - Read + control devices and trigger flows")
		fmt.Println("  full      - Full access (same as owner)")
		fmt.Println()
		fmt.Println("INDIVIDUAL SCOPES:")
		for _, scope := range availableScopes {
			fmt.Printf("  %s\n", scope)
		}
	},
}

var tokenLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login with your Athom account to manage tokens",
	Long: `Login with your Athom account using OAuth.

This opens a browser where you can authenticate with your Athom account.
After successful login, your Homey will be configured automatically.

This is required to create, list, or delete Personal Access Tokens (PATs).
PAT tokens themselves cannot manage other PATs - you need OAuth login.

Example:
  homeyctl token login
  homeyctl token create "AI Bot" --preset readonly`,
	RunE: func(cmd *cobra.Command, args []string) error {
		homey, err := oauth.Login()
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		// Determine which URL to use (prefer local secure, then local, then remote)
		homeyURL := homey.LocalURLSecure
		if homeyURL == "" {
			homeyURL = homey.LocalURL
		}
		if homeyURL == "" {
			homeyURL = homey.RemoteURL
		}

		// Parse URL to extract host and port
		parsedURL, err := url.Parse(homeyURL)
		if err != nil {
			return fmt.Errorf("failed to parse Homey URL: %w", err)
		}

		host := parsedURL.Hostname()
		port := 443
		if parsedURL.Port() != "" {
			fmt.Sscanf(parsedURL.Port(), "%d", &port)
		} else if parsedURL.Scheme == "http" {
			port = 80
		}

		// Save the configuration
		newCfg := &config.Config{
			Host:   host,
			Port:   port,
			Token:  homey.Token,
			Format: "json",
			TLS:    parsedURL.Scheme == "https",
		}

		if err := config.Save(newCfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println()
		fmt.Printf("Successfully logged in!\n")
		fmt.Printf("Homey: %s\n", homey.Name)
		fmt.Printf("URL:   %s\n", homeyURL)
		fmt.Println()
		fmt.Println("You can now use 'homeyctl token create' to create scoped tokens.")

		return nil
	},
}

func formatScopes(scopes []string) string {
	if len(scopes) == 0 {
		return "-"
	}
	if len(scopes) == 1 {
		return scopes[0]
	}
	if len(scopes) <= 3 {
		return strings.Join(scopes, ", ")
	}
	return fmt.Sprintf("%s, +%d more", scopes[0], len(scopes)-1)
}

func formatTime(timeStr string) string {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return timeStr
	}
	return t.Format("2006-01-02")
}

func init() {
	rootCmd.AddCommand(tokenCmd)
	tokenCmd.AddCommand(tokenLoginCmd)
	tokenCmd.AddCommand(tokenListCmd)
	tokenCmd.AddCommand(tokenCreateCmd)
	tokenCmd.AddCommand(tokenDeleteCmd)
	tokenCmd.AddCommand(tokenScopesCmd)

	tokenCreateCmd.Flags().String("preset", "", "Scope preset: readonly, control, or full")
	tokenCreateCmd.Flags().String("scopes", "", "Comma-separated list of scopes")
	tokenCreateCmd.Flags().Bool("no-save", false, "Don't save token to config (for external use)")
}
