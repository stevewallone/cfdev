package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	ps "github.com/gorillalabs/go-powershell"
	"github.com/gorillalabs/go-powershell/backend"
	"github.com/natefinch/npipe"
)

func main() {
	// choose a backend
	back := &backend.Local{}

	// start a local powershell process
	shell, err := ps.New(back)
	if err != nil {
		panic(err)
	}
	defer shell.Exit()

	stdout, _, err := shell.Execute("New-VM -Name 'cfdev11' -SwitchName 'Default Switch' -Generation 2 -NoVHD")
	if err != nil {
		fmt.Println("New-VM:", stdout)
		panic(err)
	}

	stdout, _, err = shell.Execute("Set-VM -Name 'cfdev11' -AutomaticStartAction Nothing -AutomaticStopAction ShutDown -CheckpointType Disabled -MemoryStartupBytes 8192MB -StaticMemory -ProcessorCount 4")
	if err != nil {
		fmt.Println("Set-VM:", stdout)
		panic(err)
	}

	stdout, _, err = shell.Execute("(Get-VM -Name 'cfdev11').Id")
	if err != nil {
		panic(err)
	}
	id := strings.Split(stdout, "\r\n")[3]
	fmt.Println("VM GUID:", id)

	// cmd1 := exec.Command(`C:\Users\WX014\Desktop\vpnkit\vpnkit.exe`,
	// 	"--ethernet", "hyperv-connect://"+id+"/30D48B34-7D27-4B0B-AAAF-BBBED334DD59",
	// 	"--port", "//./pipe/cfdevVpnKitControl",
	// 	"--port", "hyperv-connect://"+id+"/C378280D-DA14-42C8-A24E-0DE92A1028E2",
	// 	"--dns", `C:\Users\WX014\AppData\Roaming\Docker\resolv.conf`,
	// 	"--dhcp", `C:\Users\WX014\AppData\Roaming\Docker\dhcp.json`,
	// 	"--diagnostics", `\\.\pipe\cfdevVpnKitDiagnostics`,
	// 	"--host-names", "hostname",
	// 	"--gateway-names", "gateway.cfdev.internal,cfdev.gateway.internal,cfdev.http.internal",
	// 	"--listen-backlog", "32",
	// 	"--lowest-ip", "169.254.82.3",
	// 	"--highest-ip", "169.254.82.255",
	// 	"--host-ip", "169.254.82.2",
	// 	"--gateway-ip", "169.254.82.1")
	// cmd1.Stdout = os.Stdout
	// cmd1.Stderr = os.Stderr
	// if err := cmd1.Start(); err != nil {
	// 	panic(err)
	// }

	stdout, _, err = shell.Execute(`Add-VMHardDiskDrive -VMName 'cfdev11' -Path 'C:\Users\WX014\Desktop\cfdepsvhd.vhd'`)
	if err != nil {
		fmt.Println(stdout)
		panic(err)
	}

	stdout, _, err = shell.Execute(`Remove-VMNetworkAdapter -VMName cfdev11 -Name  "Network Adapter"`)
	if err != nil {
		fmt.Println(stdout)
		panic(err)
	}

	// os.Remove(`C:\Users\WX014\Desktop\empty11.vhd`)

	// stdout, _, err = shell.Execute(`New-VHD -Path 'C:\Users\WX014\Desktop\empty11.vhd' -SizeBytes '200000000000' -Dynamic`)
	// if err != nil {
	// 	fmt.Println(stdout)
	// 	panic(err)
	// }

	stdout, _, err = shell.Execute(`Add-VMHardDiskDrive -VMName 'cfdev11' -Path 'C:\Users\WX014\Desktop\moreempty.vhd'`)
	if err != nil {
		fmt.Println(stdout)
		panic(err)
	}

	stdout, _, err = shell.Execute(`Add-VMDvdDrive -VMName 'cfdev11' -Path 'C:\Users\WX014\Desktop\cfdev-efi.iso'`)
	if err != nil {
		fmt.Println(stdout)
		panic(err)
	}

	stdout, _, err = shell.Execute("$cdrom = Get-VMDvdDrive -vmname 'cfdev11'; Set-VMFirmware -VMName 'cfdev11' -EnableSecureBoot Off -FirstBootDevice $cdrom")
	if err != nil {
		fmt.Println(stdout)
		panic(err)
	}

	stdout, _, err = shell.Execute(`Set-VMComPort -VMName 'cfdev11' -number 1 -Path \\.\pipe\cfdev11-com1`)
	if err != nil {
		fmt.Println(stdout)
		panic(err)
	}

	// go func() {

	// 	cmd1 := exec.Command(`C:\Users\WX014\Desktop\vpnkit\vpnkit.exe`,
	// 		"--ethernet", "hyperv-connect://"+id,
	// 		"--port", "//./pipe/cfdevVpnKitControl",
	// 		"--port", "hyperv-connect://"+id,
	// 		"--dns", `C:\Users\WX014\AppData\Roaming\Docker\resolv.conf`,
	// 		"--dhcp", `C:\Users\WX014\AppData\Roaming\Docker\dhcp.json`,
	// 		"--diagnostics", `\\.\pipe\cfdevVpnKitDiagnostics`,
	// 		"--host-names", "host.cfdev.internal,cfdev.host.internal,cfdev.localhost",
	// 		"--gateway-names", "gateway.cfdev.internal,cfdev.gateway.internal,cfdev.http.internal",
	// 		"--listen-backlog", "32",
	// 		"--log-destination", `C:\Users\WX014\Desktop\cfdev\vpnkit.log`,
	// 		"--debug")
	// 	cmd1.Stdout = os.Stdout
	// 	cmd1.Stderr = os.Stderr
	// 	if err := cmd1.Start(); err != nil {
	// 		panic(err)
	// 	}

	// 	cmd2 := exec.Command(
	// 		`C:\Program Files\Docker\Docker\Resources\com.docker.proxy.exe`,
	// 		"-VM="+id,
	// 	)
	// 	cmd2.Stdout = os.Stdout
	// 	cmd2.Stderr = os.Stderr
	// 	if err := cmd2.Start(); err != nil {
	// 		panic(err)
	// 	}
	// }()

	stdout, _, err = shell.Execute(`Start-VM -Name 'cfdev11'`)
	if err != nil {
		fmt.Println(stdout)
		panic(err)
	}

	// 192.168.25.34 netmask 255.255.255.0  gateway 192.168.25.1       A   B   C   D
	// 192.168.25.34/24
	// 192.168.25.255

	// 10.0.0.0/8
	// 192.168.0.0/16
	// fmt.Println("STARTED*******************")
	// time.Sleep(60 * time.Second)
	// fmt.Println("Adding the adapter***************************")
	// stdout, _, err = shell.Execute(`Add-VMNetworkAdapter -VMName cfdev11 -Name "Network Adapter" -SwitchName 'Default Switch'`)
	// if err != nil {
	// 	fmt.Println(stdout)
	// 	panic(err)
	// }
	// fmt.Println("Finished Adding the adapter***************************")

	stdout, _, err = shell.Execute(`Set-ExecutionPolicy Unrestricted`)
	if err != nil {
		fmt.Println(stdout)
		panic(err)
	}

	fmt.Println("VM GUID:", id)
	fmt.Println("VM GUID:", id)

	// stdout, _, err = shell.Execute(`. C:\Users\WX014\Desktop\set-vm.ps1`)
	// if err != nil {
	// 	fmt.Println(stdout)
	// 	panic(err)
	// }

	// fmt.Println("Send DHCP")

	// for {
	// 	stdout, _, err = shell.Execute(`Get-VMNetworkAdapter -VMName cfdev10 -Name "Network Adapter" | Set-VMNetworkConfiguration -IPAddress 192.168.0.31 -Subnet 255.255.255.0 -DNSServer 8.8.8.8 -DefaultGateway 192.168.0.1`)
	// 	if err != nil {
	// 		fmt.Println(stdout)
	// 	} else {
	// 		break
	// 	}
	// }

	// fmt.Println("DHCP sent")

	/*
		oldState, err := terminal.MakeRaw(0)
		if err != nil {
			panic(err)
		}
		defer terminal.Restore(0, oldState)
	*/

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	conn, err := npipe.Dial(`\\.\pipe\cfdev11-com1`)
	if err != nil {
		panic(err)
	}
	go io.Copy(conn, os.Stdin)
	go func() {
		for _ = range c {
			// sig is a ^C, handle it
			conn.Write([]byte("\033c"))
		}
	}()
	_, err = io.Copy(os.Stdout, conn)
}
