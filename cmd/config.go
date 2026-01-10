package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/langtind/homey-cli/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `View and modify homey-cli configuration.`,
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
			// Mask token for security
			maskedToken := ""
			if loadedCfg.Token != "" {
				token := loadedCfg.Token
				if len(token) > 20 {
					maskedToken = token[:8] + "..." + token[len(token)-8:]
				} else {
					maskedToken = token
				}
			}

			output := map[string]interface{}{
				"host":   loadedCfg.Host,
				"port":   loadedCfg.Port,
				"format": loadedCfg.Format,
				"token":  maskedToken,
			}
			out, _ := json.MarshalIndent(output, "", "  ")
			fmt.Println(string(out))
			return nil
		}

		fmt.Printf("Host:   %s\n", loadedCfg.Host)
		fmt.Printf("Port:   %d\n", loadedCfg.Port)
		fmt.Printf("Format: %s\n", loadedCfg.Format)
		if loadedCfg.Token != "" {
			token := loadedCfg.Token
			if len(token) > 20 {
				token = token[:8] + "..." + token[len(token)-8:]
			}
			fmt.Printf("Token:  %s\n", token)
		} else {
			fmt.Println("Token:  (not set)")
		}
		return nil
	},
}

var configSetTokenCmd = &cobra.Command{
	Use:   "set-token <token>",
	Short: "Set API token",
	Args:  cobra.ExactArgs(1),
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
	Args:  cobra.ExactArgs(1),
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

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetTokenCmd)
	configCmd.AddCommand(configSetHostCmd)
}
