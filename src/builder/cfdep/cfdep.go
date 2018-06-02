package cfdep

import (
	"builder/dep"
	"builder/fs"
	"builder/utils"
	"fmt"
	"path/filepath"
)

type CfDep struct {
	root            string
	cfdep           string
	stemcellVersion string
	configs         map[string]utils.Yaml
}

func New(root, cfdep, stemcellVersion string) *CfDep {
	return &CfDep{root: root, cfdep: cfdep, stemcellVersion: stemcellVersion, configs: make(map[string]utils.Yaml)}
}

func (cf *CfDep) ReadManifests() error {
	var err error
	cf.configs["runtime-config"], err = utils.BoshInt(
		filepath.Join(cf.root, "images/cf/configs/dns-runtime-config.yml"),
		nil, nil,
	)
	if err != nil {
		return err
	}
	cf.configs["cloud-config"], err = utils.BoshInt(
		filepath.Join(cf.cfdep, "iaas-support/bosh-lite/cloud-config.yml"),
		[]string{
			filepath.Join(cf.root, "images/cf/cf-operations", "set-cloud-config-subnet.yml"),
		},
		map[string]string{},
	)
	if err != nil {
		return err
	}
	cf.configs["deployment"], err = utils.BoshInt(
		filepath.Join(cf.cfdep, "cf-deployment.yml"),
		[]string{
			filepath.Join(cf.cfdep, "operations/use-compiled-releases.yml"),
			filepath.Join(cf.cfdep, "operations/experimental/skip-consul-cell-registrations.yml"),
			filepath.Join(cf.cfdep, "operations/experimental/skip-consul-locks.yml"),
			filepath.Join(cf.cfdep, "operations/experimental/use-bosh-dns-for-containers.yml"),
			filepath.Join(cf.cfdep, "operations/experimental/disable-consul.yml"),
			filepath.Join(cf.cfdep, "operations/bosh-lite.yml"),
			filepath.Join(cf.cfdep, "operations/experimental/disable-consul-bosh-lite.yml"),
			filepath.Join(cf.root, "images/cf/cf-operations", "garden-disable-app-armour.yml"),
			filepath.Join(cf.root, "images/cf/cf-operations", "collocate-tcp-router.yml"),
			filepath.Join(cf.root, "images/cf/cf-operations", "set-cfdev-subnet.yml"),
			filepath.Join(cf.root, "images/cf/cf-operations", "lower-memory.yml"),
		},
		map[string]string{
			"cf_admin_password":       "admin",
			"uaa_admin_client_secret": "admin-client-secret",
		},
	)
	return err
}

func (cf *CfDep) Configs() map[string]utils.Yaml {
	return cf.configs
}

func Build(fs *fs.Dir, rootDir, cfDeployment, stemcellVersion string) error {
	cf := New(rootDir, cfDeployment, stemcellVersion)
	if err := cf.ReadManifests(); err != nil {
		return fmt.Errorf("read manifest: %s", err)
	}
	if err := dep.DownloadReleases(stemcellVersion, cf.Configs(), fs); err != nil {
		return fmt.Errorf("download releases: %s", err)
	}
	if err := dep.WriteManifests(cf.Configs(), fs); err != nil {
		return fmt.Errorf("write manifest: %s", err)
	}
	if err := fs.AddFile("bin/deploy-cf", filepath.Join(rootDir, "images/cf/deploy-cf")); err != nil {
		return fmt.Errorf("copy file: %s", err)
	}
	if err := fs.AddFile("app-security-group.json", filepath.Join(rootDir, "images/cf/app-security-group.json")); err != nil {
		return fmt.Errorf("copy file: %s", err)
	}
	return nil
}
