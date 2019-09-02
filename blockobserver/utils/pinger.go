package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/mit-dci/pooldetective/logging"
	"github.com/sparrc/go-ping"
)

func Pinger(host string, stop chan bool) {

	for {
		select {
		case <-stop:
			break
		default:
		}
		p, err := ping.NewPinger(host)
		if err != nil {
			logging.Errorf("Could not start pinger: %s", err.Error())
			return
		}
		p.SetPrivileged(true)
		p.Count = 10
		p.Run()
		stats := p.Statistics()
		target := fmt.Sprintf("%s/%s", os.Getenv("HUB_URL"), "observePeerPing")

		err = PostJson(target, map[string]interface{}{"observedFrom": os.Getenv("OBSERVER_NAME"), "timestamp": time.Now().Unix(), "host": host, "results": stats}, false, nil)
		if err != nil {
			logging.Warnf("Unable to post observation: %s\n", err.Error())
		}

		time.Sleep(time.Minute * 30)
	}
}
