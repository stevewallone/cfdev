package network

import (
	"fmt"
	"net"
	"os"
	"os/exec"

	"code.cloudfoundry.org/cfdev/resource"
)

type CFDevD struct {
	ExecutablePath string
	TimeSyncSocket string
}

func IsCFDevDInstalled(sockPath string, binPath string, expectedMD5 string) bool {
	currentMD5, err := resource.MD5(binPath)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println("failed to get md5 ", binPath)
		}
		return false
	}
	if currentMD5 != expectedMD5 {
		return false
	}
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (c *CFDevD) Install() error {
	fmt.Println("Installing networking components (requires root privileges)")
	cmd := exec.Command("sudo", "-S",
		c.ExecutablePath,
		"install",
		"--timesyncSock", c.TimeSyncSocket,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
