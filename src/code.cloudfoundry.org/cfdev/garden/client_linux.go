package garden

import (
	"path/filepath"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/garden/client/connection"
)

func newGardenConnection(Config config.Config) connection.Connection {
	return connection.New("unix", filepath.Join(Config.CFDevHome, "gdn.socket"))
}
