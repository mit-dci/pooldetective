package sentinel

import (
	"fmt"
	"net"

	"github.com/mit-dci/pooldetective/blockobserver/logging"
)

func (s *Sentinel) GetNodesFromDNSSeeds() ([]string, error) {
	var listOfNodes []string // slice of IP addrs returned from the DNS seed
	logging.Debugf("%s Resolving DNS Seeds", s.logPrefix)

	for _, seed := range s.coinParams.DNSSeeds {
		temp, err := net.LookupHost(seed)
		// need this temp in order to capture the error from net.LookupHost
		// also need this to report the number of IPs we get from a seed
		if err != nil {
			logging.Debugf("%s DNS Seed %s failed: %s", s.logPrefix, seed, err.Error())
			continue
		}
		listOfNodes = append(listOfNodes, temp...)
		logging.Debugf("%s Got %d IPs from %s", s.logPrefix, len(temp), seed)
	}
	if len(listOfNodes) == 0 {
		return nil, fmt.Errorf("No results from DNS Seeds")
	}

	return listOfNodes, nil
}
