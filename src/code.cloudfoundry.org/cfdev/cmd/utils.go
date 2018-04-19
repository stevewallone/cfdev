package cmd

import "github.com/spf13/cobra"

func HideHelpFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool("help", false, "")
	cmd.PersistentFlags().Lookup("help").Hidden = true
}
