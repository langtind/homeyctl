package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

type InsightLog struct {
	ID       string `json:"id"`
	OwnerURI string `json:"ownerUri"`
	OwnerID  string `json:"ownerId"`
	Title    string `json:"title"`
	Type     string `json:"type"`
	Units    string `json:"units"`
}

var insightsCmd = &cobra.Command{
	Use:   "insights",
	Short: "Manage insights",
	Long:  `View Homey insights logs and historical data.`,
}

// insightEntry is a single data point in an insight log.
type insightEntry struct {
	T time.Time   `json:"t"`
	V interface{} `json:"v"`
}

// parseInsightEntries unmarshals the insights entries response. The Homey API
// returns an object wrapping the data points in a "values" array (alongside
// start/end/step metadata), not a flat array.
func parseInsightEntries(raw []byte) ([]insightEntry, error) {
	var resp struct {
		Values []insightEntry `json:"values"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	return resp.Values, nil
}

// resolveInsightLog finds a log by its full ID and returns the path parameters
// for the insights API. The Homey API looks up logs by their full ID
// (homey:device:<uuid>:<capability>), so that is what must be sent as the :id
// path segment — sending ownerId (just the capability) yields a 404.
func resolveInsightLog(logs []InsightLog, logID string) (uri, id, title string, found bool) {
	for _, log := range logs {
		if log.ID == logID {
			return log.OwnerURI, log.ID, log.Title, true
		}
	}
	return "", "", "", false
}

var insightsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all insight logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := apiClient.GetInsights()
		if err != nil {
			return err
		}

		if isTableFormat() {
			var logs []InsightLog
			if err := json.Unmarshal(data, &logs); err != nil {
				return fmt.Errorf("failed to parse insights: %w", err)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "TITLE\tTYPE\tUNITS\tID")
			fmt.Fprintln(w, "-----\t----\t-----\t--")
			for _, l := range logs {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", l.Title, l.Type, l.Units, l.ID)
			}
			w.Flush()
			return nil
		}

		outputJSON(data)
		return nil
	},
}

var insightsGetCmd = &cobra.Command{
	Use:   "get <log-id>",
	Short: "Get insight log entries",
	Long: `Get historical data entries for an insight log.

The log-id is from 'homeyctl insights list' output.
Example: homey:device:abc123:measure_power

Resolutions:
  - last24Hours (default)
  - lastWeek
  - lastMonth
  - lastYear
  - last2Years

Examples:
  homeyctl insights get "homey:device:abc123:measure_power"
  homeyctl insights get "homey:device:abc123:measure_power" --resolution lastWeek`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logID := args[0]
		resolution, _ := cmd.Flags().GetString("resolution")

		// First, look up the log to get ownerUri and ownerId
		data, err := apiClient.GetInsights()
		if err != nil {
			return err
		}

		var logs []InsightLog
		if err := json.Unmarshal(data, &logs); err != nil {
			return fmt.Errorf("failed to parse insights: %w", err)
		}

		ownerURI, id, _, found := resolveInsightLog(logs, logID)
		if !found {
			return fmt.Errorf("log not found: %s\nUse 'homeyctl insights list' to see available logs", logID)
		}

		entries, err := apiClient.GetInsightEntries(ownerURI, id, resolution)
		if err != nil {
			return err
		}

		if isTableFormat() {
			entryList, err := parseInsightEntries(entries)
			if err != nil {
				return fmt.Errorf("failed to parse entries: %w", err)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "TIME\tVALUE")
			fmt.Fprintln(w, "----\t-----")
			for _, e := range entryList {
				fmt.Fprintf(w, "%s\t%v\n", e.T.Local().Format("2006-01-02 15:04"), e.V)
			}
			w.Flush()
			return nil
		}

		outputJSON(entries)
		return nil
	},
}

var insightsDeleteCmd = &cobra.Command{
	Use:   "delete <log-id>",
	Short: "Delete an insight log",
	Long: `Delete an insight log and all its historical data.

Examples:
  homeyctl insights delete "homey:device:abc123:measure_power"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logID := args[0]

		// Look up the log to get ownerUri and ownerId
		data, err := apiClient.GetInsights()
		if err != nil {
			return err
		}

		var logs []InsightLog
		if err := json.Unmarshal(data, &logs); err != nil {
			return fmt.Errorf("failed to parse insights: %w", err)
		}

		ownerURI, id, title, found := resolveInsightLog(logs, logID)
		if !found {
			return fmt.Errorf("log not found: %s", logID)
		}

		if err := apiClient.DeleteInsightLog(ownerURI, id); err != nil {
			return err
		}

		fmt.Printf("Deleted insight log: %s\n", title)
		return nil
	},
}

var insightsClearCmd = &cobra.Command{
	Use:   "clear <log-id>",
	Short: "Clear insight log entries",
	Long: `Clear all historical data from an insight log without deleting the log itself.

Examples:
  homeyctl insights clear "homey:device:abc123:measure_power"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logID := args[0]

		// Look up the log to get ownerUri and ownerId
		data, err := apiClient.GetInsights()
		if err != nil {
			return err
		}

		var logs []InsightLog
		if err := json.Unmarshal(data, &logs); err != nil {
			return fmt.Errorf("failed to parse insights: %w", err)
		}

		ownerURI, id, title, found := resolveInsightLog(logs, logID)
		if !found {
			return fmt.Errorf("log not found: %s", logID)
		}

		if err := apiClient.DeleteInsightLogEntries(ownerURI, id); err != nil {
			return err
		}

		fmt.Printf("Cleared entries for: %s\n", title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(insightsCmd)
	insightsCmd.AddCommand(insightsListCmd)
	insightsCmd.AddCommand(insightsGetCmd)
	insightsCmd.AddCommand(insightsDeleteCmd)
	insightsCmd.AddCommand(insightsClearCmd)

	insightsGetCmd.Flags().String("resolution", "last24Hours", "Resolution: last24Hours, lastWeek, lastMonth, lastYear, last2Years")
}
