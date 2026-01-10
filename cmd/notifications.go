package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var notifyCmd = &cobra.Command{
	Use:     "notify",
	Aliases: []string{"notifications"},
	Short:   "Manage notifications",
	Long:    `Send and view Homey timeline notifications.`,
}

var notifySendCmd = &cobra.Command{
	Use:   "send <message>",
	Short: "Send a notification to the timeline",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		message := args[0]

		if err := apiClient.SendNotification(message); err != nil {
			return err
		}

		fmt.Printf("Notification sent: %s\n", message)
		return nil
	},
}

var notifyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List timeline notifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := apiClient.GetNotifications()
		if err != nil {
			return err
		}

		outputJSON(data)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(notifyCmd)
	notifyCmd.AddCommand(notifySendCmd)
	notifyCmd.AddCommand(notifyListCmd)
}
