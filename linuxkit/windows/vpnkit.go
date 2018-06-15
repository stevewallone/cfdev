package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	ps "github.com/gorillalabs/go-powershell"
	"github.com/gorillalabs/go-powershell/backend"
)

func main() {
	back := &backend.Local{}

	shell, err := ps.New(back)
	if err != nil {
		panic(err)
	}
	defer shell.Exit()

	stdout, _, err := shell.Execute("(Get-VM -Name 'cfdev11').Id")
	if err != nil {
		panic(err)
	}
	id := strings.Split(stdout, "\r\n")[3]
	fmt.Println("VM GUID:", id)

	cmd1 := exec.Command(`C:\Users\WX014\Desktop\vpnkit\vpnkit.exe`,
		"--ethernet", "hyperv-connect://"+id+"/30D48B34-7D27-4B0B-AAAF-BBBED334DD59",
		// "--port", "//./pipe/cfdevVpnKitControl",
		"--port", "hyperv-connect://"+id+"/C378280D-DA14-42C8-A24E-0DE92A1028E2",
		"--port", "hyperv-connect://"+id+"/0B95756A-9985-48AD-9470-78E060895BE7",
		"--dns", `C:\Users\WX014\AppData\Roaming\Docker\resolv.conf`,
		"--dhcp", `C:\Users\WX014\AppData\Roaming\Docker\dhcp.json`,
		"--diagnostics", `\\.\pipe\cfdevVpnKitDiagnostics`,
		"--host-names", "hostname",
		"--gateway-names", "gateway.cfdev.internal,cfdev.gateway.internal,cfdev.http.internal",
		"--listen-backlog", "32",
		"--lowest-ip", "169.254.82.3",
		"--highest-ip", "169.254.82.255",
		"--host-ip", "169.254.82.2",
		"--gateway-ip", "169.254.82.1")
	cmd1.Stdout = os.Stdout
	cmd1.Stderr = os.Stderr
	if err := cmd1.Run(); err != nil {
		panic(err)
	}
}
