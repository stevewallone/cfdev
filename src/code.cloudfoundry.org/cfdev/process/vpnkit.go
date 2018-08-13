package process

import "code.cloudfoundry.org/cfdev/config"

const VpnKitLabel = "org.cloudfoundry.cfdev.vpnkit"

type VpnKit struct {
	Config  config.Config
	DaemonRunner DaemonRunner
}

func (v *VpnKit) Stop() error {
	return v.DaemonRunner.Stop(VpnKitLabel)
}

