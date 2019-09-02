package stratum

import (
	"bufio"
	"encoding/json"
	"net"
	"sync"
	"time"

	"github.com/mit-dci/pooldetective/logging"
)

type StratumMessage struct {
	MessageID    interface{}   `json:"id,omitempty"`
	RemoteMethod string        `json:"method,omitempty"`
	Parameters   interface{}   `json:"params,omitempty"`
	Result       interface{}   `json:"result,omitempty"`
	Error        []interface{} `json:"error"`
}

func (msg *StratumMessage) Id() int {
	messageIDFloat, ok := msg.MessageID.(float64)
	if ok {
		return int(messageIDFloat)
	}

	messageIDInt, ok := msg.MessageID.(int64)
	if ok {
		return int(messageIDInt)
	}

	return -1
}

func (msg *StratumMessage) String() string {
	j, err := json.Marshal(msg)
	if err != nil {
		return ""
	}
	return string(j)
}

type StratumConnection struct {
	conn         net.Conn
	connLock     sync.Mutex
	Incoming     chan StratumMessage
	Outgoing     chan StratumMessage
	Disconnected chan bool
	stopLogging  chan bool
	LogOutput    func([]CommEvent)
	logChannel   chan CommEvent
	CommsLog     []CommEvent
}

type CommEvent struct {
	When    int64          `json:"t"`
	In      bool           `json:"i"`
	Message StratumMessage `json:"msg"`
}

func (c *StratumConnection) LogLoop() {
	lastCommit := time.Now()
	for {
		stop := false
		select {
		case <-c.stopLogging:
			stop = true
		case log := <-c.logChannel:
			c.CommsLog = append(c.CommsLog, log)
			if len(c.CommsLog) > 100 {
				c.CommitLog()
				lastCommit = time.Now()
			}
		case <-time.After(time.Second * 10):
		}

		if stop {
			c.CommitLog()
			break
		}

		if time.Now().After(lastCommit.Add(time.Second * 10)) {
			c.CommitLog()
			lastCommit = time.Now()
		}
	}
}

func (c *StratumConnection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *StratumConnection) CommitLog() {
	if c.LogOutput != nil {
		copyLen := len(c.CommsLog)
		if copyLen > 0 {
			log := make([]CommEvent, copyLen)
			copy(log[:], c.CommsLog[:copyLen])
			if len(c.CommsLog) > copyLen {
				c.CommsLog = c.CommsLog[copyLen+1:]
			} else {
				c.CommsLog = make([]CommEvent, 0)
			}
			c.LogOutput(log)
		}
	} else {
		logging.Debugf("No logoutput function attached, dumping logs\n")
		c.CommsLog = make([]CommEvent, 0)
	}
}

func (c *StratumConnection) Log(in bool, msg StratumMessage) {
	if c.LogOutput == nil {
		return
	}
	c.logChannel <- CommEvent{
		When:    time.Now().UnixNano(),
		Message: msg,
		In:      in,
	}
}

func (c *StratumConnection) IncomingLoop() {
	defer func() {
		select {
		case c.Disconnected <- true:
		default:
		}
	}()

	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		msg := StratumMessage{}
		err := json.Unmarshal(scanner.Bytes(), &msg)
		if err != nil {
			logging.Warnf("Invalid JSON from Stratum Server: %s", err.Error())
			continue
		}

		c.Log(true, msg)
		c.Incoming <- msg
	}
}

func (c *StratumConnection) OutgoingLoop() {
	for msg := range c.Outgoing {
		c.Log(false, msg)
		s := []byte(msg.String() + "\n")
		w, err := c.conn.Write(s)
		if err != nil {
			logging.Warnf("Unable to send outgoing stratum message: %s\n", err.Error())
			break
		}
		if w != len(s) {
			logging.Warnf("Did not send all bytes of message. Expected %d, got %d\n", len(s), w)
			break
		}
	}
}

func (c *StratumConnection) Stop() {
	logging.Infof("Closing stratum connection")
	close(c.Incoming)
	close(c.Outgoing)
	c.stopLogging <- true
	c.conn.Close()
}

func NewStratumConnection(conn net.Conn) (*StratumConnection, error) {
	in := make(chan StratumMessage, 10)
	out := make(chan StratumMessage, 10)
	dis := make(chan bool, 1) // Need a buffer here. Client could be processing a message when disconnect happens
	sl := make(chan bool, 1)
	sc := StratumConnection{
		conn:         conn,
		connLock:     sync.Mutex{},
		logChannel:   make(chan CommEvent, 100),
		Incoming:     in,
		Outgoing:     out,
		Disconnected: dis,
		stopLogging:  sl,
		CommsLog:     make([]CommEvent, 0),
	}

	go sc.LogLoop()
	go sc.IncomingLoop()
	go sc.OutgoingLoop()
	return &sc, nil
}
