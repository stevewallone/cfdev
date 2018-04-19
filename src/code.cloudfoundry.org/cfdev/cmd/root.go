package cmd

import (
	"code.cloudfoundry.org/cfdev/config"
	"github.com/spf13/cobra"
	analytics "gopkg.in/segmentio/analytics-go.v3"
)

func NewRoot(Exit chan struct{}, UI UI, Config config.Config, AnalyticsClient analytics.Client) *cobra.Command {
	root := &cobra.Command{Use: "cf"}
	// HideHelpFlag(root)
	// usageTemplate := strings.Replace(root.UsageTemplate(), "\n"+`Use "{{.CommandPath}} [command] --help" for more information about a command.`, "", -1)
	// root.SetUsageTemplate(usageTemplate)

	dev := &cobra.Command{
		Use:   "dev",
		Short: "Start and stop a single vm CF deployment running on your workstation",
	}
	root.AddCommand(dev)

	dev.AddCommand(NewBosh(Exit, UI, Config))
	dev.AddCommand(NewCatalog(UI, Config))
	dev.AddCommand(NewDownload(Exit, UI, Config))
	dev.AddCommand(NewStart(Exit, UI, Config, AnalyticsClient))
	dev.AddCommand(NewStop(&Config, AnalyticsClient))
	dev.AddCommand(NewTelemetry(UI, Config))

	return root
}
