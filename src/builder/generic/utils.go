package generic

import (
	"builder/fs"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func BuildWorkspaceTar(fs *fs.Dir, rootDir string) error {
	if txt, err := exec.Command(filepath.Join(rootDir, "images/cf/build.sh")).CombinedOutput(); err != nil {
		// TODO inline code from build.sh (and delete original)
		fmt.Println(string(txt))
		return fmt.Errorf("images/cf/build.sh: %s", err)
	}

	// Export
	cidBytes, err := exec.Command("docker", "run", "-d", "pivotal/cf", "sleep", "infinity").Output()
	if err != nil {
		return fmt.Errorf("export docker: run: %s", err)
	}
	cid := strings.TrimSpace(string(cidBytes))
	fmt.Printf("CID: |%s|\n", cid)
	fh, err := fs.Writer("workspace.tar")
	if err != nil {
		return err
	}
	cmd := exec.Command("docker", "export", cid)
	cmd.Stdout = fh
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("export docker: export %s: %s", cid, err)
	}
	if err := fh.Close(); err != nil {
		return err
	}
	if txt, err := exec.Command("docker", "kill", cid).CombinedOutput(); err != nil {
		fmt.Println(txt)
		return fmt.Errorf("export docker: kill %s: %s", cid, err)
	}
	if txt, err := exec.Command("docker", "rm", cid).CombinedOutput(); err != nil {
		fmt.Println(txt)
		return fmt.Errorf("export docker: rm %s: %s", cid, err)
	}
	return nil
}

func BuildCfDepsIso(cacheDir, outputDir string) error {
	if txt, err := exec.Command("mkisofs", "-V", "cf-deps", "-R", "-o", filepath.Join(outputDir, "cf-deps.iso"), cacheDir).CombinedOutput(); err != nil {
		fmt.Println(string(txt))
		return fmt.Errorf("mkisofs: %s", err)
	}
	return nil
}

func BuildEfiImage(linuxkit, linuxkitDir, outputDir string) error {
	for _, name := range []string{"bosh-lite-routing", "expose-multiple-ports", "garden-runc", "openssl"} {
		fmt.Println("RUN:", linuxkit, "pkg", "build", "-hash", "dev", filepath.Join(linuxkitDir, "pkg", name))
		if txt, err := exec.Command(linuxkit, "pkg", "build", "-hash", "dev", filepath.Join(linuxkitDir, "pkg", name)).CombinedOutput(); err != nil {
			fmt.Println(string(txt))
			return fmt.Errorf("linuxkit build %s: %s", name, err)
		}
	}

	fmt.Println("RUN:", linuxkit, "build", "-disable-content-trust", "-name", "cfdev", "-format", "iso-efi", "-dir", outputDir, filepath.Join(linuxkitDir, "base.yml"), filepath.Join(linuxkitDir, "garden.yml"))
	if txt, err := exec.Command(linuxkit, "build", "-disable-content-trust", "-name", "cfdev", "-format", "iso-efi", "-dir", outputDir, filepath.Join(linuxkitDir, "base.yml"), filepath.Join(linuxkitDir, "garden.yml")).CombinedOutput(); err != nil {
		fmt.Println(string(txt))
		return fmt.Errorf("linuxkit build: %s", err)
	}

	return nil
}
