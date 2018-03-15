package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ln, err := net.ListenUnix("unix", &net.UnixAddr{
		Net:  "unix",
		Name: "/var/tmp/com.docker.vmnetd.socket",
	})
	if err != nil {
		panic("I can't listen!!")
	}
	defer os.Remove("/var/tmp/com.docker.vmnetd.socket")
	defer ln.Close()
	os.Chmod("/var/tmp/com.docker.vmnetd.socket", 0666)
	fmt.Println("Listening on socket at /var/tmp/com.docker.vmnetd.socket")

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func(ln net.Listener, c chan os.Signal) {
		sig := <-c
		log.Printf("Caught signal %s: shutting down.", sig)
		ln.Close()
		os.Remove("/var/tmp/com.docker.vmnetd.socket")
		os.Exit(0)
	}(ln, sigc)

	for {
		conn, err := ln.AcceptUnix()
		fmt.Println("connection received")
		if err != nil {
			log.Fatal("Accept error: ", err)
		}
		sayHi(conn)
		addr := parseReq(conn)
		fmt.Printf("received request for %s:%d \n", addr.IP, addr.Port)
		fmt.Println("binding")
		file := bind(addr)
		fmt.Printf("Opened fd: %d, name: %s \n", file.Fd(), file.Name())
		fmt.Println("sending fd")
		_, _, err = conn.WriteMsgUnix(make([]byte, 8, 8), syscall.UnixRights(int(file.Fd())), nil)
		if err != nil {
			log.Fatal("Error sending fd: ", err)
		}
		conn.Close()
	}
}

func bind(addr *net.TCPAddr) *os.File {
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	file, err := listener.File()
	if err != nil {
		panic(err)
	}
	return file
}

func sayHi(conn net.Conn) {
	init := make([]byte, 49, 49)
	io.ReadFull(conn, init)

	var version uint32 = 22
	conn.Write([]byte("CFD3V"))
	binary.Write(conn, binary.LittleEndian, version)
	conn.Write([]byte("0123456789012345678901234567890123456789"))
}

func parseReq(conn net.Conn) *net.TCPAddr {
	rawheader := make([]byte, 1, 1)
	rawIP := make([]byte, 4, 4)
	rawPort := make([]byte, 2, 2)

	n1, err1 := conn.Read(rawheader)
	n2, err2 := conn.Read(rawIP)
	n3, err3 := conn.Read(rawPort)
	fmt.Println("Read:", n1, err1, n2, err2, n3, err3)
	fmt.Printf("rawheader: %+v, rawip %+v, rawPort: %+v", rawheader, rawIP, rawPort)
	_, _ = binary.Uvarint(rawheader)
	ipInt := binary.LittleEndian.Uint32(rawIP)
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, ipInt)
	port := binary.LittleEndian.Uint16(rawPort)
	return &net.TCPAddr{
		IP:   ip,
		Port: int(port),
	}
}
