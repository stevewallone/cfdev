// +build darwin

package cmd

import (
	"encoding/binary"
	"io"
	"net"
	"os"
	"code.cloudfoundry.org/cfdev/daemon"
	"io/ioutil"
)

type Command interface {
	Execute(*net.UnixConn) error
}

const UninstallType = uint8(1)
const RemoveIPAliasType = uint8(2)
const AddIPAliasType = uint8(3)
const BindType = uint8(6)

func UnmarshalCommand(conn io.Reader) (Command, error) {
	var instr uint8
	binary.Read(conn, binary.LittleEndian, &instr)

	ioutil.WriteFile("/tmp/mylog1", []byte(""), 0777)

	switch instr {
	case BindType:
		ioutil.WriteFile("/tmp/mylogBindType", []byte(""), 0777)

		return UnmarshalBindCommand(conn)
	case UninstallType:
		ioutil.WriteFile("/tmp/mylogUninstall", []byte(""), 0777)

		return &UninstallCommand{
			DaemonRunner: daemon.New(""),
		}, nil
	case RemoveIPAliasType:
		ioutil.WriteFile("/tmp/mylogRemoveAlias", []byte(""), 0777)

		return &RemoveIPAliasCommand{}, nil
	case AddIPAliasType:
		ioutil.WriteFile("/tmp/mylog2", []byte(""), 0777)

		return &AddIPAliasCommand{}, nil
	default:
		return &UnimplementedCommand{
			Instruction: instr,
			Logger: os.Stdout,
		}, nil
	}
}
