package garden

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/garden"
	"gopkg.in/yaml.v2"
)

func DeployCloudFoundry(Config config.Config, client garden.Client, dockerRegistries []string) error {
	if unmount, err := mntCfDeps(Config); err != nil {
		return fmt.Errorf("mounting cf-deps.iso: %s", err)
	} else {
		defer unmount()
	}

	containerSpec := garden.ContainerSpec{
		Handle:     "deploy-cf",
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
		},
	}

	if len(dockerRegistries) > 0 {
		bytes, err := yaml.Marshal(dockerRegistries)

		if err != nil {
			return err
		}

		containerSpec.Env = append(containerSpec.Env, "DOCKER_REGISTRIES="+string(bytes))
	}

	container, err := client.Create(containerSpec)
	if err != nil {
		return err
	}

	if err := copyFileToContainer(container, "/home/dgodd/workspace/cfdev/images/cf/allow-mounting", "/usr/bin/allow-mounting"); err != nil {
		return err
	}
	if err := copyFileToContainer(container, "/home/dgodd/workspace/cfdev/images/cf/deploy-cf", "/usr/bin/deploy-cf"); err != nil {
		return err
	}

	if err := runInContainer(container, "allow-mounting", "/usr/bin/allow-mounting"); err != nil {
		return err
	}
	if err := runInContainer(container, "deploy-cf", "/usr/bin/deploy-cf"); err != nil {
		return err
	}

	client.Destroy("deploy-cf")

	return nil
}

func runInContainer(container garden.Container, id, path string, args ...string) error {
	fmt.Printf("DG: About to run %s: %s %v\n", id, path, args)
	process, err := container.Run(garden.ProcessSpec{
		ID:   id,
		Path: path,
		Args: args,
		User: "root",
	}, garden.ProcessIO{
		// TODO write to file instead
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	if err != nil {
		return err
	}
	exitCode, err := process.Wait()
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("process exited with status %v", exitCode)
	}
	return nil
}

func copyFileToContainer(container garden.Container, src, dest string) error {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	txt, _ := ioutil.ReadFile(src)
	tw.WriteHeader(&tar.Header{Name: filepath.Base(dest), Mode: 0755, Size: int64(len(txt))})
	tw.Write(txt)
	tw.Close()
	return container.StreamIn(garden.StreamInSpec{
		Path:      filepath.Dir(dest),
		User:      "root",
		TarStream: &buf,
	})
}
