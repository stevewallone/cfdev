package cmd

import (
	"encoding/binary"
	"fmt"
	"net"
)

type Command interface {
	Execute(*net.UnixConn) error
}

const BindType = uint8(6)

func UnmarshalCommand(conn *net.UnixConn) (Command, error) {
	var instr uint8
	binary.Read(conn, binary.LittleEndian, &instr)

	switch instr {
	case BindType:
		return UnmarshalBindCommand(conn)
	default:
		return nil, fmt.Errorf("Unimplemented instruction: %d", instr)
	}
}
