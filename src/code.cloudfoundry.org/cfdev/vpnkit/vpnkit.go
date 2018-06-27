package vpnkit

import (
	"net"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/process"

	"path"
	"code.cloudfoundry.org/cfdev/env"
	"encoding/json"
	"os"
	"io/ioutil"
	"code.cloudfoundry.org/cfdevd/launchd"
)

type Launchd interface {
	AddDaemon(launchd.DaemonSpec) error
	Start(label string) error
}

const retries = 5

type Vpnkit struct {
	Config  config.Config
	Launchd Launchd
}

const VpnKitLabel = "org.cloudfoundry.cfdev.vpnkit"

func (v *Vpnkit) Start() error {
	if err := v.setupVPNKit(); err != nil {
		return errors.SafeWrap(err, "Failed to setup VPNKit")
	}
	if err := v.Launchd.AddDaemon(v.daemonSpec()); err != nil {
		return errors.SafeWrap(err, "install vpnkit")
	}
	if err := v.Launchd.Start(process.VpnKitLabel); err != nil {
		return errors.SafeWrap(err, "start vpnkit")
	}
	attempt := 0
	for {
		conn, err := net.Dial("unix", filepath.Join(v.Config.VpnkitStateDir, "vpnkit_eth.sock"))
		if err == nil {
			conn.Close()
			return nil
		} else if attempt >= retries {
			return errors.SafeWrap(err, "conenct to vpnkit")
		} else {
			time.Sleep(time.Second)
			attempt++
		}
	}
}

func (v *Vpnkit) daemonSpec() launchd.DaemonSpec {
	return launchd.DaemonSpec{
		Label:       VpnKitLabel,
		Program:     path.Join(v.Config.CacheDir, "vpnkit"),
		SessionType: "Background",
		ProgramArguments: []string{
			path.Join(v.Config.CacheDir, "vpnkit"),
			"--ethernet", path.Join(v.Config.VpnkitStateDir, "vpnkit_eth.sock"),
			"--port", path.Join(v.Config.VpnkitStateDir, "vpnkit_port.sock"),
			"--vsock-path", path.Join(v.Config.StateDir, "connect"),
			"--http", path.Join(v.Config.VpnkitStateDir, "http_proxy.json"),
			"--host-names", "host.pcfdev.io",
		},
		RunAtLoad:  false,
		StdoutPath: path.Join(v.Config.CFDevHome, "vpnkit.stdout.log"),
		StderrPath: path.Join(v.Config.CFDevHome, "vpnkit.stderr.log"),
	}
}

func (v *Vpnkit) setupVPNKit() error {
	httpProxyPath := filepath.Join(v.Config.VpnkitStateDir, "http_proxy.json")

	proxyConfig := env.BuildProxyConfig(v.Config.BoshDirectorIP, v.Config.CFRouterIP)
	proxyContents, err := json.Marshal(proxyConfig)
	if err != nil {
		return errors.SafeWrap(err, "Unable to create proxy config")
	}

	if _, err := os.Stat(httpProxyPath); !os.IsNotExist(err) {
		err = os.Remove(httpProxyPath)
		if err != nil {
			return errors.SafeWrap(err, "Unable to remove 'http_proxy.json'")
		}
	}

	httpProxyConfig := []byte(proxyContents)
	err = ioutil.WriteFile(httpProxyPath, httpProxyConfig, 0777)
	if err != nil {
		return err
	}
	return nil
}
