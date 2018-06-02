package mysqldep

import (
	"builder/dep"
	"builder/fs"
	"builder/utils"
	"fmt"
	"path/filepath"
)

type MysqlDep struct {
	root            string
	mysqldep        string
	stemcellVersion string
	configs         map[string]utils.Yaml
}

func New(root, mysqldep, stemcellVersion string) *MysqlDep {
	return &MysqlDep{root: root, mysqldep: mysqldep, stemcellVersion: stemcellVersion, configs: make(map[string]utils.Yaml)}
}

func (mysql *MysqlDep) ReadManifests() error {
	var err error
	mysql.configs["mysql"], err = utils.BoshInt(
		filepath.Join(mysql.mysqldep, "cf-mysql-deployment.yml"),
		[]string{
			filepath.Join(mysql.mysqldep, "operations", "add-broker.yml"),
			filepath.Join(mysql.mysqldep, "operations", "register-proxy-route.yml"),
			filepath.Join(mysql.mysqldep, "operations", "no-arbitrator.yml"),
		},
		map[string]string{
			"cf_mysql_external_host": "p-mysql.v3.pcfdev.io",
			"cf_mysql_host":          "v3.pcfdev.io",
			"cf_admin_password":      "admin",
			"cf_api_url":             "https://api.v3.pcfdev.io",
			"cf_skip_ssl_validation": "true",
			"proxy_vm_extension":     "mysql-proxy-lb",
		},
	)
	return err
}

func (mysql *MysqlDep) Configs() map[string]utils.Yaml {
	return mysql.configs
}

func (mysql *MysqlDep) Finalize() error {
	// Set stemcell version
	if stemcells, ok := mysql.configs["mysql"]["stemcells"].([]interface{}); ok {
		for _, stemcell := range stemcells {
			if stemcell, ok := stemcell.(utils.Yaml); ok {
				if stemcell["alias"] == "default" {
					stemcell["version"] = mysql.stemcellVersion
				}
			}
		}
	}
	// Set all instance_groups to 1 instance (unless 0)
	if groups, ok := mysql.configs["mysql"]["instance_groups"].([]interface{}); ok {
		for _, group := range groups {
			if group, ok := group.(utils.Yaml); ok {
				if instances, ok := group["instances"].(int); ok {
					if instances > 1 {
						group["instances"] = 1
					}
				}
			}
		}
	}
	return nil
}

func Build(fs *fs.Dir, rootDir, mysqlDeployment, stemcellVersion string) error {
	mysql := New(rootDir, mysqlDeployment, stemcellVersion)
	if err := mysql.ReadManifests(); err != nil {
		return fmt.Errorf("read manifest: %s", err)
	}
	if err := dep.DownloadReleases(stemcellVersion, mysql.Configs(), fs); err != nil {
		return fmt.Errorf("download releases: %s", err)
	}
	if err := mysql.Finalize(); err != nil {
		return fmt.Errorf("finalize: %s", err)
	}
	if err := dep.WriteManifests(mysql.Configs(), fs); err != nil {
		return fmt.Errorf("write manifest: %s", err)
	}
	if err := fs.AddFile("bin/deploy-mysql", filepath.Join(rootDir, "images/cf/deploy-mysql")); err != nil {
		return fmt.Errorf("copy file: %s", err)
	}
	return nil
}
