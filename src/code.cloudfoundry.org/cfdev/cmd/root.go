package cmd

import (
	"strings"

	"code.cloudfoundry.org/cfdev/bosh"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/process"
	cfdevdClient "code.cloudfoundry.org/cfdevd/client"
	"github.com/spf13/cobra"
	"golang.org/x/text/message/catalog"
)

type cmdBuilder interface {
	Cmd() *cobra.Command
}

func NewRoot(exit chan struct{}, ui UI, config config.Config, launchd Launchd) *cobra.Command {
	root := &cobra.Command{Use: "cf", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().Bool("help", false, "")
	root.PersistentFlags().Lookup("help").Hidden = true

	usageTemplate := strings.Replace(root.UsageTemplate(), "\n"+`Use "{{.CommandPath}} [command] --help" for more information about a command.`, "", -1)
	root.SetUsageTemplate(usageTemplate)

	root.AddCommand(&cobra.Command{
		Use:           "dev",
		Short:         "Start and stop a single vm CF deployment running on your workstation",
		SilenceUsage:  true,
		SilenceErrors: true,
	})

	for _, cmd := range []cmdBuilder{
		version.Version{
			UI:     ui,
			Config: config,
		},
		bosh.Bosh{
			Exit:   exit,
			UI:     ui,
			Config: config,
		},
		catalog.Catalog{
			Exit:   exit,
			UI:     ui,
			Config: config,
		},
		download.download{
			Exit:   exit,
			UI:     ui,
			Config: config,
		},
		start.Start{
			Exit:        exit,
			UI:          ui,
			Config:      config,
			Launchd:     launchd,
			ProcManager: &process.Manager{},
		},
		stop.Stop{
			Config:       config,
			Launchd:      launchd,
			ProcManager:  &process.Manager{},
			CfdevdClient: cfdevdClient.New("CFD3V", config.CFDevDSocketPath),
		},
		telemetry.Telemetry{
			UI:     ui,
			Config: config,
		},
	} {
		dev.AddCommand(cmd.Cmd())
	}

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
