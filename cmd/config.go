package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/langtind/homeyctl/internal/config"
	"github.com/langtind/homeyctl/internal/discovery"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `View and modify homey-cli configuration.`,
}

func maskToken(token string) string {
	if token == "" {
		return "(not set)"
	}
	if len(token) > 20 {
		return token[:8] + "..." + token[len(token)-8:]
	}
	return token
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		loadedCfg, err := config.Load()
		if err != nil {
			return err
		}

		// Check format flag directly since cfg may not be set for config commands
		format := formatFlag
		if format == "" {
			format = loadedCfg.Format
		}

		if format != "table" {
			output := map[string]interface{}{
				"mode":          loadedCfg.Mode,
				"effectiveMode": loadedCfg.EffectiveMode(),
				"local": map[string]interface{}{
					"address": loadedCfg.Local.Address,
					"token":   maskToken(loadedCfg.Local.Token),
				},
				"cloud": map[string]interface{}{
					"token": maskToken(loadedCfg.Cloud.Token),
				},
				"legacy": map[string]interface{}{
					"host":  loadedCfg.Host,
					"port":  loadedCfg.Port,
					"token": maskToken(loadedCfg.Token),
				},
				"format": loadedCfg.Format,
			}
			out, _ := json.MarshalIndent(output, "", "  ")
			fmt.Println(string(out))
			return nil
		}

		fmt.Println("Connection Mode")
		fmt.Println("===============")
		mode := loadedCfg.Mode
		if mode == "" {
			mode = "auto"
		}
		fmt.Printf("Mode:           %s\n", mode)
		fmt.Printf("Effective:      %s\n", loadedCfg.EffectiveMode())
		fmt.Println()

		fmt.Println("Local (LAN/VPN)")
		fmt.Println("---------------")
		if loadedCfg.Local.Address != "" {
			fmt.Printf("Address:        %s\n", loadedCfg.Local.Address)
		} else {
			fmt.Printf("Address:        (not set)\n")
		}
		fmt.Printf("Token:          %s\n", maskToken(loadedCfg.Local.Token))
		fmt.Println()

		fmt.Println("Cloud")
		fmt.Println("-----")
		fmt.Printf("Token:          %s\n", maskToken(loadedCfg.Cloud.Token))
		fmt.Println()

		// Show legacy if set
		if loadedCfg.Host != "localhost" || loadedCfg.Token != "" {
			fmt.Println("Legacy (deprecated)")
			fmt.Println("-------------------")
			fmt.Printf("Host:           %s\n", loadedCfg.Host)
			fmt.Printf("Port:           %d\n", loadedCfg.Port)
			fmt.Printf("Token:          %s\n", maskToken(loadedCfg.Token))
		}

		return nil
	},
}

var configSetTokenCmd = &cobra.Command{
	Use:   "set-token <token>",
	Short: "Set API token",
	Long: `Set the API token for authenticating with your Homey.

To create a new API key:
  1. Go to https://my.homey.app/
  2. Select your Homey
  3. Click Settings (gear icon, bottom left)
  4. Click API Keys
  5. Click "+ New API Key"
  6. Copy the generated token and use it here`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			cfg = &config.Config{
				Host:   "localhost",
				Port:   4859,
				Format: "table",
			}
		}

		cfg.Token = args[0]

		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Println("Token saved successfully")
		return nil
	},
}

var configSetHostCmd = &cobra.Command{
	Use:   "set-host <host>",
	Short: "Set Homey host",
	Long: `Set the IP address of your Homey.

To find your Homey's IP address:
  - Open the Homey app on your phone
  - Go to Settings → General
  - Scroll down to find the local IP address (e.g., 192.168.1.100)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			cfg = &config.Config{
				Port:   4859,
				Format: "table",
			}
		}

		cfg.Host = args[0]

		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Printf("Host set to: %s\n", cfg.Host)
		return nil
	},
}

var configSetModeCmd = &cobra.Command{
	Use:   "set-mode <auto|local|cloud>",
	Short: "Set connection mode",
	Long: `Set the connection mode for Homey.

Modes:
  auto  - Prefer local if configured, fallback to cloud (default)
  local - Always use local connection (LAN/VPN)
  cloud - Always use cloud connection

Examples:
  homeyctl config set-mode auto
  homeyctl config set-mode local
  homeyctl config set-mode cloud`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mode := args[0]
		if mode != "auto" && mode != "local" && mode != "cloud" {
			return fmt.Errorf("invalid mode: %s (must be auto, local, or cloud)", mode)
		}

		cfg, err := config.Load()
		if err != nil {
			cfg = &config.Config{}
		}

		cfg.Mode = mode

		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Printf("Mode set to: %s\n", mode)
		return nil
	},
}

var configSetLocalCmd = &cobra.Command{
	Use:   "set-local <address> <token>",
	Short: "Set local connection settings",
	Long: `Set the local Homey address and token.

The address should be the full URL including protocol (http/https).
The token is the local API key from your Homey.

Examples:
  homeyctl config set-local http://192.168.1.50 "your-local-api-key"
  homeyctl config set-local https://homey.local:443 "your-local-api-key"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		address := args[0]
		token := args[1]

		cfg, err := config.Load()
		if err != nil {
			cfg = &config.Config{}
		}

		cfg.Local.Address = address
		cfg.Local.Token = token

		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Printf("Local address set to: %s\n", address)
		fmt.Println("Local token saved")
		return nil
	},
}

var configSetCloudCmd = &cobra.Command{
	Use:   "set-cloud <token>",
	Short: "Set cloud token",
	Long: `Set the cloud token (PAT) for remote Homey access.

Create a cloud token at:
  https://my.homey.app → Select Homey → Settings → API Keys

Examples:
  homeyctl config set-cloud "your-cloud-token"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		token := args[0]

		cfg, err := config.Load()
		if err != nil {
			cfg = &config.Config{}
		}

		cfg.Cloud.Token = token

		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Println("Cloud token saved")
		return nil
	},
}

var discoverTimeout int

var configDiscoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover Homey on local network",
	Long: `Discover Homey devices on the local network via mDNS.

This command searches for Homey devices broadcasting on the network
and verifies them by checking the ping endpoint.

Returns JSON array with discovered devices (easy for AI/scripts to parse).

Examples:
  homeyctl config discover
  homeyctl config discover --timeout 10`,
	RunE: func(cmd *cobra.Command, args []string) error {
		timeout := time.Duration(discoverTimeout) * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout+2*time.Second)
		defer cancel()

		// Load config to check format (config commands skip PersistentPreRunE)
		loadedCfg, err := config.Load()
		if err != nil {
			loadedCfg = &config.Config{Format: "json"}
		}

		format := formatFlag
		if format == "" {
			format = loadedCfg.Format
		}
		useTable := format == "table"

		if useTable {
			fmt.Printf("Searching for Homey devices (timeout: %ds)...\n", discoverTimeout)
		}

		candidates, err := discovery.DiscoverAndVerify(ctx, timeout)
		if err != nil {
			return fmt.Errorf("discovery failed: %w", err)
		}

		// JSON output (default)
		if !useTable {
			result := make([]map[string]interface{}, len(candidates))
			for i, c := range candidates {
				result[i] = map[string]interface{}{
					"address": c.Address,
					"homeyId": c.HomeyID,
					"host":    c.Host,
					"port":    c.Port,
				}
			}
			out, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(out))
			return nil
		}

		// Table output for humans
		if len(candidates) == 0 {
			fmt.Println("\nNo Homey devices found.")
			fmt.Println("\nTips:")
			fmt.Println("  - Make sure you're on the same network as your Homey")
			fmt.Println("  - Try increasing timeout: --timeout 10")
			fmt.Println("  - Set address manually: homeyctl config set-local <address> <token>")
			return nil
		}

		fmt.Printf("\nFound %d Homey device(s):\n\n", len(candidates))

		for i, c := range candidates {
			fmt.Printf("  [%d] %s\n", i+1, c.Address)
			if c.HomeyID != "" {
				fmt.Printf("      Homey ID: %s\n", c.HomeyID)
			}
		}

		fmt.Println("\nTo use a discovered Homey:")
		fmt.Println("  homeyctl config set-local <address> <your-api-key>")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetTokenCmd)
	configCmd.AddCommand(configSetHostCmd)
	configCmd.AddCommand(configSetModeCmd)
	configCmd.AddCommand(configSetLocalCmd)
	configCmd.AddCommand(configSetCloudCmd)
	configCmd.AddCommand(configDiscoverCmd)
	configDiscoverCmd.Flags().IntVar(&discoverTimeout, "timeout", 5, "Discovery timeout in seconds")
}
