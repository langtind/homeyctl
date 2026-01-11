package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/langtind/homeyctl/internal/client"
	"github.com/langtind/homeyctl/internal/config"
	"github.com/langtind/homeyctl/internal/oauth"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to your Homey",
	Long: `Log in to your Homey with your Athom account.

This is the easiest way to get started with homeyctl:
  1. Opens your browser to log in with your Athom account
  2. Creates an API token with device control access
  3. Saves it to your config

After login, you can immediately use homeyctl:
  homeyctl devices list
  homeyctl flows list

For read-only access (safer for AI bots), use:
  homeyctl token create "AI Bot" --preset readonly`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Logging in to your Homey...")
		fmt.Println()

		// Do OAuth login
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

		// Parse URL
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

		// Create temporary config for API client
		tempCfg := &config.Config{
			Host:  host,
			Port:  port,
			Token: homey.Token,
			TLS:   parsedURL.Scheme == "https",
		}

		// Create a "control" preset token for the user
		tempClient := client.New(tempCfg)
		scopes := scopePresets["control"]

		data, err := tempClient.CreatePAT("homeyctl", scopes)
		if err != nil {
			// If token creation fails, save the OAuth session token instead
			// (less ideal but still works)
			fmt.Println("Note: Could not create scoped token, using session token.")
			if saveErr := config.Save(tempCfg); saveErr != nil {
				return fmt.Errorf("failed to save config: %w", saveErr)
			}
			fmt.Println()
			fmt.Printf("Logged in to: %s\n", homey.Name)
			fmt.Println()
			fmt.Println("You're ready to use homeyctl!")
			fmt.Println("Try: homeyctl devices list")
			return nil
		}

		var resp PATCreateResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		// Save the PAT to config
		newCfg := &config.Config{
			Host:   host,
			Port:   port,
			Token:  resp.Token,
			Format: "json",
			TLS:    parsedURL.Scheme == "https",
		}

		if err := config.Save(newCfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println()
		fmt.Printf("Logged in to: %s\n", homey.Name)
		fmt.Println()
		fmt.Println("You're ready to use homeyctl!")
		fmt.Println("Try: homeyctl devices list")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
