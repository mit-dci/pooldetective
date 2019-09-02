package main

import (
	"fmt"
	"time"

	"github.com/mit-dci/pooldetective/logging"
	"github.com/mit-dci/pooldetective/util"
	"github.com/mit-dci/pooldetective/wire"
)

func main() {
	logging.SetLogLevel(util.GetLoglevelFromEnv())
	logging.Debugf("Pubsub Host...")

	pub, err := wire.NewPublisher(wire.PortPubSubSubscribers)
	if err != nil {
		logging.Fatal(err)
	}
	rep, err := wire.NewServer(wire.PortPubSubPublishers)
	if err != nil {
		logging.Fatal(err)
	}
	defer rep.Close()
	for {
		msg, channel, ok, err := rep.RecvSub()
		if err != nil && !ok {
			logging.Error(err)
			time.Sleep(time.Millisecond * 500)
			continue
		}

		logging.Debugf("Received request to publish %T to channel %s\n", msg, string(channel))

		var response wire.PoolDetectiveMsg
		if err != nil {
			response = &wire.ErrorMsg{YourID: 0, Error: fmt.Sprintf("Unable to parse request: %s", err.Error())}
		} else {
			err = pub.Publish(channel, msg)
			if err != nil {
				response = &wire.ErrorMsg{YourID: 0, Error: fmt.Sprintf("Unable to publish request: %s", err.Error())}
			} else {
				response = &wire.AckMsg{YourID: wire.GetMessageID(msg)}
			}
		}

		errMsg, ok := response.(*wire.ErrorMsg)
		if ok {
			logging.Error(errMsg.Error)
		}

		rep.Send(response)
	}

}
