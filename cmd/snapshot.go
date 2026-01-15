package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var snapshotIncludeFlows bool

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Get a snapshot of Homey state",
	Long: `Get status, zones, and devices in one call.

Useful for AI assistants and scripts that need a complete overview.

Examples:
  homeyctl snapshot
  homeyctl snapshot --include-flows
  homeyctl snapshot --format json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get system status
		systemData, err := apiClient.GetSystem()
		if err != nil {
			return fmt.Errorf("failed to get system status: %w", err)
		}

		// Get zones
		zonesData, err := apiClient.GetZones()
		if err != nil {
			return fmt.Errorf("failed to get zones: %w", err)
		}

		// Get devices
		devicesData, err := apiClient.GetDevices()
		if err != nil {
			return fmt.Errorf("failed to get devices: %w", err)
		}

		// Parse for counting
		var zones map[string]interface{}
		var devices map[string]interface{}
		json.Unmarshal(zonesData, &zones)
		json.Unmarshal(devicesData, &devices)

		snapshot := map[string]json.RawMessage{
			"system":  systemData,
			"zones":   zonesData,
			"devices": devicesData,
		}

		// Optionally include flows
		if snapshotIncludeFlows {
			flowsData, err := apiClient.GetFlows()
			if err != nil {
				return fmt.Errorf("failed to get flows: %w", err)
			}
			snapshot["flows"] = flowsData

			advFlowsData, err := apiClient.GetAdvancedFlows()
			if err != nil {
				return fmt.Errorf("failed to get advanced flows: %w", err)
			}
			snapshot["advancedFlows"] = advFlowsData
		}

		if isTableFormat() {
			var system map[string]interface{}
			json.Unmarshal(systemData, &system)

			fmt.Println("Homey Snapshot")
			fmt.Println("==============")
			fmt.Printf("Model:     %v\n", system["homeyModelName"])
			fmt.Printf("Version:   %v\n", system["homeyVersion"])
			fmt.Printf("Platform:  %v (v%v)\n", system["homeyPlatform"], system["homeyPlatformVersion"])
			fmt.Printf("Hostname:  %v\n", system["hostname"])
			fmt.Printf("Zones:     %d\n", len(zones))
			fmt.Printf("Devices:   %d\n", len(devices))

			if snapshotIncludeFlows {
				var flows, advFlows map[string]interface{}
				json.Unmarshal(snapshot["flows"], &flows)
				json.Unmarshal(snapshot["advancedFlows"], &advFlows)
				fmt.Printf("Flows:     %d\n", len(flows))
				fmt.Printf("Advanced:  %d\n", len(advFlows))
			}
			return nil
		}

		out, _ := json.MarshalIndent(snapshot, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(snapshotCmd)
	snapshotCmd.Flags().BoolVar(&snapshotIncludeFlows, "include-flows", false, "Include flows in snapshot")
}
