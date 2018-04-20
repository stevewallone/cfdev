package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/env"
	gdn "code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/garden/client"
	"github.com/spf13/cobra"
	analytics "gopkg.in/segmentio/analytics-go.v3"
)

type UI interface {
	Say(message string, args ...interface{})
}

type start struct {
	Exit            chan struct{}
	UI              UI
	Config          config.Config
	AnalyticsClient analytics.Client
	Registries      string
	gdnServer       *exec.Cmd
}

func NewStart(Exit chan struct{}, UI UI, Config config.Config, AnalyticsClient analytics.Client) *cobra.Command {
	s := start{Exit: Exit, UI: UI, Config: Config, AnalyticsClient: AnalyticsClient}
	cmd := &cobra.Command{
		Use: "start",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := s.RunE(); err != nil {
				UI.Say("Failed to start cfdev: %v\n", err)
			}
			return nil
		},
	}
	pf := cmd.PersistentFlags()
	pf.StringVar(&s.Registries, "r", "", "docker registries that skip ssl validation - ie. host:port,host2:port2")

	return cmd
}

func (s *start) RunE() error {
	go func() {
		<-s.Exit
		if s.gdnServer != nil {
			s.gdnServer.Process.Kill()
		}
		os.Exit(128)
	}()

	cfanalytics.TrackEvent(cfanalytics.START_BEGIN, map[string]interface{}{"type": "cf"}, s.AnalyticsClient)

	if err := env.Setup(s.Config); err != nil {
		return err
	}

	// TODO test existence of /sys/fs/cgroup/memory/memory.memsw.limit_in_bytes
	// if not exist then tell user to add GRUB_CMDLINE_LINUX_DEFAULT="cgroup_enable=memory swapaccount=1"
	// to /etc/default/grub and run sudo grub-update && sudo reboot (or update-grub)

	// TODO should this be the same on linux as on darwin????
	registries, err := s.parseDockerRegistriesFlag(s.Registries)
	if err != nil {
		return fmt.Errorf("Unable to parse docker registries %v\n", err)
	}

	garden := gdn.NewClient(s.Config)
	if garden.Ping() == nil {
		s.UI.Say("CF Dev is already running...")
		cfanalytics.TrackEvent(cfanalytics.START_END, map[string]interface{}{"type": "cf", "alreadyrunning": true}, s.AnalyticsClient)
		return nil
	}

	s.UI.Say("Downloading Resources...")
	if err := download(s.Config.Dependencies, s.Config.CacheDir); err != nil {
		return err
	}

	s.UI.Say("Starting Garden Server...")
	if err := s.startGarden(garden); err != nil {
		return fmt.Errorf("Unable to start garden server %v\n", err)
	}

	s.UI.Say("Deploying the BOSH Director...")
	if err := gdn.DeployBosh(s.Config, garden, registries); err != nil {
		fmt.Printf("Failed to deploy the BOSH Director: %v\n", err)
		return fmt.Errorf("Failed to deploy the BOSH Director: %v\n", err)
	}

	// TODO we need to do `sudo route add -host 10.144.0.34 gw 10.245.0.2`
	_ = exec.Command("sudo", "route", "add", "-host", "10.144.0.34", "gw 10.245.0.2").Run()

	s.UI.Say("Deploying CF...")
	if err := gdn.DeployCloudFoundry(garden, registries); err != nil {
		return fmt.Errorf("Failed to deploy the Cloud Foundry: %v\n", err)
	}

	s.UI.Say(cfdevStartedMessage)

	cfanalytics.TrackEvent(cfanalytics.START_END, map[string]interface{}{"type": "cf"}, s.AnalyticsClient)

	return nil
}

func (s *start) startGarden(garden client.Client) error {
	// TODO download gdn cli
	// TODO Inform user they need xfsprogs
	fh, err := os.Create(filepath.Join(s.Config.CFDevHome, "gdn.server.log"))
	if err != nil {
		return err
	}
	// Add to below? --dns-server=8.8.8.8
	s.gdnServer = exec.Command("sudo", filepath.Join(s.Config.CFDevHome, "cache", "gdn"), "server", "--bind-socket="+filepath.Join(s.Config.CFDevHome, "gdn.socket"), "--dns-server=8.8.8.8")
	s.gdnServer.Stdout = fh
	s.gdnServer.Stderr = fh
	s.gdnServer.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	// TODO the below is a terrible way to get password
	s.gdnServer.Env = append(os.Environ(), "SUDO_ASKPASS=/tmp/askpass")
	if err := s.gdnServer.Start(); err != nil {
		return fmt.Errorf("starting garden: %s", err)
	}
	fmt.Println("DEBUG: Waiting for Garden")
	if err := gdn.WaitForGarden(garden, 2*time.Second); err != nil {
		// s.UI.Say(buf.String())
		s.gdnServer.Process.Kill()
		return fmt.Errorf("starting garden: %s", err)
	}
	if err := ioutil.WriteFile(filepath.Join(s.Config.CFDevHome, "garden.pid"), []byte(fmt.Sprintf("%d", s.gdnServer.Process.Pid)), 0644); err != nil {
		s.gdnServer.Process.Kill()
		return fmt.Errorf("writing garden pid file: %s", err)
	}
	return nil
}
