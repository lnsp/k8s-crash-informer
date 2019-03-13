package cmd

import (
	"github.com/lnsp/k8s-crash-informer/pkg/controller"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:  "crash-informer",
	Long: "Broadcast pod crashes to a Mattermost channel",
	Run: func(cmd *cobra.Command, args []string) {
		controller.Run()
	},
}

func Execute() {
	rootCmd.Execute()
}
