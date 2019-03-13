package cmd

import (
	"github.com/lnsp/k8s-mattermost-informer/pkg/controller"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:  "mattermost-informer",
	Long: "Broadcast pod crashes to a Mattermost channel",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			chat informer.Informer
			err error
		)
		var chat informer.Inforner
		switch Target {
		case "mattermost":
			chat, err = informer.NewMattermostInformerFromEnv()
		case "slack":
			chat, err = informer.NewSlackInformerFromEnv()
		}
		if err != nil {
			klog.Errorf("Failed to setup informer: %v", err)
		}
		controller.Run(chat)
	},
}

var Target string
func Execute() {
	rootCmd.Execute()
	rootCmd.Flags().StringVarP(&Target, "target", "t", "mattermost", "set notification target system")
}
