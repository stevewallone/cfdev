package garden

import (
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/garden/client/connection"
)

func newGardenConnection(Config config.Config) connection.Connection {
	return connection.New("tcp", "localhost:8888")
}
