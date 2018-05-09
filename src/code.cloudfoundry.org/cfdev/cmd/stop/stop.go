package stop

import (
	"path/filepath"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/process"
	"github.com/spf13/cobra"
)

type LaunchdStop interface {
	Stop(label string) error
}

type CfdevdClient interface {
	Uninstall() (string, error)
}

type ProcManager interface {
	SafeKill(string, string) error
}

type UI interface {
	Say(message string, args ...interface{})
}

type Stop struct {
	Config       config.Config
	Launchd      LaunchdStop
	ProcManager  ProcManager
	CfdevdClient CfdevdClient
}

func (s *Stop) Cmd() *cobra.Command {
	return &cobra.Command{
		Use:  "stop",
		RunE: s.RunE,
	}
}

func (s *Stop) RunE(cmd *cobra.Command, args []string) error {
	s.Config.Analytics.Event(cfanalytics.STOP, map[string]interface{}{"type": "cf"})

	var reterr error

	if err := s.Launchd.Stop(process.LinuxKitLabel); err != nil {
		reterr = errors.SafeWrap(err, "failed to stop linuxkit")
	}

	if err := s.Launchd.Stop(process.VpnKitLabel); err != nil {
		reterr = errors.SafeWrap(err, "failed to stop vpnkit")
	}

	if err := s.ProcManager.SafeKill(filepath.Join(s.Config.StateDir, "hyperkit.pid"), "hyperkit"); err != nil {
		reterr = errors.SafeWrap(err, "failed to kill hyperkit")
	}

	if _, err := s.CfdevdClient.Uninstall(); err != nil {
		reterr = errors.SafeWrap(err, "failed to uninstall cfdevd")
	}

	if reterr != nil {
		return errors.SafeWrap(reterr, "cf dev stop")
	}
	return nil
}
