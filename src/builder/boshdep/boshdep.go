package boshdep

import (
	"builder/dep"
	"builder/fs"
	"builder/utils"
	"fmt"
	"path/filepath"
)

type BoshDep struct {
	root            string
	boshdep         string
	stemcellVersion string
	configs         map[string]utils.Yaml
}

func New(root, boshdep, stemcellVersion string) *BoshDep {
	return &BoshDep{root: root, boshdep: boshdep, stemcellVersion: stemcellVersion, configs: make(map[string]utils.Yaml)}
}

func (bosh *BoshDep) ReadManifests() error {
	var err error
	// TODO ??? credhub and uaa (no longer required)
	bosh.configs["director"], err = utils.BoshInt(
		filepath.Join(bosh.boshdep, "bosh.yml"),
		[]string{
			filepath.Join(bosh.boshdep, "bosh-lite.yml"),
			filepath.Join(bosh.boshdep, "bosh-lite-runc.yml"),
			filepath.Join(bosh.boshdep, "bosh-lite-grootfs.yml"),
			filepath.Join(bosh.boshdep, "warden/cpi.yml"),
			filepath.Join(bosh.boshdep, "warden/cpi-grootfs.yml"),
			filepath.Join(bosh.boshdep, "jumpbox-user.yml"),
			filepath.Join(bosh.root, "images/cf/bosh-operations", "disable-app-armor.yml"),
			filepath.Join(bosh.root, "images/cf/bosh-operations", "remove-ports.yml"),
			filepath.Join(bosh.root, "images/cf/bosh-operations", "use-warden-cpi-v39.yml"),
		},
		map[string]string{
			"director_name": "warden",
			"internal_cidr": "10.245.0.0/24",
			"internal_gw":   "10.245.0.1",
			"internal_ip":   "10.245.0.2",
			"garden_host":   "10.0.0.10",
		},
	)
	if err != nil {
		return err
	}
	bosh.configs["dns"], err = utils.BoshInt(
		filepath.Join(bosh.boshdep, "runtime-configs/dns.yml"),
		[]string{
			filepath.Join(bosh.root, "images/cf/bosh-operations", "add-host-pcfdev-dns-record.yml"),
		},
		nil,
	)
	return err
}

func (bosh *BoshDep) Configs() map[string]utils.Yaml {
	return bosh.configs
}

func (bosh *BoshDep) Finalize() error {
	stemcellFilename := dep.StemcellFilename(bosh.stemcellVersion)
	if pools, ok := bosh.configs["director"]["resource_pools"].([]interface{}); ok {
		for _, pool := range pools {
			if pool, ok := pool.(utils.Yaml); ok {
				if stemcell, ok := pool["stemcell"].(utils.Yaml); ok {
					stemcell["url"] = fmt.Sprintf("file:///var/vcap/cache/%s", stemcellFilename)
					delete(stemcell, "sha1")
				}
			}
		}
	}
	return nil
}

func Build(fs *fs.Dir, rootDir, boshDeployment, stemcellVersion string) error {
	bosh := New(rootDir, boshDeployment, stemcellVersion)
	if err := bosh.ReadManifests(); err != nil {
		return fmt.Errorf("read manifest: %s", err)
	}
	if err := dep.DownloadReleases(stemcellVersion, bosh.Configs(), fs); err != nil {
		return fmt.Errorf("download releases: %s", err)
	}
	if err := bosh.Finalize(); err != nil {
		return fmt.Errorf("finalize: %s", err)
	}
	if err := dep.WriteManifests(bosh.Configs(), fs); err != nil {
		return fmt.Errorf("write manifest: %s", err)
	}
	if err := fs.AddFile("bin/deploy-bosh", filepath.Join(rootDir, "images/cf/deploy-bosh")); err != nil {
		return fmt.Errorf("copy file: %s", err)
	}
	return nil
}
