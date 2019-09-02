package wire

import (
	"fmt"

	binser "github.com/kelindar/binary"
)

func msgFromFrames(frames [][]byte) (PoolDetectiveMsg, error) {
	if len(frames) != 2 {
		return nil, fmt.Errorf("Expected two or more frames, got %d", len(frames))
	}

	var mt MessageType
	err := binser.Unmarshal(frames[0], &mt)
	if err != nil {
		return nil, err
	}
	msg := NewMessage(mt)
	err = binser.Unmarshal(frames[1], msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func msgToFrames(msg PoolDetectiveMsg) (frame [][]byte, err error) {
	frame = make([][]byte, 2)
	frame[0], err = binser.Marshal(GetMessageType(msg))
	if err != nil {
		return
	}
	frame[1], err = binser.Marshal(msg)
	if err != nil {
		return
	}

	return
}
