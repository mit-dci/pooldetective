package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mit-dci/pooldetective/logging"
	"github.com/mit-dci/pooldetective/stratum"
	"github.com/mit-dci/pooldetective/util"
	"github.com/mit-dci/pooldetective/wire"
	"gopkg.in/Graylog2/go-gelf.v1/gelf"
)

var subscriptionID string
var nextSubmitIndex uint64 = 4
var difficulty float64
var stratumConn *stratum.StratumConnection
var stratumOutChan = make(chan stratum.StratumMessage, 100)
var hubChan = make(chan wire.PoolDetectiveMsg, 100)
var poolObserverID int
var connected bool
var authorized bool
var stratumHost string
var stratumLogin string
var stratumPassword string
var stratumProtocol int

var shareMap = map[uint64]int64{}
var shareMapLock = sync.Mutex{}

type publish struct {
	Channel []byte
	Msg     wire.PoolDetectiveMsg
}

var pubChan = make(chan publish, 100)
var gelfWriter *gelf.Writer

func main() {
	logging.SetLogLevel(util.GetLoglevelFromEnv())
	var err error

	poolObserverID, err = strconv.Atoi(os.Getenv("POOLOBSERVERID"))
	if err != nil {
		logging.Fatalf("Could not parse POOLOBSERVERID from environment: %v", err)
	}

	gelfWriter, err = gelf.NewWriter("graylog.blocksource.nl:12201")
	if err != nil {
		log.Fatalf("gelf.NewWriter: %s", err)
	}

	go func() {
		for {
			pub, err := wire.NewClient(os.Getenv("HUBHOST"), wire.PortPubSubPublishers)
			if err != nil {
				logging.Warnf("[PubSubPublisher] Could not connect to pubsub host on hub: %v", err)
				time.Sleep(time.Second * 5)
				continue
			}
			for p := range pubChan {
				err := pub.Publish(p.Channel, p.Msg)
				if err != nil {
					logging.Errorf("[PubSubPublisher] Error publishing message: %v", err)
					pubChan <- p // Requeue message
					break
				}

				// Have to call recv (will be ack/error)
				msg, _, err := pub.Recv()
				if err != nil {
					logging.Errorf("[PubSubPublisher] Could not receive: %v", err)
					break
				}

				switch t := msg.(type) {
				case *wire.ErrorMsg:
					logging.Errorf("[PubSubPublisher] Received an ErrorMsg: %v", t.Error)
				default:
				}

			}
			pub.Close()
			time.Sleep(time.Second)
		}
	}()

	go func() {
		for {
			logging.Debugf("[PubSubSubscriber] (Re)connecting to hub")
			sub, err := wire.NewSubscriber(os.Getenv("HUBHOST"), wire.PortPubSubSubscribers)
			if err != nil {
				logging.Warnf("[PubSubSubscriber] Could not connect to pubsub host on hub: %v", err)
				time.Sleep(time.Second * 5)
				continue
			}

			logging.Debugf("[PubSubSubscriber] Subscribing to topics")
			err = sub.Subscribe([]byte(fmt.Sprintf("sc-%03d", poolObserverID)))
			if err != nil {
				logging.Warnf("[PubSubSubscriber] Could not subscribe to topic on hub: %v", err)
				time.Sleep(time.Second * 5)
				continue
			}
			for {
				msg, _, _, err := sub.RecvSub()
				if err != nil {
					logging.Errorf("[PubSubSubscriber] Error receiving message: %v", err)
					break
				}

				switch t := msg.(type) {
				case *wire.StratumClientSubmitShareMsg:
					submitShare(t)
				}
			}
			sub.Close()
			time.Sleep(time.Second)
		}
	}()

	hubChan <- &wire.StratumClientLoginDetailsRequestMsg{PoolObserverID: poolObserverID}
	go func() {
		for {
			req, err := wire.NewClient(os.Getenv("HUBHOST"), wire.PortStratumClient)
			if err != nil {
				logging.Warnf("[StratumClientHost] Could not connect to hub: %v", err)
				time.Sleep(time.Second * 5)
				continue
			}
			for {
				msg := <-hubChan
				err := req.Send(msg)
				if err != nil {
					logging.Errorf("[StratumClientHost] Error sending message: %v", err)
					hubChan <- msg // Requeue message!
					break
				}

				msg, _, err = req.Recv()
				if err != nil {
					logging.Errorf("[StratumClientHost] Error receiving message: %v", err)
					break
				}

				switch t := msg.(type) {
				case *wire.StratumClientLoginDetailsResponseMsg:
					stratumHost = fmt.Sprintf("%s:%d", t.Host, t.Port)
					stratumLogin = t.Login
					stratumPassword = t.Password
					stratumProtocol = t.Protocol
					logging.Debugf("Received new stratum login details: [%s] [%s] [%s] [%d]", stratumHost, stratumLogin, stratumPassword, stratumProtocol)
				}
			}
			req.Close()
			time.Sleep(time.Second)
		}
	}()

	reconnect()
}

func messageToWork(msg *wire.StratumClientSubmitShareMsg) []interface{} {
	switch stratumProtocol {
	case 0:
		return []interface{}{
			stratumLogin,
			string(msg.PoolJobID),
			fmt.Sprintf("%x", msg.ExtraNonce2),
			fmt.Sprintf("%x", msg.Time),
			fmt.Sprintf("%x", msg.Nonce),
		}
	case 1, 2: // ZCash, BTG
		var timestampBytes bytes.Buffer
		binary.Write(&timestampBytes, binary.LittleEndian, int32(msg.Time))
		return []interface{}{
			stratumLogin,
			string(msg.PoolJobID),
			fmt.Sprintf("%x", timestampBytes.Bytes()),
			fmt.Sprintf("%x", msg.ExtraNonce2),
			fmt.Sprintf("%x", msg.AdditionalSolutionData[0]),
		}
	}

	return []interface{}{}
}

func submitShare(msg *wire.StratumClientSubmitShareMsg) {
	workID := atomic.AddUint64(&nextSubmitIndex, 1)

	shareMapLock.Lock()
	shareMap[workID] = msg.ShareID
	shareMapLock.Unlock()

	logging.Debugf("Submitting share %d", msg.ShareID)

	stratumOutChan <- stratum.StratumMessage{
		MessageID:    workID,
		RemoteMethod: "mining.submit",
		Parameters:   messageToWork(msg),
	}
	hubChan <- &wire.StratumClientShareEventMsg{
		Observed: time.Now().UnixNano(),
		Event:    wire.ShareEventSubmitted,
		ShareID:  msg.ShareID,
	}
}

func reconnect() {
	for {
		if stratumHost == "" {
			time.Sleep(time.Second * 1)
			continue
		}

		c, err := stratum.NewStratumClient(stratumHost)
		if err != nil {
			logging.Warnf("Unable to connect to stratum server: %s, retrying in 5 seconds", err.Error())
			time.Sleep(time.Second * 5)
			continue
		}

		c.LogOutput = func(logs []stratum.CommEvent) {
			logging.Debugf("Writing %d logs to graylog...", len(logs))
			for _, l := range logs {

				dir := "> "
				if l.In {
					dir = "< "
				}
				logging.Debugf("%s %s", dir, l.Message)

				msg, _ := json.Marshal(l.Message)
				err := gelfWriter.WriteMessage(&gelf.Message{
					Version: "1.1",
					Host:    fmt.Sprintf("stratum-client-%d", poolObserverID),
					Short:   string(msg),
					Extra: map[string]interface{}{
						"in": l.In,
					},
					TimeUnix: float64(l.When) / float64(1000000000),
				})
				if err != nil {
					logging.Errorf("Could not write logs to graylog: %v", err)
					break
				}
			}
		}

		logging.Infof("Connected to stratum server %s", stratumHost)
		connected = true
		hubChan <- &wire.StratumClientConnectionEventMsg{
			Observed:       time.Now().UnixNano(),
			Event:          wire.ConnectionEventConnected,
			PoolObserverID: poolObserverID,
		}

		c.Outgoing <- stratum.StratumMessage{
			MessageID:    1,
			RemoteMethod: "mining.subscribe",
			Parameters:   []string{"Miner/1.0"},
		}

		c.Outgoing <- stratum.StratumMessage{
			MessageID:    2,
			RemoteMethod: "mining.authorize",
			Parameters: []string{
				stratumLogin,
				stratumPassword,
			},
		}

		suggestDiff, err := strconv.ParseFloat(os.Getenv("STRATUM_SUGGESTDIFF"), 64)
		if err == nil && suggestDiff > 0 {
			c.Outgoing <- stratum.StratumMessage{
				MessageID:    3,
				RemoteMethod: "mining.suggest_difficulty",
				Parameters:   []float64{suggestDiff},
			}
		}

		stratumConn = c
		processConnection(c)
	}
}

func processConnection(c *stratum.StratumConnection) {
	for {
		close := false
		select {
		case <-c.Disconnected:
			hubChan <- &wire.StratumClientConnectionEventMsg{
				Observed:       time.Now().UnixNano(),
				Event:          wire.ConnectionEventDisconnected,
				PoolObserverID: poolObserverID,
			}
			go c.Stop()
			close = true
		case msg := <-c.Incoming:
			processStratumMessage(msg)
		case msg := <-stratumOutChan:
			if !authorized {
				log.Printf("Delaying message %s - not authorized. Will retry after 1s", msg.RemoteMethod)
				go func(m stratum.StratumMessage) {
					time.Sleep(time.Second)
					stratumOutChan <- m
				}(msg)
			} else {
				log.Printf("Sending message %s", msg.RemoteMethod)
				c.Outgoing <- msg
			}
		}
		if close {
			break
		}
	}
	connected = false
	authorized = false
	logging.Infof("Disconnected from stratum server")

}

func processStratumMessage(msg stratum.StratumMessage) {
	err := msg.Error
	if err != nil {
		logging.Warnf("Error response received: %v\n", err)
	}

	switch msg.Id() {
	case 1:
		if err == nil {
			processSubscriptionResponse(msg)
		}
	case 2:
		if err == nil {
			resultBool, ok := msg.Result.(bool)
			if ok && resultBool {
				hubChan <- &wire.StratumClientConnectionEventMsg{
					Observed:       time.Now().UnixNano(),
					Event:          wire.ConnectionEventAuthenticationSucceeded,
					PoolObserverID: poolObserverID,
				}
				logging.Infof("Succesfully authorized\n")
				authorized = true
			} else {
				hubChan <- &wire.StratumClientConnectionEventMsg{
					Observed:       time.Now().UnixNano(),
					Event:          wire.ConnectionEventAuthenticationFailed,
					PoolObserverID: poolObserverID,
				}
				logging.Warnf("Authorization failed\n")
			}
		}
	case 3:
		if err == nil {
			resultBool, ok := msg.Result.(bool)
			if ok && resultBool {
				logging.Debugf("Suggested difficulty accepted")
			}
		}
	default:
		ok := processRemoteInstruction(msg)
		if !ok && msg.Id() >= 4 {
			msgID := uint64(msg.Id())
			shareMapLock.Lock()
			shareID, shareIDOK := shareMap[msgID]
			shareMapLock.Unlock()
			// Response to a submitted share
			result, ok := msg.Result.(bool)
			if result && ok {
				if shareIDOK {
					hubChan <- &wire.StratumClientShareEventMsg{
						Observed: time.Now().UnixNano(),
						Event:    wire.ShareEventAccepted,
						ShareID:  shareID,
					}
				}
				logging.Info("Share accepted\n")
			} else {
				if shareIDOK {
					hubChan <- &wire.StratumClientShareEventMsg{
						Observed: time.Now().UnixNano(),
						Event:    wire.ShareEventDeclined,
						ShareID:  shareID,
						Details:  fmt.Sprintf("%v", msg.Error),
					}
				}
				if !result {
					logging.Info("Share declined\n")
				} else {
					logging.Info("Incorrect response to mining.submit\n")
				}
			}
		}
	}
}

func processSubscriptionResponse(msg stratum.StratumMessage) {
	arr, ok := msg.Result.([]interface{})
	if !ok {
		logging.Warnf("Result of subscription response is not an []interface{}\n")
		return
	}

	if stratumProtocol == 0 && len(arr) > 2 {
		logging.Infof("Setting extranonce1 [%s], extranonce2_size: [%f] (from subscription response)\n", arr[1].(string), arr[2].(float64))
		extraNonce1, _ := hex.DecodeString(arr[1].(string))
		extraNonce2Size := int8(arr[2].(float64))
		msg := &wire.StratumClientExtraNonceMsg{
			ExtraNonce1:     extraNonce1,
			ExtraNonce2Size: extraNonce2Size,
			PoolObserverID:  poolObserverID,
		}
		hubChan <- msg
		pubChan <- publish{Msg: msg, Channel: []byte("extranonce")}
	} else if (stratumProtocol == 1 || stratumProtocol == 2) && len(arr) >= 2 {
		extraNonce1, _ := hex.DecodeString(arr[1].(string))
		extraNonce2Size := int8(32 - len(extraNonce1))
		logging.Infof("Setting extranonce1 [%x], extranonce2_size: [%d] (from subscription response)\n", extraNonce1, extraNonce2Size)
		msg := &wire.StratumClientExtraNonceMsg{
			ExtraNonce1:     extraNonce1,
			ExtraNonce2Size: extraNonce2Size,
			PoolObserverID:  poolObserverID,
		}
		hubChan <- msg
		pubChan <- publish{Msg: msg, Channel: []byte("extranonce")}
	}

	arr, ok = arr[0].([]interface{})
	if !ok {
		return
	}

	// first element is either an array k,v or an array of arrays
	// [[k,v],[k,v]]

	_, multi := arr[0].([]interface{})
	if !multi {
		arr = []interface{}{[]string{arr[0].(string), arr[1].(string)}}
	}

	for _, e := range arr {

		arr, ok := e.([]interface{})
		if !ok {
			stringArr, ok := e.([]string)
			if ok {
				arr = []interface{}{stringArr[0], stringArr[1]}
			}
		}

		if len(arr) == 2 {
			param := arr[0]
			switch param {
			case "mining.notify":
				subscriptionID = arr[1].(string)
				logging.Debugf("Received subscription ID %s\n", subscriptionID)
			case "mining.set_difficulty":
				// ignore
			default:
				logging.Warnf("Received unexpected parameter in mining.subscribe reply: %s\n", param)
			}
		} else {
			logging.Warnf("Received unexpected mining.subscribe reply parameter length: %d\n", len(arr))
		}
	}
}

func processRemoteInstruction(msg stratum.StratumMessage) bool {
	switch msg.RemoteMethod {
	case "mining.set_difficulty":
		// Adjusted difficulty
		params, ok := msg.Parameters.([]interface{})
		difficulty, ok = params[0].(float64)
		if ok {
			logging.Infof("New difficulty received: %f", difficulty)
			pdmsg := &wire.StratumClientDifficultyMsg{
				Difficulty:     difficulty,
				PoolObserverID: poolObserverID,
			}
			hubChan <- pdmsg
			pubChan <- publish{Msg: pdmsg, Channel: []byte("difficulty")}
		} else {
			logging.Errorf("Could not determine difficulty from stratum: [%v]\n", params[0])
		}
	case "mining.set_target":
		// Adjusted target
		params, ok := msg.Parameters.([]interface{})
		if ok {
			target, ok := params[0].(string)
			if ok {
				logging.Infof("New target received: %s", target)
				targetBytes, err := hex.DecodeString(target)
				if err == nil {
					pdmsg := &wire.StratumClientTargetMsg{
						Target:         targetBytes,
						PoolObserverID: poolObserverID,
					}
					hubChan <- pdmsg
					pubChan <- publish{Msg: pdmsg, Channel: []byte("target")}
				}
			} else {
				logging.Errorf("Could not determine target from stratum: [%v]\n", params[0])
			}
		} else {
			logging.Error("Could not determine target from stratum\n")
		}
	case "mining.set_extranonce":
		// Adjusted extranonce
		extraNonce1, _ := hex.DecodeString(msg.Parameters.([]interface{})[0].(string))
		extraNonce2Size := int8(32 - len(extraNonce1))
		if stratumProtocol == 0 {
			extraNonce2Size = int8(msg.Parameters.([]interface{})[1].(float64))
		}

		logging.Infof("Setting extranonce1 [%x], extranonce2_size: [%d] (from set_extranonce)\n", extraNonce1, extraNonce2Size)

		pdmsg := &wire.StratumClientExtraNonceMsg{
			ExtraNonce1:     extraNonce1,
			ExtraNonce2Size: extraNonce2Size,
			PoolObserverID:  poolObserverID,
		}
		hubChan <- pdmsg
		pubChan <- publish{Msg: pdmsg, Channel: []byte("extranonce")}
	case "mining.notify":
		hubMsg := new(wire.StratumClientJobMsg)
		params := msg.Parameters.([]interface{})
		switch stratumProtocol {
		case 0:
			hubMsg.JobID = []byte(params[0].(string))
			hubMsg.PreviousBlockHash, _ = hex.DecodeString(params[1].(string))
			hubMsg.GenTX1, _ = hex.DecodeString(params[2].(string))
			hubMsg.GenTX2, _ = hex.DecodeString(params[3].(string))
			merkleBranches := params[4].([]interface{})
			merkleBranchBytes := make([][]byte, len(merkleBranches))
			for i, m := range merkleBranches {
				merkleBranchBytes[i], _ = hex.DecodeString(m.(string))
			}
			hubMsg.MerkleBranches = merkleBranchBytes
			hubMsg.BlockVersion, _ = hex.DecodeString(params[5].(string))
			hubMsg.DifficultyBits, _ = hex.DecodeString(params[6].(string))
			hubMsg.Timestamp, _ = strconv.ParseInt(params[7].(string), 16, 64)
			hubMsg.CleanJobs = params[8].(bool)
		case 1, 2: // ZCash, BTG
			hubMsg.JobID = []byte(params[0].(string))
			hubMsg.BlockVersion, _ = hex.DecodeString(params[1].(string))

			hubMsg.PreviousBlockHash, _ = hex.DecodeString(params[2].(string))
			// Encode into stratum format matching non-ZEC/BTG to match storage to blockhashes from full node
			hubMsg.PreviousBlockHash = util.RevHashBytes(util.ReverseByteArray(hubMsg.PreviousBlockHash))
			merkleRoot, _ := hex.DecodeString(params[3].(string))
			hubMsg.MerkleBranches = [][]byte{merkleRoot}
			hubMsg.Reserved, _ = hex.DecodeString(params[4].(string))

			tsBytes, err := hex.DecodeString(params[5].(string))
			if err != nil {
				logging.Errorf("Could not decode timestamp from stratum message: %v", err)
			}
			buf := bytes.NewBuffer(tsBytes)
			var timestamp int32
			err = binary.Read(buf, binary.LittleEndian, &timestamp)
			if err != nil {
				logging.Errorf("Could not decode timestamp from stratum message: %v", err)
			}
			hubMsg.Timestamp = int64(timestamp)
			hubMsg.DifficultyBits, _ = hex.DecodeString(params[6].(string))
			hubMsg.CleanJobs = params[7].(bool)
		}

		hubMsg.PoolObserverID = poolObserverID
		hubMsg.Observed = time.Now().UnixNano()
		hubChan <- hubMsg
	default:
		return false
	}
	return true
}
