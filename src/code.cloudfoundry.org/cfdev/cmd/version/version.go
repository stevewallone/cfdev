package cmd

import (
	"code.cloudfoundry.org/cfdev/config"
	"github.com/spf13/cobra"
)

type Version struct {
	UI     UI
	Config config.Config
}

func (v *Version) Run() {
	v.UI.Say("Version: %s", v.Config.CliVersion.Original)
}

func (v *Version) Cmd() *cobra.Command {
	return &cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {
			v.Run()
		},
	}
}
