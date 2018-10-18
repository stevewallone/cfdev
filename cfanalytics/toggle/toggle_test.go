package toggle_test

import (
	"code.cloudfoundry.org/cfdev/cfanalytics/toggle"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"path/filepath"
)

var _ = Describe("Toggle", func() {
	var (
		tmpDir, saveFile string
	)
	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "analytics")
		Expect(err).ToNot(HaveOccurred())
		saveFile = filepath.Join(tmpDir, "somefile.txt")
	})
	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	Describe("Analytics file exists", func() {
		Context("cf and custom are enabled", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(saveFile, []byte(`{"cfAnalyticsEnabled":true,"customAnalyticsEnabled":true}`), 0644)).To(Succeed())
			})

			It("returns enabled true and custom is true", func() {
				t := toggle.New(saveFile)

				Expect(t.Enabled()).To(BeTrue())
				Expect(t.IsCustom()).To(BeTrue())
			})
		})

		Context("cf and custom are disabled", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(saveFile, []byte(`{"cfAnalyticsEnabled":false,"customAnalyticsEnabled":false}`), 0644)).To(Succeed())
			})

			It("returns enabled true and custom is false", func() {
				t := toggle.New(saveFile)

				Expect(t.Enabled()).To(BeFalse())
				Expect(t.IsCustom()).To(BeFalse())
			})
		})

		Context("cf is enabled and custom is disabled", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(saveFile, []byte(`{"cfAnalyticsEnabled":true,"customAnalyticsEnabled":false}`), 0644)).To(Succeed())
			})

			It("returns enabled true and custom is false", func() {
				t := toggle.New(saveFile)

				Expect(t.Enabled()).To(BeTrue())
				Expect(t.IsCustom()).To(BeFalse())
			})
		})

		Context("cf is disabled and custom is enabled", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(saveFile, []byte(`{"cfAnalyticsEnabled":false,"customAnalyticsEnabled":true}`), 0644)).To(Succeed())
			})

			It("returns enabled true and custom is true", func() {
				t := toggle.New(saveFile)

				Expect(t.Enabled()).To(BeTrue())
				Expect(t.IsCustom()).To(BeTrue())
			})
		})

		Context("update cfanalytics from false to true and save", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(saveFile, []byte(`{"cfAnalyticsEnabled":false,"customAnalyticsEnabled":false}`), 0644)).To(Succeed())
			})

			It("updates somefile.txt", func() {
				t := toggle.New(saveFile)
				Expect(t.SetCustomAnalyticsEnabled(true)).ToNot(HaveOccurred())
				Expect(t.IsCustom()).To(BeTrue())

				txt, err := ioutil.ReadFile(saveFile)
				Expect(err).ToNot(HaveOccurred())

				fmt.Printf("TXT: %v", string(txt))

				Expect(string(txt)).To(Equal(`{"cfAnalyticsEnabled":true,"customAnalyticsEnabled":true}`))
			})
		})

		Describe("Analytics file does NOT exist", func() {
			Context("cf is disabled and custom is enabled", func() {
				It("returns enabled true and custom is true and defined is true", func() {
					t := toggle.New("/non-existing-path")

					Expect(t.Defined()).To(BeFalse())
				})
			})
		})
	})
})
