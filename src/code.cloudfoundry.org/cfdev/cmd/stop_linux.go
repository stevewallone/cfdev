package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"code.cloudfoundry.org/cfdev/config"
	gdn "code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/cfdev/process"
	"github.com/spf13/cobra"
	analytics "gopkg.in/segmentio/analytics-go.v3"
)

func NewStop(Config *config.Config, AnalyticsClient analytics.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use: "stop",
		RunE: func(cmd *cobra.Command, args []string) error {
			garden := gdn.NewClient(*Config)
			containers, err := garden.Containers(nil)
			if err != nil {
				return nil
			}
			for _, container := range containers {
				fmt.Printf("Delete Container: %s\n", container.Handle())
				if err := garden.Destroy(container.Handle()); err != nil {
					return nil
				}
			}

			fmt.Println("Cleanup state directories")
			for _, dir := range []string{"cf", "cfdev_cache", "director", "store"} {
				os.RemoveAll(filepath.Join("/var/vcap", dir))
			}

			fmt.Println("Stop Garden")
			if err := process.SignalAndCleanup(filepath.Join(Config.CFDevHome, "garden.pid"), filepath.Join(Config.CFDevHome, "gdn.socket"), syscall.SIGTERM); err != nil {
				return fmt.Errorf("try using sudo ; failed to terminate garden: %s", err)
			}
			return nil
		},
	}
	return cmd
}
