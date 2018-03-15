package main_test

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Binder test", func() {
	var session *gexec.Session

	BeforeEach(func() {
		bin, err := gexec.Build("code.cloudfoundry.org/tcpbinder")
		Expect(err).NotTo(HaveOccurred())
		session, err = gexec.Start(exec.Command("sudo", "--non-interactive", bin), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gbytes.Say("Listening on socket at /var/tmp/com.docker.vmnetd.socket"))
	})

	AfterEach(func() {
		gexec.Start(exec.Command("sudo", "--non-interactive", "kill", fmt.Sprintf("%d", session.Command.Process.Pid)), GinkgoWriter, GinkgoWriter)
		gexec.KillAndWait()
		gexec.CleanupBuildArtifacts()
	})

	It("binds ports", func() {
		conn, err := net.DialUnix("unix", nil, &net.UnixAddr{
			Net:  "unix",
			Name: "/var/tmp/com.docker.vmnetd.socket",
		})
		Expect(err).NotTo(HaveOccurred())

		var version uint32 = 22
		conn.Write([]byte("VMN3T"))
		binary.Write(conn, binary.LittleEndian, version)
		conn.Write([]byte("0123456789012345678901234567890123456789"))

		bytes := make([]byte, 49, 49)
		Expect(io.ReadFull(conn, bytes)).To(Equal(49))
		Expect(bytes[0:5]).To(Equal([]byte("CFD3V")))

		conn.Write([]byte{0x6})
		ip := []byte(net.ParseIP("10.245.0.2").To4())
		conn.Write(append([]byte{}, ip[3], ip[2], ip[1], ip[0]))
		var port uint16 = 1888
		binary.Write(conn, binary.LittleEndian, port)

		b := make([]byte, 8, 8)
		oob := make([]byte, 16, 16)
		_, _, _, _, err = conn.ReadMsgUnix(b, oob)
		Expect(err).NotTo(HaveOccurred())
		scms, err := syscall.ParseSocketControlMessage(oob)
		Expect(err).NotTo(HaveOccurred())
		fds, err := syscall.ParseUnixRights(&scms[0])
		Expect(err).NotTo(HaveOccurred())
		file := os.NewFile(uintptr(fds[0]), "tcp:10.245.0.2:1888")
		ln, err := net.FileListener(file)
		Expect(err).NotTo(HaveOccurred())
		msg := "Hello from test"
		go func() {
			defer GinkgoRecover()
			wconn, err := net.Dial("tcp", "10.245.0.2:1888")
			Expect(err).NotTo(HaveOccurred())
			wconn.Write([]byte(msg))
		}()
		conn2, err := ln.Accept()
		Expect(err).NotTo(HaveOccurred())
		received := make([]byte, len(msg), len(msg))
		_, err = conn2.Read(received)
		Expect(err).NotTo(HaveOccurred())
		Expect(received).To(Equal([]byte(msg)))
	})
})
