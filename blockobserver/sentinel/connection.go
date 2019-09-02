package sentinel

import (
	"fmt"
	"net"
	"time"

	"github.com/mit-dci/lit/btcutil/chaincfg/chainhash"
	"github.com/mit-dci/pooldetective/blockobserver/logging"
	"github.com/mit-dci/pooldetective/blockobserver/utils"
	"github.com/mit-dci/pooldetective/blockobserver/wire"
	pdwire "github.com/mit-dci/pooldetective/wire"
)

var (
	VERSION = uint32(70012)
)

type Connection struct {
	host               string
	ip                 []byte
	port               int
	s                  *Sentinel
	lastHeaderLocators []*chainhash.Hash
	headerQueue        []chainhash.Hash
	outMsgQueue        chan wire.Message // Messages going out to remote node
	logPrefix          string
	con                net.Conn
	remoteVersion      uint32
	remoteHeight       int32
	localVersion       uint32
	Running            bool
	tryAfter           time.Time
	stop               chan bool
}

func NewConnection(ip string, s *Sentinel) *Connection {
	return &Connection{stop: make(chan bool), tryAfter: time.Now(), host: ip, s: s, logPrefix: fmt.Sprintf("[%s/%s]", s.coinParams.Name, ip)}
}

func (c *Connection) Start() error {
	if !time.Now().After(c.tryAfter) {
		return fmt.Errorf("Not supposed to run this one yet")
	}
	c.Running = true
	conString, conMode, err := utils.ResolveNode(c.host, c.s.coinParams.DefaultPort)
	if err != nil {
		logging.Debugf("%s Unable to parse host: %s", c.logPrefix, err.Error())
		c.tryAfter = time.Date(2099, time.January, 1, 0, 0, 0, 0, time.UTC) // Never again
		c.Running = false
		return err
	}

	d := net.Dialer{Timeout: 2 * time.Second}
	c.con, err = d.Dial(conMode, conString)
	if err != nil {
		logging.Debugf("%s Unable to connect: %s", c.logPrefix, err.Error())
		c.tryAfter = time.Now().Add(time.Minute * 15)
		c.Running = false
		return err
	}

	c.ip, c.port, err = utils.IPAndPort(conString)
	if err != nil {
		logging.Debugf("%s Unable to parse IP and port: %s", c.logPrefix, err.Error())
		c.tryAfter = time.Now().Add(time.Minute * 15)
		c.Running = false
		return err
	}

	err = c.Handshake()
	if err != nil {
		c.tryAfter = time.Now().Add(time.Minute * 15)
		logging.Debugf("%s Unable to complete handshake: %s", c.logPrefix, err.Error())
		return err
	}

	c.outMsgQueue = make(chan wire.Message, 10)
	go c.incomingMessageHandler()
	go c.outgoingMessageHandler()

	/*pingerChan := make(chan bool)

	if addr, ok := c.con.RemoteAddr().(*net.TCPAddr); ok {
		go utils.Pinger(addr.IP.String(), pingerChan)
	}*/
	go func() {
		for {
			select {
			case <-c.stop:
				c.Running = false

			default:
			}
			time.Sleep(time.Second * 5)
		}
	}()
	return nil
}

func (c *Connection) incomingMessageHandler() {
	for {
		_, xm, _, err := wire.ReadMessageWithEncodingN(c.con, c.localVersion,
			wire.BitcoinNet(c.s.coinParams.NetMagicBytes), wire.LatestEncoding)
		if err != nil {
			logging.Debugf("%s ReadMessageWithEncodingN error: %s - Closing connection", c.logPrefix, err.Error())
			c.tryAfter = time.Now().Add(time.Minute * 5)
			c.con.Close() // close the connection to prevent spam messages from crashing lit.
			c.stop <- true
			return
		}

		switch m := xm.(type) {
		case *wire.MsgAddr:
			for _, adr := range m.AddrList {
				c.s.AddConnection(fmt.Sprintf("%s:%d", adr.IP.String(), adr.Port))
			}
		case *wire.MsgVersion:
			logging.Debugf("Got version message.  Agent %s, version %d, at height %d\n",
				m.UserAgent, m.ProtocolVersion, m.LastBlock)
			c.remoteVersion = uint32(m.ProtocolVersion) // weird cast! bug?
		case *wire.MsgPing:
			c.outMsgQueue <- wire.NewMsgPong(m.Nonce)
		case *wire.MsgXVersion:
			c.outMsgQueue <- wire.NewMsgXVerAck()
		case *wire.MsgInv:
			c.InvHandler(m)
		case *wire.MsgVerAck:
			//logging.Debugf("%s Got verack.  Whatever.", c.logPrefix)
		case *wire.MsgPong:
			logging.Debugf("%s Got a pong response", c.logPrefix)
		case *wire.MsgAlert:
			c.AlertHandler(m)
			// BSV stuff
		case *wire.MsgProtoConf:
			c.outMsgQueue <- wire.NewMsgProtoConf(m.MaxRecvPayloadLength)
		default:
			if m != nil {
				logging.Debugf("Got unknown message type %s\n", m.Command())
			} else {
				logging.Warnf("Got nil message")
			}
		}
		xm = nil
	}
}

// this one seems kindof pointless?  could get ridf of it and let
// functions call WriteMessageWithEncodingN themselves...
func (c *Connection) outgoingMessageHandler() {
	for {
		msg := <-c.outMsgQueue
		if msg == nil {
			logging.Warnf("ERROR: nil message to outgoingMessageHandler\n")
			continue
		}
		_, err := wire.WriteMessageWithEncodingN(c.con, msg, c.localVersion,
			wire.BitcoinNet(c.s.coinParams.NetMagicBytes), wire.LatestEncoding)

		if err != nil {
			logging.Debugf("Write message error: %s", err.Error())
			c.tryAfter = time.Now().Add(time.Minute * 5)
			c.con.Close() // close the connection to prevent spam messages from crashing lit.
			c.stop <- true
			return
		}
	}
}

func (c *Connection) Handshake() error {
	// assign version bits for local node
	c.localVersion = VERSION
	myMsgVer, err := wire.NewMsgVersionFromConn(c.con, 0, 0)
	if err != nil {
		return err
	}
	err = myMsgVer.AddUserAgent(c.s.coinParams.IdentifyAsClient, c.s.coinParams.IdentifyAsVersion)
	if err != nil {
		return err
	}
	// must set this to enable SPV stuff
	myMsgVer.AddService(wire.SFNodeBloom)
	// set this to enable segWit
	myMsgVer.AddService(wire.SFNodeWitness)
	// this actually sends
	_, err = wire.WriteMessageWithEncodingN(
		c.con, myMsgVer, c.localVersion,
		wire.BitcoinNet(c.s.coinParams.NetMagicBytes), wire.LatestEncoding)
	if err != nil {
		return err
	}

	_, m, _, err := wire.ReadMessageWithEncodingN(c.con, c.localVersion,
		wire.BitcoinNet(c.s.coinParams.NetMagicBytes), wire.LatestEncoding)
	if err != nil {
		logging.Debugf("%s %s", c.logPrefix, err.Error())
		return err
	}

	mv, ok := m.(*wire.MsgVersion)
	if !ok {
		return fmt.Errorf("Unexpected message. Expected MsgVersion, got %T", m)
		//logging.Debugf("%s Connected - %s", c.logPrefix, mv.UserAgent)
	}
	// set remote height
	c.remoteHeight = mv.LastBlock
	// set remote version
	c.remoteVersion = uint32(mv.ProtocolVersion)

	mva := wire.NewMsgVerAck()
	_, err = wire.WriteMessageWithEncodingN(
		c.con, mva, c.localVersion,
		wire.BitcoinNet(c.s.coinParams.NetMagicBytes), wire.LatestEncoding)
	if err != nil {
		return err
	}

	return nil
}

func (c *Connection) AlertHandler(m *wire.MsgAlert) {
	//logging.Debugf("Got alert: %s\n", m.Payload.Comment)
}

// InvHandler ...
func (c *Connection) InvHandler(m *wire.MsgInv) {
	for _, thing := range m.InvList {
		if thing.Type == wire.InvTypeBlock {
			logging.Debugf("%s Discovered block %s", c.logPrefix, thing.Hash.String())
			msg := &pdwire.BlockObserverBlockObservedMsg{
				PeerIP:     c.ip,
				PeerPort:   c.port,
				LocationID: c.s.LocationID,
				CoinID:     c.s.coinID,
				Observed:   time.Now().UnixNano(),
			}
			copy(msg.BlockHash[:], thing.Hash.CloneBytes())
			c.s.observations <- msg
		}
	}
}

// PongBack ...
func (c *Connection) PongBack(nonce uint64) {
	mpong := wire.NewMsgPong(nonce)
	c.outMsgQueue <- mpong
	return
}
