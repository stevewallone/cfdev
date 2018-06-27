package start_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cfdev/cmd/start/mocks"
	"github.com/golang/mock/gomock"
	"code.cloudfoundry.org/cfdev/cmd/start"
	"code.cloudfoundry.org/cfdev/config"
	"io/ioutil"
	"path/filepath"
	"os"
	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/process"
	"code.cloudfoundry.org/cfdev/resource"
)

var _ = Describe("Start", func() {

	var (
		mockController      *gomock.Controller
		mockUI              *mocks.MockUI
		mockLaunchd         *mocks.MockLaunchd
		mockProcManager     *mocks.MockProcManager
		mockAnalyticsClient *mocks.MockAnalyticsClient
		mockToggle          *mocks.MockToggle
		mockHostNet         *mocks.MockHostNet
		mockCache           *mocks.MockCache
		mockCFDevD          *mocks.MockCFDevD
		mockVpnkit          *mocks.MockVpnkit
		mockLinuxkit        *mocks.MockLinuxkit
		mockGardenClient    *mocks.MockGardenClient

		startCmd      start.Start
		exitChan      chan struct{}
		localExitChan chan struct{}
		tmpDir        string
	)

	BeforeEach(func() {
		var err error
		mockController = gomock.NewController(GinkgoT())
		mockUI = mocks.NewMockUI(mockController)
		mockLaunchd = mocks.NewMockLaunchd(mockController)
		mockProcManager = mocks.NewMockProcManager(mockController)
		mockAnalyticsClient = mocks.NewMockAnalyticsClient(mockController)
		mockToggle = mocks.NewMockToggle(mockController)
		mockHostNet = mocks.NewMockHostNet(mockController)
		mockCache = mocks.NewMockCache(mockController)
		mockCFDevD = mocks.NewMockCFDevD(mockController)
		mockVpnkit = mocks.NewMockVpnkit(mockController)
		mockLinuxkit = mocks.NewMockLinuxkit(mockController)
		mockLinuxkit = mocks.NewMockGardenClient(mockController)

		tmpDir, err = ioutil.TempDir("", "start-test-home")
		Expect(err).NotTo(HaveOccurred())

		startCmd = start.Start{
			Config: config.Config{
				CFDevHome:      tmpDir,
				StateDir:       filepath.Join(tmpDir, "some-state-dir"),
				VpnkitStateDir: filepath.Join(tmpDir, "some-vpnkit-state-dir"),
				CacheDir:       filepath.Join(tmpDir, "some-cache-dir"),
				CFRouterIP:     "some-cf-router-ip",
				BoshDirectorIP: "some-bosh-director-ip",
				Dependencies: resource.Catalog{
					Items: []resource.Item{{Name: "some-item"}},
				},
			},
			Exit:            exitChan,
			LocalExit:       localExitChan,
			UI:              mockUI,
			Launchd:         mockLaunchd,
			ProcManager:     mockProcManager,
			Analytics:       mockAnalyticsClient,
			AnalyticsToggle: mockToggle,
			HostNet:         mockHostNet,
			Cache:           mockCache,
			CFDevD:          mockCFDevD,
			Vpnkit:          mockVpnkit,
			Linuxkit:        mockLinuxkit,
		}
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
		mockController.Finish()
	})

	Describe("Execute", func() {
		Context("when no args are provided", func() {
			It("starts the vm with default settings", func() {
				mockToggle.EXPECT().SetProp("type", "cf")
				mockAnalyticsClient.EXPECT().Event(cfanalytics.START_BEGIN)
				mockLaunchd.EXPECT().IsRunning(process.LinuxKitLabel).Return(false, nil)
				mockHostNet.EXPECT().AddLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip")
				mockUI.EXPECT().Say("Downloading Resources...")
				mockCache.EXPECT().Sync(resource.Catalog{
					Items: []resource.Item{{Name: "some-item"}},
				})
				mockUI.EXPECT().Say("Installing cfdevd network helper...")
				mockCFDevD.EXPECT().Install()
				mockUI.EXPECT().Say("Starting VPNKit...")
				mockVpnkit.EXPECT().Start()
				mockUI.EXPECT().Say("Starting the VM...")
				mockLinuxkit.EXPECT().Start(7, 6666)
				mockUI.EXPECT().Say("Waiting for Garden...")

				Expect(startCmd.Execute(start.Args{
					Cpus: 7,
					Mem: 6666,
				})).To(Succeed())
			})
		})
		Context("when linuxkit is already running", func() {
			It("says cf dev is already running", func() {
				mockToggle.EXPECT().SetProp("type", "cf")
				mockAnalyticsClient.EXPECT().Event(cfanalytics.START_BEGIN)
				mockLaunchd.EXPECT().IsRunning(process.LinuxKitLabel).Return(true, nil)
				mockUI.EXPECT().Say("CF Dev is already running...")
				mockAnalyticsClient.EXPECT().Event(cfanalytics.START_END, map[string]interface{}{"alreadyrunning": true})

				Expect(startCmd.Execute(start.Args{})).To(Succeed())
			})
		})
	})
})
