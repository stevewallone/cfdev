package host_test

import (
	"os/exec"

	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/host"
	"code.cloudfoundry.org/cfdev/hosts/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Host", func() {

	var (
		mockController *gomock.Controller
		mockPowershell *mocks.MockPowershell
		h			   *host.Host
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockPowershell = mocks.NewPowershell(mockController)
		h = &host.Host{
			Powershell: mockPowershell,
		}
	})

	Describe("check requirements", func() {
		Context("when not running in an admin shell", func() {
			It("returns an error", func() {
				mockPowershell.EXPECT(`IsInRole`).Return("FALSE", nil)

				err := h.CheckRequirements()
				Expect(err.Error()).To(ContainSubstring(`Running without admin privileges: You must run cf dev with an admin privileged powershell`))
				Expect(errors.SafeError(err)).To(Equal("Running without admin privileges"))
			})
		})

		Context("when running in an admin shell", func() {
			Context("Hyper-V is enabled on a Windows 10 machine", func() {
				It("succeeds", func() {
					mockPowershell.EXPECT(`IsInRole`).Return("TRUE", nil)
					mockPowershell.EXPECT(`(Get-WindowsOptionalFeature -FeatureName Microsoft-Hyper-V-All -Online).State`).Return("Enabled", nil)
					mockPowershell.EXPECT(gomock.Any).AnyTimes()

					Expect(h.CheckRequirements()).To(Succeed())
				})
			})

			Context("Hyper-V is enabled on a Windows Server 2016 Machine", func() {
				It("succeeds", func() {
					mockPowershell.EXPECT(`IsInRole`).Return("TRUE", nil)
					mockPowershell.EXPECT(`(Get-WindowsOptionalFeature -FeatureName Microsoft-Hyper-V-Management-PowerShell -Online).State`).Return("Enabled", nil)
					mockPowershell.EXPECT(gomock.Any).AnyTimes()

					Expect(h.CheckRequirements()).To(Succeed())
				})
			})

			Context("Hyper-V is disabled", func() {
				It("returns an error", func() {
					mockPowershell.EXPECT(`IsInRole`).Return("TRUE", nil)
					mockPowershell.EXPECT(`(Get-WindowsOptionalFeature -FeatureName Microsoft-Hyper-V-All -Online).State`).Return("Disabled", nil)
					mockPowershell.EXPECT(`(Get-WindowsOptionalFeature -FeatureName Microsoft-Hyper-V-Management-PowerShell -Online).State`).Return("Disabled", nil)
					mockPowershell.EXPECT(gomock.Any).AnyTimes()

					err := h.CheckRequirements()
					Expect(err.Error()).To(ContainSubstring(`Hyper-V disabled: You must first enable Hyper-V on your machine`))
					Expect(errors.SafeError(err)).To(Equal("Hyper-V disabled"))
				})
			})
		})
	})
})
