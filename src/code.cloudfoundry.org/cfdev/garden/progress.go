package garden

import (
	"code.cloudfoundry.org/cfdev/bosh"
)

func (g *Garden) Report(ui bosh.UI, name string, isErrand bool, doneChan chan bool) {
	config, _ := g.FetchBOSHConfig()

	b, _ := bosh.New(config)

	b.ReportProgress(ui, name, isErrand, doneChan)
}