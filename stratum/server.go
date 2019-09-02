package stratum

import (
	"fmt"
	"net"
)

type StratumListener struct {
	listen net.Listener
}

func NewStratumListener(port int) (*StratumListener, error) {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	return &StratumListener{
		listen: listen,
	}, nil
}

func (sl *StratumListener) Accept() (*StratumConnection, error) {
	conn, err := sl.listen.Accept()
	if err != nil {
		return nil, err
	}

	return NewStratumConnection(conn)
}
