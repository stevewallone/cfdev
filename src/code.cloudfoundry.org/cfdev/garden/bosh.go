package garden

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/garden"
)

func DeployBosh(Config config.Config, client garden.Client, dockerRegistries []string) error {
	if unmount, err := mntCfDeps(Config); err != nil {
		return fmt.Errorf("mounting cf-deps.iso: %s", err)
	} else {
		defer unmount()
	}

	// _ = os.MkdirAll(filepath.Join(Config.CFDevHome, "vcap", "director"), 0755)
	// _ = os.MkdirAll(filepath.Join(Config.CFDevHome, "vcap", "store"), 0755)
	_ = os.MkdirAll("/var/vcap/director", 0755)
	_ = os.MkdirAll("/var/vcap/store", 0755)

	containerSpec := garden.ContainerSpec{
		Handle:     "deploy-bosh",
		Privileged: true,
		Network:    "10.246.0.0/16",
		Image: garden.ImageRef{
			URI: filepath.Join(Config.CFDevHome, "cache", "cf-deps", "workspace.tar"),
		},
		BindMounts: []garden.BindMount{
			{
				SrcPath: "/var/vcap/director", // filepath.Join(Config.CFDevHome, "vcap", "director"),
				DstPath: "/var/vcap/director",
				Mode:    garden.BindMountModeRW,
			},
			{
				SrcPath: "/var/vcap/store", // filepath.Join(Config.CFDevHome, "vcap", "store"),
				DstPath: "/var/vcap/store",
				Mode:    garden.BindMountModeRW,
			},
			{
				SrcPath: filepath.Join(Config.CFDevHome, "cache", "cf-deps"),
				DstPath: "/var/vcap/cache",
				Mode:    garden.BindMountModeRO,
			},
			{
				SrcPath: filepath.Join(Config.CFDevHome, "gdn.socket"),
				DstPath: "/var/vcap/gdn.socket",
				Mode:    garden.BindMountModeRO,
			},
		},
	}

	container, err := client.Create(containerSpec)
	if err != nil {
		return err
	}

	if err := copyFileToContainer(container, "/home/dgodd/workspace/cfdev/images/cf/allow-mounting", "/usr/bin/allow-mounting"); err != nil {
		return err
	}
	if err := copyFileToContainer(container, "/home/dgodd/workspace/cfdev/images/cf/deploy-bosh", "/usr/bin/deploy-bosh"); err != nil {
		return err
	}
	if err := copyFileToContainer(container, "/home/dgodd/workspace/cfdev/images/cf/bosh-operations/use_gdn_unix_socket.yml", "/var/vcap/use_gdn_unix_socket.yml"); err != nil {
		return err
	}
	if err := copyFileToContainer(container, "/home/dgodd/workspace/cfdev/images/cf/bosh-operations/use_gdn_1_12_1.yml", "/var/vcap/use_gdn_1_12_1.yml"); err != nil {
		return err
	}

	if err := runInContainer(container, "allow-mounting", "/usr/bin/allow-mounting"); err != nil {
		return err
	}
	// TODO copy back to workspace.tar // "/usr/bin/deploy-bosh",
	if err := runInContainer(container, "deploy-bosh", "/usr/bin/deploy-bosh"); err != nil {
		return err
	}

	client.Destroy("deploy-bosh")

	return nil
}

func mntCfDeps(Config config.Config) (func() error, error) {
	if err := os.MkdirAll(filepath.Join(Config.CFDevHome, "cache", "cf-deps"), 0755); err != nil {
		return nil, err
	}
	currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}
	if err := exec.Command("sudo", "mount", "-o", "loop,ro,uid="+currentUser.Username, filepath.Join(Config.CFDevHome, "cache", "cf-deps.iso"), filepath.Join(Config.CFDevHome, "cache", "cf-deps")).Run(); err != nil {
		return nil, err
	}
	return exec.Command("sudo", "umount", filepath.Join(Config.CFDevHome, "cache", "cf-deps")).Run, nil
}
