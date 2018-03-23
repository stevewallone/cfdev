package launchd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"path/filepath"
	"code.cloudfoundry.org/cfdevd/launchd"
	"os"
	"fmt"
	"github.com/onsi/gomega/gexec"
	"os/exec"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("launchd", func() {
	Describe("Load", func() {

		var plistDir string
		var binDir string
		var lnchd launchd.Launchd

		BeforeEach(func() {
			plistDir, _ = ioutil.TempDir("", "plist")
			binDir, _ = ioutil.TempDir("", "bin")
			lnchd = launchd.Launchd{
				PListDir: plistDir,
			}
			ioutil.WriteFile(filepath.Join(binDir, "some-executable"), []byte(`some-content`), 0777)
		})

		AfterEach(func() {
			Expect(os.RemoveAll(plistDir)).To(Succeed())
			Expect(os.RemoveAll(binDir)).To(Succeed())
		})

		It("Should load the daemon", func() {
			installationPath := filepath.Join(binDir, "org.some-org.some-daemon-executable")
			spec := launchd.DaemonSpec{
				Label:            "org.some-org.some-daemon-name",
				Program:          installationPath,
				ProgramArguments: []string{"some-installation-path", "some-arg"},
				RunAtLoad:        true,
			}

			Expect(lnchd.AddDaemon(spec, filepath.Join(binDir, "some-executable"))).To(Succeed())
			plistPath := filepath.Join(plistDir, "/org.some-org.some-daemon-name.plist")
			Expect(plistPath).Should(BeAnExistingFile())
			plistFile, err := os.Open(plistPath)
			Expect(err).NotTo(HaveOccurred())
			plistData, err := ioutil.ReadAll(plistFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(plistData)).To(Equal(fmt.Sprintf(
				`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>org.some-org.some-daemon-name</string>
  <key>Program</key>
  <string>%s</string>
  <key>ProgramArguments</key>
  <array>
    <string>some-installation-path</string>
    <string>some-arg</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
</dict>
</plist>
`, filepath.Join(binDir, "org.some-org.some-daemon-executable"))))
			plistFileInfo, err := plistFile.Stat()
			Expect(err).ToNot(HaveOccurred())
			var expectedPlistMode os.FileMode = 0644
			Expect(plistFileInfo.Mode()).Should(Equal(expectedPlistMode))

			Expect(installationPath).To(BeAnExistingFile())
			installedBinary, err := os.Open(installationPath)
			Expect(err).NotTo(HaveOccurred())
			binFileInfo, err := installedBinary.Stat()
			var expectedBinMode os.FileMode = 0700
			Expect(binFileInfo.Mode()).To(Equal(expectedBinMode))
			contents, err := ioutil.ReadAll(installedBinary)
			Expect(string(contents)).To(Equal("some-content"))

			session, err := gexec.Start(exec.Command("launchctl", "list"), GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			defer Expect(exec.Command("launchctl", "unload", plistPath).Run()).To(Succeed())
			Eventually(session).Should(gbytes.Say("org.some-org.some-daemon-name"))
		})
	})
})
