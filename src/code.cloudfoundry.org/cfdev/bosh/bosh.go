package bosh

import (
	"time"

	"code.cloudfoundry.org/cfdev/errors"
	"fmt"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	"io"
)

const (
	StateUploadingReleases = "uploading-releases"
	StateDeploying         = "deploying"

	vmProgressInterval     = 1 * time.Second
)

type Bosh struct {
	dir boshdir.Director
}

type Config struct {
	AdminUsername   string
	AdminPassword   string
	CACertificate   string
	DirectorAddress string

	GatewayHost       string
	GatewayPrivateKey string
	GatewayUsername   string
}

type UI interface {
	Writer() io.Writer
}

func New(config Config) (*Bosh, error) {
	cfg := boshdir.FactoryConfig{
		Host:         config.DirectorAddress,
		Port:         25555,
		CACert:       config.CACertificate,
		Client:       config.AdminUsername,
		ClientSecret: config.AdminPassword,
	}
	f := boshdir.NewFactory(&Logger{})
	dir, err := f.New(cfg, &TaskReporter{}, &FileReporter{})
	if err != nil {
		return nil, errors.SafeWrap(err, "failed to connect to bosh director")
	}
	return NewWithDirector(dir), nil
}

func NewWithDirector(dir boshdir.Director) *Bosh {
	return &Bosh{dir: dir}
}

type VMProgress struct {
	State    string
	Releases int
	Total    int
	Done     int
	Duration time.Duration
}

func (b *Bosh) ReportProgress(ui UI, name string, isErrand bool, doneChan chan bool) {
	var (
		start               = time.Now()
		clearTerminalPrefix = "\r\033[K  "
		dep                 boshdir.Deployment
	)

	if isErrand {
		for {
			select {
			case <-doneChan:
				return
			default:
				ui.Writer().Write([]byte(fmt.Sprintf(clearTerminalPrefix+"Running Errand (%s)", time.Now().Sub(start))))

				time.Sleep(vmProgressInterval)
			}
		}
	}

	for {
		var err error
		if dep, err = b.dir.FindDeployment(name); err == nil {
			break
		}
	}

	for {
		select {
		case <-doneChan:
			return
		default:
			p := b.getVMProgress(start, dep)

			switch p.State {
			case StateUploadingReleases:
				ui.Writer().Write([]byte(fmt.Sprintf(clearTerminalPrefix+"Uploaded Releases: %d (%s)", p.Releases, p.Duration.Round(time.Second))))
			case StateDeploying:
				ui.Writer().Write([]byte(fmt.Sprintf(clearTerminalPrefix+"Progress: %d of %d (%s)", p.Done, p.Total, p.Duration.Round(time.Second))))
			}

			time.Sleep(vmProgressInterval)
		}
	}
}

func (b *Bosh) getVMProgress(start time.Time, dep boshdir.Deployment) VMProgress {
	vmInfos, err := dep.VMInfos()
	if len(vmInfos) == 0 || err != nil {
		rels, _ := b.dir.Releases()
		return VMProgress{State: StateUploadingReleases, Releases: len(rels), Duration: time.Now().Sub(start)}
	}

	total := len(vmInfos)
	numDone := 0
	for _, v := range vmInfos {
		if v.ProcessState == "running" && len(v.Processes) > 0 {
			numDone++
		}
	}

	return VMProgress{State: StateDeploying, Total: total, Done: numDone, Duration: time.Now().Sub(start)}
}
