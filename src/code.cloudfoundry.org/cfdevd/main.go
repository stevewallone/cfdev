package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"code.cloudfoundry.org/cfdevd/cmd"
	"code.cloudfoundry.org/cfdevd/launchd"
)

const Sock = "/var/tmp/cfdevd.socket"

func listen() (*net.UnixListener, error) {
	listener, err := net.ListenUnix("unix", &net.UnixAddr{
		Net:  "unix",
		Name: Sock,
	})
	if err != nil {
		return nil, err
	}
	os.Chmod(Sock, 0666)
	fmt.Println("Listening on socket at", Sock)
	return listener, err
}

func handleRequest(conn *net.UnixConn) {
	if err := doHandshake(conn); err != nil {
		fmt.Println("Handshake Error: ", err)
		return
	}
	command, err := cmd.UnmarshalCommand(conn)
	if err != nil {
		fmt.Println("Command:", err)
		return
	}
	command.Execute(conn)
}

func registerSignalHandler(listener net.Listener) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func(c chan os.Signal, listener net.Listener) {
		sig := <-c
		log.Printf("Caught signal %s: shutting down.", sig)
		listener.Close()
		os.Remove(Sock)
		os.Exit(0)
	}(sigc, listener)
}

func install(programSrc string) {
	lctl := launchd.Launchd{
		PListDir: "/Library/LaunchDaemons",
	}
	program := "/Library/PrivilegedHelperTools/org.cloudfoundry.cfdevd"
	cfdevdSpec := launchd.DaemonSpec{
		Label:   "org.cloudfoundry.cfdevd",
		Program: program,
		ProgramArguments: []string{
			program,
		},
		RunAtLoad: true,
	}
	if err := lctl.AddDaemon(cfdevdSpec, programSrc); err != nil {
		fmt.Println("Failed to install cfdevd: ", err)
	}
}

func uninstall(prog string) {
	lctl := launchd.Launchd{
		PListDir: "/Library/LaunchDaemons",
	}
	cfdevdSpec := launchd.DaemonSpec{
		Label:   "org.cloudfoundry.cfdevd",
		Program: "/Library/PrivilegedHelperTools/org.cloudfoundry.cfdevd",
		ProgramArguments: []string{
			prog,
		},
		RunAtLoad: true,
	}
	if err := lctl.RemoveDaemon(cfdevdSpec); err != nil {
		fmt.Println("Failed to uninstall cfdevd: ", err)
	}
}

func run() {
	listener, err := listen()
	if err != nil {
		log.Fatal("failed to listen on socket %s", Sock)
	}
	defer listener.Close()
	registerSignalHandler(listener)
	for {
		conn, err := listener.AcceptUnix()
		if err != nil {
			continue
		}
		defer conn.Close()
		go handleRequest(conn)
	}
}

func main() {
	fmt.Println(os.Args)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			install(os.Args[0])
		case "uninstall":
			uninstall(os.Args[0])
		default:
			log.Fatal("unrecognized command ", os.Args[1])
		}
	} else {
		run()
	}
}
