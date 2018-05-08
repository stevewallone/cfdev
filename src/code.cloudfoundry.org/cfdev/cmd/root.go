package cmd

import (
	"strings"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/process"
	cfdevdClient "code.cloudfoundry.org/cfdevd/client"
	"github.com/spf13/cobra"
)

func NewRoot(exit chan struct{}, ui UI, config config.Config, launchd Launchd) *cobra.Command {
	root := &cobra.Command{Use: "cf", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().Bool("help", false, "")
	root.PersistentFlags().Lookup("help").Hidden = true

	usageTemplate := strings.Replace(root.UsageTemplate(), "\n"+`Use "{{.CommandPath}} [command] --help" for more information about a command.`, "", -1)
	root.SetUsageTemplate(usageTemplate)

	dev := &cobra.Command{
		Use:           "dev",
		Short:         "Start and stop a single vm CF deployment running on your workstation",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(dev)

	version := Version{
		UI:     ui,
		Config: config,
	}

	start := Start{
		Exit:        exit,
		UI:          ui,
		Config:      config,
		Launchd:     launchd,
		ProcManager: &process.Manager{},
	}

	dev.AddCommand(version.Cmd())
	dev.AddCommand(NewBosh(exit, ui, config))
	dev.AddCommand(NewCatalog(ui, config))
	dev.AddCommand(NewDownload(exit, ui, config))
	dev.AddCommand(start.Cmd())
	dev.AddCommand(NewStop(config, launchd, cfdevdClient.New("CFD3V", config.CFDevDSocketPath), &process.Manager{}))
	dev.AddCommand(NewTelemetry(ui, config))
	dev.AddCommand(&cobra.Command{
		Use:   "help [command]",
		Short: "Help about any command",
		Run: func(c *cobra.Command, args []string) {
			cmd, _, _ := dev.Find(args)
			cmd.Help()
		},
	})

	return root
}
