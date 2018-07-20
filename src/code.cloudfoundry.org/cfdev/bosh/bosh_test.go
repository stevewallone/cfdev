package bosh_test

import (
	"code.cloudfoundry.org/cfdev/bosh"
	"code.cloudfoundry.org/cfdev/bosh/mocks"
	"fmt"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"time"
)

//go:generate mockgen -package mocks -destination mocks/director.go github.com/cloudfoundry/bosh-cli/director Director
//go:generate mockgen -package mocks -destination mocks/deployment.go github.com/cloudfoundry/bosh-cli/director Deployment

var _ = Describe("Bosh", func() {
	var (
		subject        *bosh.Bosh
		mockController *gomock.Controller
		mockDir        *mocks.MockDirector
		mockDep        *mocks.MockDeployment
		mockUI         *fakeUI
		doneChan       chan bool
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockDir = mocks.NewMockDirector(mockController)
		mockDep = mocks.NewMockDeployment(mockController)
		mockUI = newFakeUI()
		subject = bosh.NewWithDirector(mockDir)
		doneChan = make(chan bool, 1)
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Describe("ReportProgress", func() {
		Context("when the deployment is not an errand", func() {
			It("reports release progress and then running vm with processes progress", func() {
				vmInfos := []boshdir.VMInfo{}
				mockDir.EXPECT().FindDeployment("cf").Return(mockDep, nil)
				mockDir.EXPECT().Releases().AnyTimes().Return([]boshdir.Release{nil, nil}, nil)
				mockDep.EXPECT().VMInfos().AnyTimes().DoAndReturn(func() ([]boshdir.VMInfo, error) {
					return vmInfos, nil
				})

				go func() {
					defer GinkgoRecover()

					subject.ReportProgress(mockUI, "cf", false, doneChan)
				}()

				time.Sleep(1500 * time.Millisecond)
				vmInfos = []boshdir.VMInfo{
					boshdir.VMInfo{ProcessState: "queued", Processes: []boshdir.VMInfoProcess{}},
					boshdir.VMInfo{ProcessState: "running", Processes: []boshdir.VMInfoProcess{}},
					boshdir.VMInfo{ProcessState: "running", Processes: []boshdir.VMInfoProcess{{}, {}}},
				}
				time.Sleep(2000 * time.Millisecond)
				doneChan <- true

				Expect(len(mockUI.writer.receivedWrites)).To(Equal(4))
				Expect(mockUI.writer.receivedWrites[0]).To(ContainSubstring("Uploaded Releases: 2"))
				Expect(mockUI.writer.receivedWrites[1]).To(ContainSubstring("Uploaded Releases: 2"))
				Expect(mockUI.writer.receivedWrites[2]).To(ContainSubstring("Progress: 1 of 3"))
				Expect(mockUI.writer.receivedWrites[3]).To(ContainSubstring("Progress: 1 of 3"))
			})
		})

		Context("when the deployment is an errand", func() {
			It("reports errand progress", func() {
				go func() {
					defer GinkgoRecover()

					subject.ReportProgress(mockUI, "some-errand", true, doneChan)
				}()

				time.Sleep(1500 * time.Millisecond)
				doneChan <- true

				Expect(len(mockUI.writer.receivedWrites)).To(Equal(2))
				Expect(mockUI.writer.receivedWrites[0]).To(ContainSubstring("Running Errand"))
				Expect(mockUI.writer.receivedWrites[1]).To(ContainSubstring("Running Errand"))
			})
		})
	})
})

type fakeWriter struct {
	receivedWrites []string
}

func (f *fakeWriter) Write(p []byte) (n int, err error) {
	f.receivedWrites = append(f.receivedWrites, string(p))
	return 0, nil
}

type fakeUI struct {
	writer *fakeWriter
}

func newFakeUI() *fakeUI {
	return &fakeUI{
		writer: &fakeWriter{},
	}
}

func (f *fakeUI) Writer() io.Writer {
	return f.writer
}
