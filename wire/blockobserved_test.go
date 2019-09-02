package wire

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"

	"github.com/mit-dci/pooldetective/logging"
)

func setup() {
	logging.SetLogLevel(4)
}

func TestBlockObservedTransport(t *testing.T) {
	serverR, clientW := io.Pipe()
	clientR, serverW := io.Pipe()
	dc1 := DummyConn{}
	dc2 := DummyConn{}

	clientConn := NewConnWithReaderAndWriter(dc1, clientR, clientW)
	serverConn := NewConnWithReaderAndWriter(dc2, serverR, serverW)

	blockHash, _ := hex.DecodeString("8c17bd3b5b0c71dceec955383bb4a9fa1482fa38c2ccb54e9b7bec8dd26c37b5")
	ip := net.ParseIP("1.2.3.4")
	outMsg := &BlockObservedMsg{
		LocationID: 10,
		PeerIP:     []byte(ip),
		PeerPort:   8333,
	}
	copy(outMsg.BlockHash[:], blockHash)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		inMsg := <-clientConn.Incoming
		m, ok := inMsg.(*BlockObservedMsg)
		if !ok {
			msgType := fmt.Sprintf("%T", inMsg)
			t.Logf("Wrong message type %s", msgType)
			t.Fail()
			wg.Done()
			return
		}

		if !bytes.Equal(m.PeerIP, outMsg.PeerIP) {
			t.Log("IP Mismatch")
			t.Fail()
			wg.Done()
			return
		}

		if m.PeerPort != outMsg.PeerPort {
			t.Log("Port Mismatch")
			t.Fail()
			wg.Done()
			return
		}

		if m.LocationID != outMsg.LocationID {
			t.Log("Location ID Mismatch")
			t.Fail()
			wg.Done()
			return
		}

		if !bytes.Equal(m.BlockHash[:], outMsg.BlockHash[:]) {
			t.Log("BlockHash Mismatch")
			t.Fail()
			wg.Done()
			return
		}

		wg.Done()
	}()

	serverConn.Outgoing <- outMsg
	wg.Wait()
}
