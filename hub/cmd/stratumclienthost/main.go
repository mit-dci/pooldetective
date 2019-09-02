package main

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/lib/pq"
	"github.com/mit-dci/pooldetective/logging"
	"github.com/mit-dci/pooldetective/util"
	"github.com/mit-dci/pooldetective/wire"
)

var db *sql.DB

// Caches connections mapped by their PoolObserverID
var connections = sync.Map{}

func main() {
	logging.SetLogLevel(util.GetLoglevelFromEnv())
	logging.Debugf("Starting Stratum Client Host...")
	var err error
	db, err = sql.Open("postgres", os.Getenv("PGSQL_CONNECTION"))
	if err != nil {
		logging.Fatal(err)
	}

	rep, err := wire.NewServer(wire.PortStratumClient)
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
	var reply wire.PoolDetectiveMsg
	switch t := msg.(type) {
	case *wire.StratumClientExtraNonceMsg:
		err = processExtraNonce(t)
	case *wire.StratumClientDifficultyMsg:
		err = processDifficulty(t)
	case *wire.StratumClientTargetMsg:
		err = processTarget(t)
	case *wire.StratumClientLoginDetailsRequestMsg:
		reply, err = processStratumDetailsRequest(t)
	case *wire.StratumClientConnectionEventMsg:
		err = processConnectionEvent(t)
	case *wire.StratumClientShareEventMsg:
		err = processShareEvent(t)
	case *wire.StratumClientJobMsg:
		err = processJob(t)
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

func processShareEvent(msg *wire.StratumClientShareEventMsg) error {
	field := ""
	accepted := false
	switch msg.Event {
	case wire.ShareEventSubmitted:
		field = "submitted"
	case wire.ShareEventAccepted:
		accepted = true
		field = "responsereceived"
	case wire.ShareEventDeclined:
		field = "responsereceived"
	}
	_, err := db.Exec(fmt.Sprintf("update shares set %s=$1, accepted=$2, details=$3 where id=$4", field), time.Unix(0, msg.Observed), accepted, msg.Details, msg.ShareID)
	return err
}

func processJob(msg *wire.StratumClientJobMsg) error {

	// Translate hash format
	h, err := util.DecodeStratumHash(msg.PreviousBlockHash)
	if err != nil {
		return err
	}
	msg.PreviousBlockHash = h.CloneBytes()

	merkleBranches, err := pq.ByteaArray(msg.MerkleBranches).Value()
	if err != nil {
		return err
	}
	var jobID int64
	err = db.QueryRow(`insert into jobs(
							observed, pool_observer_id, pool_job_id, previous_block_hash, 
							generation_transaction_part_1, generation_transaction_part_2, 
							merkle_branches, block_version, difficulty_bits, clean_jobs, timestamp, reserved) 
						values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) returning id`,
		time.Unix(0, msg.Observed),
		msg.PoolObserverID,
		msg.JobID,
		msg.PreviousBlockHash,
		msg.GenTX1,
		msg.GenTX2,
		merkleBranches,
		msg.BlockVersion,
		msg.DifficultyBits,
		msg.CleanJobs,
		msg.Timestamp,
		msg.Reserved).Scan(&jobID)
	if err != nil {
		return err
	}

	_, err = db.Exec(`update pool_observers set last_job_received=NOW(), last_job_id=$1 WHERE id=$2`, jobID, msg.PoolObserverID)
	return err

}

func processExtraNonce(msg *wire.StratumClientExtraNonceMsg) error {
	_, err := db.Exec("update pool_observers set stratum_extranonce1=$1, stratum_extranonce2size=$2 where id=$3", msg.ExtraNonce1, msg.ExtraNonce2Size, msg.PoolObserverID)
	return err
}

func processDifficulty(msg *wire.StratumClientDifficultyMsg) error {
	_, err := db.Exec("update pool_observers set stratum_difficulty=$1 where id=$2", msg.Difficulty, msg.PoolObserverID)
	return err
}

func processTarget(msg *wire.StratumClientTargetMsg) error {
	_, err := db.Exec("update pool_observers set stratum_target=$1 where id=$2", msg.Target, msg.PoolObserverID)
	return err
}

func processStratumDetailsRequest(msg *wire.StratumClientLoginDetailsRequestMsg) (wire.PoolDetectiveMsg, error) {
	response := &wire.StratumClientLoginDetailsResponseMsg{}
	err := db.QueryRow("select stratum_host, stratum_port, stratum_username, COALESCE(stratum_password,''), stratum_protocol from pool_observers where id=$1", msg.PoolObserverID).Scan(&response.Host, &response.Port, &response.Login, &response.Password, &response.Protocol)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func processConnectionEvent(msg *wire.StratumClientConnectionEventMsg) error {
	_, err := db.Exec("insert into pool_observer_events (pool_observer_id, event, timestamp) values ($1,$2,$3)", msg.PoolObserverID, msg.Event, time.Unix(0, msg.Observed))
	return err
}
