package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"

	"github.com/pkg/profile"

	"github.com/mit-dci/pooldetective/wire"

	"github.com/mit-dci/pooldetective/blockobserver/coinparam"
	"github.com/mit-dci/pooldetective/blockobserver/logging"
	"github.com/mit-dci/pooldetective/blockobserver/sentinel"
	"github.com/mit-dci/pooldetective/util"
)

func main() {
	logging.SetLogLevel(util.GetLoglevelFromEnv())

	defer profile.Start().Stop()

	locationID, err := strconv.Atoi(os.Getenv("LOCATIONID"))
	if err != nil {
		log.Fatalf("Could not parse LOCATIONID from environment: %v", err)
	}

	observations := make(chan *wire.BlockObserverBlockObservedMsg, 100)

	go func() {
		http.ListenAndServe("localhost:8080", nil)
	}()

	for {
		req, err := wire.NewClient(os.Getenv("HUBHOST"), wire.PortBlockObserver)
		if err != nil {
			logging.Errorf("Could not connect to hub: %v", err)
			time.Sleep(time.Second * 5)
			continue
		}
		defer req.Close()

		err = req.Send(&wire.BlockObserverGetCoinsRequestMsg{})
		if err != nil {
			logging.Errorf("Could not send GetCoins request to hub: %v", err)
			time.Sleep(time.Second * 5)
			continue
		}

		msg, ok, err := req.Recv()
		if err != nil || !ok {
			logging.Errorf("Could not receive GetCoins response from hub: %v", err)
			time.Sleep(time.Second * 5)
			continue
		}

		coins, ok := msg.(*wire.BlockObserverGetCoinsResponseMsg)
		if !ok {
			logging.Errorf("GetCoins response from hub is wrong type: %T", msg)
			time.Sleep(time.Second * 5)
			continue
		}
		logging.Debugf("Starting %d sentinels", len(coins.Coins))
		sentinels := make([]*sentinel.Sentinel, len(coins.Coins))
		i := -1
		// Start one sentinel per coin
		for _, p := range coinparam.RegisteredNets {
			for _, c := range coins.Coins {
				if c.Ticker == p.Ticker {
					logging.Debugf("Starting sentinel for %s", c.Ticker)
					i++
					sentinels[i] = sentinel.NewSentinel(p, observations, c.CoinID)
					go sentinels[i].Start()
				} else {
					logging.Debugf("Ticker mismatch: %s vs %s", c.Ticker, p.Ticker)
				}
			}
		}

		for msg := range observations {
			msg.LocationID = locationID
			err := req.Send(msg)
			if err != nil {
				logging.Fatal(err)
				break
			}
			msg, _, err := req.Recv()
			if err != nil {
				logging.Error(err)
				break
			}
			switch t := msg.(type) {
			case *wire.ErrorMsg:
				logging.Error(t.Error)

			}
		}

		for i, s := range sentinels {
			s.Stop()
			sentinels[i] = nil
		}
		sentinels = nil

		time.Sleep(time.Second * 1)
	}
}
