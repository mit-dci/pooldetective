package main

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/mit-dci/pooldetective/logging"
	"github.com/mit-dci/pooldetective/util"
	"github.com/mit-dci/pooldetective/wire"
)

var db *sql.DB

type publish struct {
	Channel []byte
	Msg     wire.PoolDetectiveMsg
}

var pubChan = make(chan publish, 100)

// Caches connections mapped by their PoolObserverID
var connections = sync.Map{}

func main() {
	logging.SetLogLevel(util.GetLoglevelFromEnv())
	logging.Debugf("Starting Coordinator Host...")
	var err error
	db, err = sql.Open("postgres", os.Getenv("PGSQL_CONNECTION"))
	if err != nil {
		logging.Fatal(err)
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

	go monitor()

	rep, err := wire.NewServer(wire.PortCoordinator)
	if err != nil {
		logging.Fatal(err)
	}
	defer rep.Close()
	for {
		msg, ok, err := rep.Recv()
		if err != nil && !ok {
			logging.Error(err)
			time.Sleep(time.Millisecond * 500)
			continue
		}

		var response wire.PoolDetectiveMsg
		if err != nil {
			response = &wire.ErrorMsg{YourID: 0, Error: fmt.Sprintf("Unable to parse request: %s", err.Error())}
		} else {
			response = processMessage(msg)
		}

		errMsg, ok := response.(*wire.ErrorMsg)
		if ok {
			logging.Error(errMsg.Error)
		}

		rep.Send(response)
	}

}

func monitor() {
	for {
		// If pool observers don't register work for over 10 minutes we should restart them
		time.Sleep(time.Minute * 1)

		rows, err := db.Query("SELECT location_id, id FROM pool_observers WHERE disabled=false AND last_job_received < (NOW() - INTERVAL '10 MINUTES')")
		if err != nil {
			continue
		}
		for rows.Next() {
			lid := int(0)
			poid := int(0)
			err = rows.Scan(&lid, &poid)
			if err != nil {
				continue
			}

			logging.Warnf("Pool observer %d has not received work for more than 10 minutes, signaling a restart", poid)

			pubChan <- publish{
				Channel: []byte(fmt.Sprintf("co-%03d", lid)),
				Msg: &wire.CoordinatorRestartPoolObserverMsg{
					PoolObserverID: poid,
				},
			}
		}

		rows, err = db.Query("SELECT id FROM locations WHERE reload_coordinator=true")
		if err != nil {
			continue
		}
		for rows.Next() {
			lid := int(0)
			err = rows.Scan(&lid)
			if err != nil {
				continue
			}

			logging.Infof("Signaling a config reload to location %d", lid)

			pubChan <- publish{
				Channel: []byte(fmt.Sprintf("co-%03d", lid)),
				Msg:     &wire.CoordinatorRefreshConfigMsg{},
			}
			db.Exec("UPDATE locations SET reload_coordinator=false WHERE id=$1", lid)
		}
	}
}

func processMessage(msg wire.PoolDetectiveMsg) wire.PoolDetectiveMsg {

	var err error
	var reply wire.PoolDetectiveMsg
	switch t := msg.(type) {
	case *wire.CoordinatorGetConfigRequestMsg:
		reply, err = processGetConfigMessage(t)
	default:
		err = fmt.Errorf("Unrecognized message type %T", t)
	}

	msgID := wire.GetMessageID(msg)
	if err != nil {
		return &wire.ErrorMsg{YourID: msgID, Error: err.Error()}
	}
	if reply == nil {
		return &wire.AckMsg{YourID: msgID}
	}
	return reply
}

func processGetConfigMessage(msg *wire.CoordinatorGetConfigRequestMsg) (wire.PoolDetectiveMsg, error) {
	response := new(wire.CoordinatorGetConfigResponseMsg)

	servers, err := getStratumServers(msg.LocationID)
	if err != nil {
		return nil, err
	}

	clients, err := getStratumClients(msg.LocationID)
	if err != nil {
		return nil, err
	}

	response.StratumClientPoolObserverIDs = clients
	response.StratumServers = servers

	return response, nil
}

func getStratumServers(locationID int) ([]wire.CoordinatorGetConfigResponseStratumServer, error) {
	rows, err := db.Query("SELECT id, algorithm_id, port, stratum_protocol FROM stratum_servers WHERE location_id=$1", locationID)
	if err != nil {
		return nil, err
	}

	results := make([]wire.CoordinatorGetConfigResponseStratumServer, 0)
	for rows.Next() {
		result := wire.CoordinatorGetConfigResponseStratumServer{}
		err = rows.Scan(&result.ID, &result.AlgorithmID, &result.Port, &result.Protocol)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func getStratumClients(locationID int) ([]int, error) {
	rows, err := db.Query("SELECT id FROM pool_observers WHERE location_id=$1 and disabled=false", locationID)
	if err != nil {
		return nil, err
	}

	results := make([]int, 0)
	for rows.Next() {
		result := int(0)
		err = rows.Scan(&result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}
