package process

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/resource"
)

func IsCFDevDInstalled(sockPath string, binPath string, expectedMD5 string) bool {
	currentMD5, err := resource.MD5(binPath)
	if err != nil {
		fmt.Println("failed to get md5 ", binPath)
		return false
	}
	if currentMD5 != expectedMD5 {
		fmt.Println("md5 didn't match")
		return false
	}
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		fmt.Println("failed to dial socket ", sockPath)
		return false
	}
	conn.Close()
	return true
}

func InstallCFDevD(cacheDir string) error {
	cfdevdPath := filepath.Join(cacheDir, "cfdevd")
	fmt.Println("Installing networking components (requires root privileges)")
	cmd := exec.Command("sudo", "--stdin", cfdevdPath, "install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
