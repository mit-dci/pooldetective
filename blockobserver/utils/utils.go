package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mit-dci/lit/btcutil/chaincfg/chainhash"
)

var jsonClient = &http.Client{Timeout: 60 * time.Second}

func ResolveNode(remoteNode, defaultPort string) (string, string, error) {
	colonCount := strings.Count(remoteNode, ":")
	var conMode string
	if colonCount <= 1 {
		if colonCount == 0 {
			remoteNode = remoteNode + ":" + defaultPort
		}
		return remoteNode, "tcp4", nil
	} else if colonCount >= 5 {
		// ipv6 without remote port
		// assume users don't give ports with ipv6 nodes
		if !strings.Contains(remoteNode, "[") && !strings.Contains(remoteNode, "]") {
			remoteNode = "[" + remoteNode + "]" + ":" + defaultPort
		}
		conMode = "tcp6"
		return remoteNode, conMode, nil
	} else {
		return "", "", fmt.Errorf("Invalid ip")
	}
}

func IPAndPort(host string) ([]byte, int, error) {
	idx := strings.LastIndex(host, ":")
	ip := net.ParseIP(host[:idx])
	if ip == nil {
		return nil, 0, fmt.Errorf("Could not parse IP")
	}
	port, err := strconv.ParseInt(host[idx+1:], 10, 32)
	if err != nil {
		return nil, 0, err
	}
	return []byte(ip), int(port), nil
}

func ConcatHashes(hs []*chainhash.Hash) string {
	s := ""
	for _, h := range hs {
		s += h.String()
	}
	return s
}

func GetJson(url string, target interface{}) error {
	r, err := jsonClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func PostJson(url string, payload interface{}, decodeResponse bool, target interface{}) error {
	var b bytes.Buffer
	json.NewEncoder(&b).Encode(payload)
	r, err := jsonClient.Post(url, "application/json", bytes.NewBuffer(b.Bytes()))
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if decodeResponse {
		return json.NewDecoder(r.Body).Decode(target)
	}
	if r.StatusCode >= 200 && r.StatusCode <= 299 {
		return nil
	}
	return fmt.Errorf("Unexpected response code %d", r.StatusCode)
}
