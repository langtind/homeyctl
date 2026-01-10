package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

type SystemInfo struct {
	HomeyVersion         string      `json:"homeyVersion"`
	HomeyModelID         string      `json:"homeyModelId"`
	HomeyModelName       string      `json:"homeyModelName"`
	HomeyPlatformVersion interface{} `json:"homeyPlatformVersion"`
	Uptime               float64     `json:"uptime"`
	Date                 string      `json:"date"`
	WifiSSID             string      `json:"wifiSsid"`
	CloudConnected       bool        `json:"cloudConnected"`
	Address              string      `json:"address"`
	BootDate             string      `json:"bootDate"`
	Country              string      `json:"country"`
}

var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "System information and control",
	Long:  `View system information and perform system operations.`,
}

var systemInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show system information",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := apiClient.GetSystem()
		if err != nil {
			return err
		}

		if isTableFormat() {
			var info SystemInfo
			if err := json.Unmarshal(data, &info); err != nil {
				return fmt.Errorf("failed to parse system info: %w", err)
			}

			fmt.Printf("Homey Version:    %s\n", info.HomeyVersion)
			fmt.Printf("Model:            %s (%s)\n", info.HomeyModelName, info.HomeyModelID)
			fmt.Printf("Platform Version: %v\n", info.HomeyPlatformVersion)
			fmt.Printf("Address:          %s\n", info.Address)
			fmt.Printf("Country:          %s\n", info.Country)
			fmt.Printf("Boot Date:        %s\n", info.BootDate)
			fmt.Printf("Uptime:           %.0f seconds\n", info.Uptime)
			fmt.Printf("Cloud Connected:  %v\n", info.CloudConnected)
			return nil
		}

		outputJSON(data)
		return nil
	},
}

var systemRebootCmd = &cobra.Command{
	Use:   "reboot",
	Short: "Reboot Homey",
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			return fmt.Errorf("use --force to confirm reboot")
		}

		if err := apiClient.Reboot(); err != nil {
			return err
		}

		fmt.Println("Reboot initiated")
		return nil
	},
}

var systemUsersCmd = &cobra.Command{
	Use:   "users",
	Short: "List users",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := apiClient.GetUsers()
		if err != nil {
			return err
		}

		outputJSON(data)
		return nil
	},
}

var systemInsightsCmd = &cobra.Command{
	Use:   "insights",
	Short: "List insight logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := apiClient.GetInsights()
		if err != nil {
			return err
		}

		outputJSON(data)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(systemCmd)
	systemCmd.AddCommand(systemInfoCmd)
	systemCmd.AddCommand(systemRebootCmd)
	systemCmd.AddCommand(systemUsersCmd)
	systemCmd.AddCommand(systemInsightsCmd)

	systemRebootCmd.Flags().Bool("force", false, "Confirm reboot")
}
