package cmd

import (
	"github.com/lnsp/mattermost-informer/pkg/controller"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:  "mattermost-informer",
	Long: "Broadcast pod crashes to a Mattermost channel",
	Run: func(cmd *cobra.Command, args []string) {
		controller.Run()
	},
}

func Execute() {
	rootCmd.Execute()
}
