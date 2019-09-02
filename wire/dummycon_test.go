package wire

import (
	"net"
	"time"
)

var _ net.Conn = DummyConn{}

type DummyConn struct {
}

func (dc DummyConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (dc DummyConn) Write(b []byte) (n int, err error) {
	return 0, nil
}

func (dc DummyConn) Close() error {
	return nil
}
func (dc DummyConn) LocalAddr() net.Addr {
	return &net.IPAddr{net.IPv4(0xff, 0xff, 0xff, 0xff), ""}
}
func (dc DummyConn) RemoteAddr() net.Addr {
	return &net.IPAddr{net.IPv4(0xff, 0xff, 0xff, 0xff), ""}
}
func (dc DummyConn) SetDeadline(t time.Time) error {
	return nil
}
func (dc DummyConn) SetReadDeadline(t time.Time) error {
	return nil
}
func (dc DummyConn) SetWriteDeadline(t time.Time) error {
	return nil
}
