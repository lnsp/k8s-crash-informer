package cmd

import (
	"github.com/lnsp/k8s-mattermost-informer/pkg/controller"
	"github.com/lnsp/k8s-mattermost-informer/pkg/informer"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

var rootCmd = &cobra.Command{
	Use:  "mattermost-informer",
	Long: "Broadcast pod crashes to a Mattermost channel",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			chat informer.Informer
			err  error
		)
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

// Target is the notification target of this informer service.
var Target string

func Execute() {
	rootCmd.Execute()
	rootCmd.Flags().StringVarP(&Target, "target", "t", "mattermost", "Notification target of informer service")
}
