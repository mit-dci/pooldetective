package sentinel

import (
	"fmt"
	"sync"
	"time"

	"github.com/mit-dci/pooldetective/wire"

	"github.com/mit-dci/pooldetective/blockobserver/coinparam"
	"github.com/mit-dci/pooldetective/blockobserver/logging"
	"github.com/mit-dci/pooldetective/blockobserver/utils"
)

type Sentinel struct {
	observations        chan *wire.BlockObserverBlockObservedMsg
	coinParams          *coinparam.Params
	coinID              int
	connections         []*Connection
	connectionsLock     sync.Mutex
	existingConnections map[string]bool
	logPrefix           string
	LocationID          int
}

func NewSentinel(p *coinparam.Params, observations chan *wire.BlockObserverBlockObservedMsg, coinID int) *Sentinel {
	logPrefix := fmt.Sprintf("[Sentinel/%s]", p.Name)

	return &Sentinel{
		observations:        observations,
		existingConnections: map[string]bool{},
		connectionsLock:     sync.Mutex{},
		logPrefix:           logPrefix,
		coinParams:          p,
		coinID:              coinID,
	}
}

func (s *Sentinel) DiscoveredNode(ip string) {
	s.connectionsLock.Lock()
	s.AddConnection(ip)
	s.connectionsLock.Unlock()
}

func (s *Sentinel) ManageConnections() {
	for {
		if s.NumConnections() < 4 {
			s.AddOutgoingConnection()
		}
		time.Sleep(time.Second * 5)
	}
}

func (s *Sentinel) NumConnections() int {
	active := 0
	for _, c := range s.connections {
		if c.Running {
			active++
		}
	}
	return active
}

func (s *Sentinel) AddOutgoingConnection() {
	added := false
	for _, c := range s.connections {
		if !c.Running {
			err := c.Start()
			if err == nil {
				added = true
				break
			}
		}
	}

	if !added {
		logging.Warnf("Unable to add new connection. None could be started succesfully")
	}
}

func (s *Sentinel) AddConnection(ip string) {
	conString, _, err := utils.ResolveNode(ip, s.coinParams.DefaultPort)
	if err != nil {
		logging.Warnf("%s Unable to parse host: %s", s.logPrefix, err.Error())
		return
	}
	_, ok := s.existingConnections[conString]
	if !ok {
		logging.Debugf("%s New peer discovered: %s", s.logPrefix, ip)
		s.existingConnections[conString] = true
		conn := NewConnection(conString, s)
		s.connections = append(s.connections, conn)
	} else {
		logging.Debugf("%s Ignoring duplicate peer %s", s.logPrefix, ip)
	}
}

func (s *Sentinel) Start() {
	logging.Debugf("%s Starting sentinel", s.logPrefix)

	ips, err := s.GetNodesFromDNSSeeds()
	if err != nil {
		logging.Errorf("%s DNS Seed failed: %s", s.logPrefix, err.Error())
		return
	}

	logging.Debugf("%s Found %d potential nodes", s.logPrefix, len(ips))
	s.connections = make([]*Connection, 0)
	s.connectionsLock.Lock()
	for _, ip := range ips {
		s.AddConnection(ip)
	}
	s.connectionsLock.Unlock()

	s.ManageConnections()
}

func (s *Sentinel) Stop() {
	if s != nil {
		if s.connections != nil {
			for _, c := range s.connections {
				if c.Running {
					select {
					case c.stop <- true:
					default:
					}
				}
			}
		}
	}
}
