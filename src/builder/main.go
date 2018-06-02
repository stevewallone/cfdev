package main

import (
	"builder/boshdep"
	"builder/cfdep"
	"builder/dep"
	"builder/fs"
	"builder/generic"
	"builder/mysqldep"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	var rootDir, cacheDir, outputDir, cfDeployment, boshDeployment, mysqlDeployment, linuxkit string
	var skipBuildEfiImage, onlyBuildEfiImage bool

	flag.StringVar(&rootDir, "cfdev-repo", "", "root directory of cfdev repo")
	flag.StringVar(&cacheDir, "cache", filepath.Join(rootDir, "output/cache"), "directory to cache output")
	flag.StringVar(&outputDir, "output", filepath.Join(rootDir, "output"), "directory for output files")
	flag.StringVar(&cfDeployment, "cf-deployment", "../cf-deployment", "input directory for cf-deployment")
	flag.StringVar(&boshDeployment, "bosh-deployment", "../bosh-deployment", "input directory for bosh-deployment")
	flag.StringVar(&mysqlDeployment, "mysql-deployment", "../cf-mysql-deployment", "input directory for cf-mysql-deployment")
	flag.BoolVar(&skipBuildEfiImage, "skip-build-efi-image", false, "skip building cfdev-efi.iso")
	flag.BoolVar(&onlyBuildEfiImage, "only-build-efi-image", false, "only build cfdev-efi.iso")
	flag.StringVar(&linuxkit, "linuxkit", "linuxkit", "path to linuxkit")
	flag.Parse()

	if rootDir == "" {
		var err error
		rootDir, err = os.Getwd()
		if err != nil {
			panic(err)
		}
	}

	if !onlyBuildEfiImage {
		stemcellVersion, err := dep.StemcellVersion(filepath.Join(cfDeployment, "cf-deployment.yml"))
		if err != nil {
			panic(fmt.Errorf("stemcellVersion: %s", err))
		}

		fs, err := fs.New(cacheDir)
		if err != nil {
			panic(err)
		}

		fmt.Println("=== DOWNLOAD STEMCELL")
		if err := dep.DownloadStemcell(stemcellVersion, fs); err != nil {
			panic(fmt.Errorf("download stemcell: %s", err))
		}

		fmt.Println("=== GENERATE/DOWNLOAD BOSH DEPS")
		if err := boshdep.Build(fs, rootDir, boshDeployment, stemcellVersion); err != nil {
			panic(err)
		}

		fmt.Println("=== GENERATE/DOWNLOAD CF DEPS")
		if err := cfdep.Build(fs, rootDir, cfDeployment, stemcellVersion); err != nil {
			panic(err)
		}

		fmt.Println("=== GENERATE/DOWNLOAD MYSQL DEPS")
		if err := mysqldep.Build(fs, rootDir, mysqlDeployment, stemcellVersion); err != nil {
			panic(err)
		}

		fmt.Println("=== GENERATE WORKSPACE.TAR")
		if err := generic.BuildWorkspaceTar(fs, rootDir); err != nil {
			panic(err)
		}

		if err := fs.DeleteOld(); err != nil {
			panic(err)
		}

		fmt.Println("=== GENERATE CF.ISO")
		if err := generic.BuildCfDepsIso(cacheDir, outputDir); err != nil {
			panic(err)
		}
	}

	if !skipBuildEfiImage {
		fmt.Println("=== BUILD EFI IMAGE")
		if err := generic.BuildEfiImage(linuxkit, filepath.Join(rootDir, "linuxkit"), outputDir); err != nil {
			panic(err)
		}
	}

	fmt.Println("\n\nNOW, PLEASE GENERATE CF PLUGIN VIA:", filepath.Join(rootDir, "src/code.cloudfoundry.org/cfdev/generate-plugin.sh"))
}
