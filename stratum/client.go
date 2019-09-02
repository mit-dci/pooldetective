package stratum

import "net"

func NewStratumClient(addr string) (*StratumConnection, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return NewStratumConnection(conn)
}
