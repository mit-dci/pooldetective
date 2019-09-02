package wire

import (
	"fmt"

	"github.com/mit-dci/pooldetective/logging"
	zmq "github.com/pebbe/zmq4"
)

type PoolDetectiveConn struct {
	conn *zmq.Socket
}

func NewServer(port int) (*PoolDetectiveConn, error) {
	rep, err := zmq.NewSocket(zmq.REP)
	if err != nil {
		return nil, err
	}

	err = rep.Bind(fmt.Sprintf("tcp://*:%d", port))
	if err != nil {
		return nil, err
	}

	return &PoolDetectiveConn{
		conn: rep,
	}, nil
}

func NewPublisher(port int) (*PoolDetectiveConn, error) {
	pub, err := zmq.NewSocket(zmq.PUB)
	if err != nil {
		return nil, err
	}

	err = pub.Bind(fmt.Sprintf("tcp://*:%d", port))
	if err != nil {
		return nil, err
	}

	return &PoolDetectiveConn{
		conn: pub,
	}, nil
}

func NewClient(host string, port int) (*PoolDetectiveConn, error) {
	req, err := zmq.NewSocket(zmq.REQ)

	err = req.Connect(fmt.Sprintf("tcp://%s:%d", host, port))
	if err != nil {
		return nil, fmt.Errorf("could not dial: %v", err)
	}

	return &PoolDetectiveConn{
		conn: req,
	}, nil
}

func NewSubscriber(host string, port int) (*PoolDetectiveConn, error) {
	sub, err := zmq.NewSocket(zmq.SUB)
	err = sub.Connect(fmt.Sprintf("tcp://%s:%d", host, port))
	if err != nil {
		return nil, fmt.Errorf("could not dial: %v", err)
	}

	return &PoolDetectiveConn{
		conn: sub,
	}, nil
}

func (pdc *PoolDetectiveConn) Close() error {
	return pdc.conn.Close()
}

func (pdc *PoolDetectiveConn) Recv() (PoolDetectiveMsg, bool, error) {
	zm, err := pdc.conn.RecvMessageBytes(0)
	if err != nil {
		return nil, false, err
	}

	msg, err := msgFromFrames(zm)
	return msg, true, err
}

func (pdc *PoolDetectiveConn) RecvSub() (PoolDetectiveMsg, []byte, bool, error) {
	zm, err := pdc.conn.RecvMessageBytes(0)
	if err != nil {
		return nil, nil, false, err
	}

	msg, err := msgFromFrames(zm[1:])
	return msg, zm[0], true, err
}

func (pdc *PoolDetectiveConn) Send(msg PoolDetectiveMsg) error {
	frames, err := msgToFrames(msg)
	if err != nil {
		return err
	}
	logging.Debugf("Sending %d frames for %T", len(frames), msg)
	_, err = pdc.conn.SendMessage(frames)
	return err
}

func (pdc *PoolDetectiveConn) Publish(channel []byte, msg PoolDetectiveMsg) error {
	frames, err := msgToFrames(msg)
	if err != nil {
		return err
	}
	frames = append([][]byte{channel}, frames...)
	logging.Debugf("Sending %d frames for publish %T", len(frames), msg)
	_, err = pdc.conn.SendMessage(frames)
	return err
}

func (pdc *PoolDetectiveConn) Subscribe(channel []byte) error {
	return pdc.conn.SetSubscribe(string(channel))
}
