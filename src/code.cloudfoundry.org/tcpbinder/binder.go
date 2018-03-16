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

type Binder struct {
	socket string
	conn   *net.UnixListener
}

func New(socket string) (*Binder, error) {
	conn, err := net.ListenUnix("unix", &net.UnixAddr{
		Net:  "unix",
		Name: socket,
	})
	os.Chmod(socket, 0666)
	return &Binder{
		socket: socket,
		conn:   conn,
	}, err
}

func (b *Binder) Cleanup() {
	fmt.Println("Cleanup")
	os.Remove(b.socket)
	b.conn.Close()
}

func (b *Binder) Socket() string {
	return b.socket
}

func (b *Binder) Accept() (*net.UnixConn, error) {
	return b.conn.AcceptUnix()
}

func main() {
	binder, err := New("/var/tmp/com.docker.vmnetd.socket")
	if err != nil {
		panic("I can't listen!!")
	}
	defer binder.Cleanup()
	fmt.Println("Listening on socket at", binder.Socket())

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func(c chan os.Signal, binder *Binder) {
		sig := <-c
		log.Printf("Caught signal %s: shutting down.", sig)
		binder.Cleanup()
		os.Exit(0)
	}(sigc, binder)

	for {
		conn, err := binder.Accept()
		if err != nil {
			log.Fatal("Accept error: ", err)
		}
		go Run(conn)
	}
}

func Run(conn *net.UnixConn) {
	defer conn.Close()
	handshake(conn)
	addr, err := parseReq(conn)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("received request for %s:%d \n", addr.IP, addr.Port)

	file, err := bind(addr)
	if file != nil {
		defer file.Close()
	}

	msg, scmsg := response(file, err)
	fmt.Printf("sending msg: %+v scmsg %+v\n", msg, scmsg)
	if _, _, err := conn.WriteMsgUnix(msg, scmsg, nil); err != nil {
		fmt.Println("Error writing fd msg: ", err)
		return
	}
}

func bind(addr *net.TCPAddr) (*os.File, error) {
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}
	defer listener.Close()
	return listener.File()
}

func response(file *os.File, err error) ([]byte, []byte) {
	msg := make([]byte, 8, 8)
	var scmsg []byte
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok {
			if sysErr, ok := opErr.Err.(*os.SyscallError); ok {
				switch sysErr.Err {
				case syscall.EADDRINUSE:
					msg[0] = uint8(48)
				case syscall.EADDRNOTAVAIL:
					msg[0] = uint8(49)
				default:
					msg[0] = uint8(66)
				}
			}
		}
		fmt.Println("Failed to bind: ", err)
	}
	if file != nil {
		scmsg = syscall.UnixRights(int(file.Fd()))
	}
	return msg, scmsg
}

func handshake(conn net.Conn) {
	init := make([]byte, 49, 49)
	io.ReadFull(conn, init)
	fmt.Printf("Connection received from: %s\n", init[0:5])

	var version uint32 = 22
	conn.Write([]byte("CFD3V"))
	binary.Write(conn, binary.LittleEndian, version)
	conn.Write([]byte("0123456789012345678901234567890123456789"))
}

func parseReq(conn net.Conn) (*net.TCPAddr, error) {
	var instr uint8
	ip := make([]byte, 4, 4)
	var port uint16
	binary.Read(conn, binary.LittleEndian, &instr)
	binary.Read(conn, binary.LittleEndian, ip)
	binary.Read(conn, binary.LittleEndian, &port)

	if instr != 6 {
		return nil, fmt.Errorf("Unimplemented instruction: %d", instr)
	}

	return &net.TCPAddr{
		IP:   []byte{ip[3], ip[2], ip[1], ip[0]},
		Port: int(port),
	}, nil
}
