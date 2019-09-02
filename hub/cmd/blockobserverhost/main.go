package main

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/mit-dci/pooldetective/logging"
	"github.com/mit-dci/pooldetective/util"
	"github.com/mit-dci/pooldetective/wire"
)

var db *sql.DB

func main() {
	logging.SetLogLevel(util.GetLoglevelFromEnv())
	logging.Debugf("Starting Block Observer Host...")
	var err error
	db, err = sql.Open("postgres", os.Getenv("PGSQL_CONNECTION"))
	if err != nil {
		logging.Fatal(err)
	}

	rep, err := wire.NewServer(wire.PortBlockObserver)
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

func processMessage(msg wire.PoolDetectiveMsg) wire.PoolDetectiveMsg {
	var err error
	switch t := msg.(type) {
	case *wire.BlockObserverBlockObservedMsg:
		err = processBlockObserved(t)
	case *wire.BlockObserverGetCoinsRequestMsg:
		return processGetCoins(t)
	default:
		msgid := wire.GetMessageID(msg)
		return &wire.ErrorMsg{YourID: msgid, Error: fmt.Sprintf("Unexpected message type %T received", msg)}
	}

	msgid := wire.GetMessageID(msg)
	if err == nil {
		return &wire.AckMsg{YourID: msgid}
	}
	return &wire.ErrorMsg{YourID: msgid, Error: err.Error()}
}

func processGetCoins(msg *wire.BlockObserverGetCoinsRequestMsg) wire.PoolDetectiveMsg {
	msgid := wire.GetMessageID(msg)
	rows, err := db.Query("select id, trim(ticker) from coins")
	if err != nil {
		return &wire.ErrorMsg{YourID: msgid, Error: err.Error()}
	}

	reply := &wire.BlockObserverGetCoinsResponseMsg{Coins: []wire.BlockObserverGetCoinsResponseCoin{}}
	for rows.Next() {
		coin := wire.BlockObserverGetCoinsResponseCoin{}
		err = rows.Scan(&coin.CoinID, &coin.Ticker)
		if err != nil {
			return &wire.ErrorMsg{YourID: msgid, Error: err.Error()}
		}
		reply.Coins = append(reply.Coins, coin)
	}
	return reply
}

func processBlockObserved(msg *wire.BlockObserverBlockObservedMsg) error {
	_, err := db.Exec("insert into blocks(coin_id, block_hash) values ($1,$2) on conflict do nothing", msg.CoinID, msg.BlockHash[:])
	if err != nil {
		return err
	}
	var id int64
	err = db.QueryRow("select id from blocks where coin_id=$1 and block_hash=$2", msg.CoinID, msg.BlockHash[:]).Scan(&id)
	if err != nil {
		return err
	}
	ip := net.IPAddr{net.IP(msg.PeerIP), ""}
	_, err = db.Exec("insert into block_observations(location_id, block_id, observed, peer_ip, peer_port) values ($1, $2, $3, $4::INET, $5)", msg.LocationID, id, time.Unix(0, msg.Observed), ip.String(), msg.PeerPort)
	return err
}
