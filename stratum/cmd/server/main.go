package main

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/mit-dci/pooldetective/logging"
	"github.com/mit-dci/pooldetective/stratum"
	"github.com/mit-dci/pooldetective/util"
	"github.com/mit-dci/pooldetective/wire"
)

type StratumJob struct {
	ID              int64
	PoolObserverID  int
	PoolJobID       []byte
	Difficulty      float64
	ExtraNonce1     []byte
	ExtraNonce2Size int8
	Target          []byte
	JobData         []interface{}
}

func (job StratumJob) AsJobData(protocol int) []interface{} {
	return append([]interface{}{fmt.Sprintf("%x", job.ID)}, job.JobData[1:]...)
}

func (job StratumJob) AsJobDataWithNewTimestamp() []interface{} {
	return job.AsJobDataWithNewTimestampForProtocol(0)
}

func (job StratumJob) AsJobDataWithNewTimestampForProtocol(protocol int) []interface{} {
	j := job.AsJobData(protocol)
	switch protocol {
	case 0:
		return append(j[:7], fmt.Sprintf("%x", time.Now().Unix()), true)
	case 1, 2:
		var timestampBytes bytes.Buffer
		binary.Write(&timestampBytes, binary.LittleEndian, int32(time.Now().Unix()))

		j[5] = fmt.Sprintf("%x", timestampBytes.Bytes())
		j[7] = true
		return j
	}
	return []interface{}{}
}

type StratumClient struct {
	ID                      int64
	WorkingOnJobID          int64
	WorkingOnPoolObserverID int
	WorkingOnPoolJobID      []byte
	conn                    *stratum.StratumConnection
	Difficulty              float64
	ExtraNonce1             []byte
	ExtraNonce2Size         int8
	Target                  []byte
	SubscribedToExtraNonce  bool
	Disconnected            bool
}

type ShareNotify struct {
	PoolObserverID int
	ShareMsg       *wire.StratumClientSubmitShareMsg
}

var db *sql.DB
var stratumConn *stratum.StratumConnection
var nextClientID int64 = 1
var algorithmID int
var stratumProtocol int
var clients = sync.Map{}
var sharesChan = make(chan ShareNotify, 100)

func main() {
	logging.SetLogLevel(util.GetLoglevelFromEnv())

	var err error
	db, err = sql.Open("postgres", os.Getenv("PGSQL_CONNECTION"))
	if err != nil {
		logging.Fatal(err)
	}

	port, err := strconv.Atoi(os.Getenv("STRATUMPORT"))
	if err != nil {
		panic(err)
	}

	algorithmID, err = strconv.Atoi(os.Getenv("ALGORITHMID"))
	if err != nil {
		logging.Fatalf("Could not parse ALGORITHMID from environment: %v", err)
	}

	stratumProtocol, err = strconv.Atoi(os.Getenv("STRATUMPROTOCOL"))
	if err != nil {
		stratumProtocol = 0
	}

	srv, err := stratum.NewStratumListener(port)
	if err != nil {
		panic(err)
	}

	go func() {
		for {

			pub, err := wire.NewClient(os.Getenv("HUBHOST"), wire.PortPubSubPublishers)
			if err != nil {
				panic(err)
			}

			for s := range sharesChan {
				err := pub.Publish([]byte(fmt.Sprintf("sc-%03d", s.PoolObserverID)), s.ShareMsg)
				if err != nil {
					logging.Errorf("[PubSubPublisher] Could not publish: %v", err)
					sharesChan <- s
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
			logging.Debugf("[PubSubSubscriber] (Re)connecting to host")
			sub, err := wire.NewSubscriber(os.Getenv("HUBHOST"), wire.PortPubSubSubscribers)
			if err != nil {
				logging.Warnf("Could not connect to hub host: %v", err)
				time.Sleep(time.Second * 5)
				continue
			}

			logging.Debugf("[PubSubSubscriber] Subscribing to topics")
			err = sub.Subscribe([]byte("extranonce"))
			if err == nil {
				err = sub.Subscribe([]byte("difficulty"))
				if err == nil {
					err = sub.Subscribe([]byte("target"))
				}
			}
			if err != nil {
				logging.Warnf("[PubSubSubscriber] Could not subscribe to topic on hub: %v", err)
				time.Sleep(time.Second * 5)
				continue
			}

			for {
				msg, _, _, err := sub.RecvSub()
				if err != nil {
					logging.Errorf("[PubSubSubscriber] Error receiving subscription message: %v", err)
					break
				}

				poolObserverID := 0
				switch t := msg.(type) {
				case *wire.StratumClientDifficultyMsg:
					poolObserverID = t.PoolObserverID
				case *wire.StratumClientExtraNonceMsg:
					poolObserverID = t.PoolObserverID
				case *wire.StratumClientTargetMsg:
					poolObserverID = t.PoolObserverID
				}

				clients.Range(func(k interface{}, c interface{}) bool {
					clt, ok := c.(*StratumClient)
					if !ok {
						return true
					}

					if clt.WorkingOnPoolObserverID != poolObserverID {
						return true
					}

					newDifficulty := clt.Difficulty
					newExtraNonce1 := clt.ExtraNonce1
					newExtraNonce2Size := clt.ExtraNonce2Size
					newTarget := clt.Target

					switch t := msg.(type) {
					case *wire.StratumClientDifficultyMsg:
						newDifficulty = t.Difficulty
					case *wire.StratumClientExtraNonceMsg:
						newExtraNonce1 = t.ExtraNonce1
						newExtraNonce2Size = t.ExtraNonce2Size
					case *wire.StratumClientTargetMsg:
						newTarget = t.Target
					}

					clt.SendWork(StratumJob{
						ID:              clt.WorkingOnJobID,
						PoolObserverID:  clt.WorkingOnPoolObserverID,
						Difficulty:      newDifficulty,
						ExtraNonce1:     newExtraNonce1,
						ExtraNonce2Size: newExtraNonce2Size,
						Target:          newTarget,
					})

					return true
				})
			}
			sub.Close()
			time.Sleep(time.Second)
		}
	}()

	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				checkClientsWork()
			}
		}
	}()

	for {
		conn, err := srv.Accept()

		clientID := atomic.AddInt64(&nextClientID, 1)

		if err != nil {
			panic(err)
		}

		clt := StratumClient{
			ID:                     clientID,
			conn:                   conn,
			Difficulty:             -100,
			ExtraNonce2Size:        -10,
			ExtraNonce1:            []byte{0x00, 0x01},
			Target:                 []byte{},
			SubscribedToExtraNonce: false,
		}

		conn.LogOutput = func(logs []stratum.CommEvent) {
			for _, l := range logs {
				j, _ := json.Marshal(l.Message)
				dir := "> "
				if l.In {
					dir = "< "
				}
				logging.Debugf("%s %s", dir, string(j))
			}
		}

		clients.Store(clientID, &clt)

		go serveClient(&clt)
	}
}

func serveClient(client *StratumClient) {
	logging.Debugf("New stratum client connected: %d", client.ID)
	for {
		close := false
		select {
		case msg := <-client.conn.Incoming:
			processStratumMessage(client, msg)
		case <-client.conn.Disconnected:
			logging.Warnf("Stratum client %d disconnected", client.ID)
			close = true
		}

		if close {
			clients.Delete(client.ID)
			break
		}
	}
	client.Disconnected = true
	time.Sleep(time.Second)
	client.conn.Stop()
}

func (client *StratumClient) SendWork(job StratumJob) {
	logging.Debugf("Sending job %s to client\n", job.ID)
	if client.Disconnected {
		logging.Debugf("Client is disconnected, returning\n")
		return
	}
	if !bytes.Equal(client.ExtraNonce1, job.ExtraNonce1) || client.ExtraNonce2Size != job.ExtraNonce2Size {
		logging.Debugf("Sending extranonce info\n")
		client.conn.Outgoing <- stratum.StratumMessage{
			RemoteMethod: "mining.set_extranonce",
			Parameters: []interface{}{
				fmt.Sprintf("%x", job.ExtraNonce1),
				job.ExtraNonce2Size,
			},
		}
		client.ExtraNonce2Size = job.ExtraNonce2Size
		client.ExtraNonce1 = job.ExtraNonce1
	}

	if client.Difficulty != job.Difficulty && stratumProtocol == 0 { // Not for ZEC/BTG
		logging.Debugf("Sending difficulty info\n")
		client.conn.Outgoing <- stratum.StratumMessage{
			RemoteMethod: "mining.set_difficulty",
			Parameters:   []interface{}{job.Difficulty},
		}
		client.Difficulty = job.Difficulty
	}

	if !bytes.Equal(client.Target, job.Target) && stratumProtocol == 1 { // Only for ZEC/BTG
		client.conn.Outgoing <- stratum.StratumMessage{
			RemoteMethod: "mining.set_target",
			Parameters:   []interface{}{hex.EncodeToString(job.Target)},
		}
		client.Target = make([]byte, len(job.Target))
		copy(client.Target[:], job.Target)
	}

	if client.WorkingOnJobID != job.ID && len(job.JobData) > 0 { // Don't do this for diff/extranonce only stuff
		logging.Debugf("Sending job info\n")

		client.conn.Outgoing <- stratum.StratumMessage{
			RemoteMethod: "mining.notify",
			Parameters:   job.AsJobDataWithNewTimestampForProtocol(stratumProtocol),
		}
		client.WorkingOnJobID = job.ID
		client.WorkingOnPoolJobID = job.PoolJobID
		client.WorkingOnPoolObserverID = job.PoolObserverID
	}
}

func getTopJobForPoolObserver(poolObserverID int) (int64, error) {
	out := int64(0)
	err := db.QueryRow(`select j.id FROM jobs j where j.pool_observer_id=$1 ORDER BY j.id desc limit 1`, poolObserverID).Scan(&out)
	return out, err
}

func storeShare(jobID int64, extranonce2 []byte, timestamp int64, nonce int64, stale bool, poolObserverID int, additionalSolutionData [][]byte) (int64, error) {
	byteaarr, err := pq.ByteaArray(additionalSolutionData).Value()
	if err != nil {
		return 0, err
	}

	var shareID int64
	err = db.QueryRow(`insert into shares(job_id, extranonce2, timestamp, nonce, found, stale, additional_solution_data) values ($1, $2, $3, $4, now(), $5, $6) RETURNING id`, jobID, extranonce2, timestamp, nonce, stale, byteaarr).Scan(&shareID)
	if err != nil {
		return 0, err
	}

	_, err = db.Exec(`update pool_observers set last_share_found=NOW(), last_share_id=$1 WHERE id=$2`, shareID, poolObserverID)
	return shareID, err
}

func getTopJob() (StratumJob, error) {
	var poolObserverID int
	var jobID int64

	var difficulty sql.NullFloat64
	var extranonce1 []byte
	var extranonce2size sql.NullInt32
	var poolJobID []byte
	var previousBlockHash []byte
	var genTx1 []byte
	var genTx2 []byte
	var merkleBranches pq.ByteaArray
	var blockVersion []byte
	var difficultyBits []byte
	var timestamp int64
	var cleanJobs bool
	var stratumTarget []byte
	var reserved []byte

	err := db.QueryRow(`SELECT 
							po.id,
							j.id,
							stratum_difficulty, 
							stratum_extranonce1, 
							stratum_extranonce2size,
							pool_job_id, 
							previous_block_hash, 
							generation_transaction_part_1,
							generation_transaction_part_2, 
							merkle_branches, 
							block_version,
							difficulty_bits, 
							timestamp, 
							clean_jobs,
							stratum_target,
							reserved
						from 	
							pool_observers po 
							left join jobs j on j.id=po.last_job_id 
						WHERE 
							po.disabled=false 
							AND po.last_job_id IS NOT NULL
							AND po.algorithm_id=$1
							AND j.observed > (NOW() - INTERVAL '15 minutes')
						ORDER BY 
							po.last_share_found nulls first limit 1 `, algorithmID).Scan(
		&poolObserverID,
		&jobID,
		&difficulty,
		&extranonce1,
		&extranonce2size,
		&poolJobID,
		&previousBlockHash,
		&genTx1,
		&genTx2,
		&merkleBranches,
		&blockVersion,
		&difficultyBits,
		&timestamp,
		&cleanJobs,
		&stratumTarget,
		&reserved)

	if err != nil {
		return StratumJob{}, err
	}

	merkleBranchesString := make([]string, len(merkleBranches))
	for i, m := range merkleBranches {
		// Convert back to stratum format!
		merkleBranchesString[i] = fmt.Sprintf("%x", m)
	}

	diff := float64(128)
	if difficulty.Valid {
		diff = difficulty.Float64
	}

	extranonce2sizeInt8 := int8(0)
	if extranonce2size.Valid {
		extranonce2sizeInt8 = int8(extranonce2size.Int32)
	}

	job := StratumJob{
		ID:              jobID,
		PoolObserverID:  poolObserverID,
		Difficulty:      diff,
		ExtraNonce1:     extranonce1,
		ExtraNonce2Size: extranonce2sizeInt8,
		PoolJobID:       poolJobID,
		Target:          stratumTarget,
	}

	switch stratumProtocol {
	case 0:
		job.JobData = []interface{}{
			fmt.Sprintf("%x", poolJobID),
			// Convert back to stratum format!
			fmt.Sprintf("%x", util.RevHashBytes(util.ReverseByteArray(previousBlockHash))),
			fmt.Sprintf("%x", genTx1),
			fmt.Sprintf("%x", genTx2),
			merkleBranchesString,
			fmt.Sprintf("%x", blockVersion),
			fmt.Sprintf("%x", difficultyBits),
			fmt.Sprintf("%x", timestamp),
			cleanJobs,
		}
	case 1, 2: // ZCash, BTG
		var timestampBytes bytes.Buffer
		binary.Write(&timestampBytes, binary.LittleEndian, timestamp)

		job.JobData = []interface{}{
			fmt.Sprintf("%x", poolJobID),
			fmt.Sprintf("%x", blockVersion),
			// Convert back to stratum format!
			fmt.Sprintf("%x", util.ReverseByteArray(previousBlockHash)),
			merkleBranchesString[0],
			fmt.Sprintf("%x", reserved),
			fmt.Sprintf("%x", timestampBytes.Bytes()),
			fmt.Sprintf("%x", difficultyBits),
			cleanJobs,
		}

		if stratumProtocol == 2 {
			job.JobData = append(job.JobData, []interface{}{"144_5", "BgoldPoW"}...)
		}
	}

	return job, nil
}

func (client *StratumClient) SendNewJob() {
	job, err := getTopJob()
	if err != nil {
		logging.Warnf("Could not send client work: %s\n", err.Error())
		return
	}
	client.SendWork(job)
}

func getNonceAndDiffData(poolObserverID int) (difficulty float64, extranonce1 []byte, extranonce2size int8, err error) {
	err = db.QueryRow("select stratum_difficulty, stratum_extranonce1, stratum_extranonce2size from pool_observers where id=$1", poolObserverID).Scan(&difficulty, &extranonce1, &extranonce2size)
	return
}

func getJobPoolInfo(jobID int64) (poolJobID []byte, poolObserverID int, stale bool, err error) {
	err = db.QueryRow("select pool_job_id, pool_observer_id, NOT (SELECT j.id IN (select j2.id FROM jobs j2 where j2.pool_observer_id=j.pool_observer_id ORDER BY j2.id desc limit 3)) from jobs j where id=$1", jobID).Scan(&poolJobID, &poolObserverID, &stale)
	return
}

func checkClientsWork() {
	clients.Range(func(k interface{}, c interface{}) bool {
		clt, ok := c.(*StratumClient)
		if !ok {
			return true
		}
		validJob := false
		jobID, _ := getTopJobForPoolObserver(clt.WorkingOnPoolObserverID)

		if clt.WorkingOnJobID == jobID {
			validJob = true
		}

		if !validJob {
			clt.SendNewJob()
		}
		return true
	})
}

func processStratumMessage(client *StratumClient, msg stratum.StratumMessage) {
	err := msg.Error
	if err != nil {
		logging.Warnf("Error response received: %v\n", err)
	}

	if client.Disconnected {
		logging.Debugf("Client is disconnected, returning\n")
		return
	}

	switch msg.RemoteMethod {
	case "mining.authorize":
		client.conn.Outgoing <- stratum.StratumMessage{
			MessageID: msg.Id(),
			Result:    true,
		}
	case "mining.extranonce.subscribe":
		logging.Debugf("Client subscribed to extranonce!")
		client.SubscribedToExtraNonce = true
		client.conn.Outgoing <- stratum.StratumMessage{
			MessageID: msg.Id(),
			Result:    true,
		}
		client.conn.Outgoing <- stratum.StratumMessage{
			RemoteMethod: "mining.set_extranonce",
			Parameters: []interface{}{
				fmt.Sprintf("%x", client.ExtraNonce1),
				client.ExtraNonce2Size,
			},
		}

	case "mining.subscribe":
		b := make([]byte, 8)
		rand.Read(b)
		clientID := fmt.Sprintf("%x", b)
		job, err := getTopJob()
		if err != nil {
			logging.Errorf("Can't get top job in mining.subscribe")
		}

		switch stratumProtocol {
		case 0:
			client.conn.Outgoing <- stratum.StratumMessage{
				MessageID: msg.Id(),
				Result: []interface{}{
					[][]string{[]string{"mining.notify", clientID}, []string{"mining.set_difficulty", clientID}},
					fmt.Sprintf("%x", job.ExtraNonce1),
					job.ExtraNonce2Size,
				},
			}
			client.ExtraNonce2Size = job.ExtraNonce2Size
		case 1, 2:
			client.conn.Outgoing <- stratum.StratumMessage{
				MessageID: msg.Id(),
				Result: []interface{}{
					clientID,
					fmt.Sprintf("%x", job.ExtraNonce1),
				},
			}
			client.ExtraNonce2Size = int8(32 - len(job.ExtraNonce1))
		}
		client.ExtraNonce1 = job.ExtraNonce1
		client.SendNewJob()
	case "mining.configure":
		client.conn.Outgoing <- stratum.StratumMessage{
			MessageID: msg.Id(),
			Result:    nil,
		}
	case "mining.get_transactions":
		client.conn.Outgoing <- stratum.StratumMessage{
			MessageID: msg.Id(),
			Result:    []interface{}{},
		}
	case "mining.submit":
		params := msg.Parameters.([]interface{})
		success := true
		jobID, err := strconv.ParseInt(params[1].(string), 16, 64)
		if err != nil {
			logging.Warnf("Error parsing Job ID: %s", err.Error())
			success = false
		}

		var extranonce2 []byte
		var additionalSolutionData [][]byte
		var timestamp int64
		var nonce int64

		switch stratumProtocol {
		case 0:
			extranonce2, err = hex.DecodeString(params[2].(string))
			if err != nil {
				logging.Warnf("Error parsing extranonce2: %s", err.Error())
				success = false
			}

			timestamp, err = strconv.ParseInt(params[3].(string), 16, 64)
			if err != nil {
				logging.Warnf("Error parsing timestamp: %s", err.Error())
				success = false
			}

			nonce, err = strconv.ParseInt(params[4].(string), 16, 64)
			if err != nil {
				logging.Warnf("Error parsing nonce: %s", err.Error())
				success = false
			}
		case 1, 2: // ZCash, BTG
			extranonce2, err = hex.DecodeString(params[3].(string))
			if err != nil {
				logging.Warnf("Error parsing extranonce2: %s", err.Error())
				success = false
			}

			tsBytes, err := hex.DecodeString(params[2].(string))
			if err != nil {
				logging.Errorf("Could not decode timestamp from stratum message: %v", err)
			}
			buf := bytes.NewBuffer(tsBytes)
			var timestamp32 int32
			err = binary.Read(buf, binary.LittleEndian, &timestamp32)
			if err != nil {
				logging.Errorf("Could not decode timestamp from stratum message: %v", err)
			}

			timestamp = int64(timestamp32)
			additionalSolutionData = make([][]byte, 1)
			additionalSolutionData[0], err = hex.DecodeString(params[4].(string))
			if err != nil {
				logging.Warnf("Error parsing equihash solution: %s", err.Error())
				success = false
			}
		}

		if success {
			poolJobID, poolObserverID, stale, err := getJobPoolInfo(jobID)
			if err != nil {
				logging.Warnf("Error storing share: %s", err.Error())
				success = false
			}

			shareID, err := storeShare(jobID, extranonce2, timestamp, nonce, stale, poolObserverID, additionalSolutionData)
			if err != nil {
				logging.Warnf("Error storing share: %s", err.Error())
				success = false
			}

			if success && !stale {
				sharesChan <- ShareNotify{
					PoolObserverID: poolObserverID,
					ShareMsg: &wire.StratumClientSubmitShareMsg{
						ShareID:                shareID,
						PoolJobID:              poolJobID,
						ExtraNonce2:            extranonce2,
						Time:                   timestamp,
						Nonce:                  nonce,
						AdditionalSolutionData: additionalSolutionData,
					},
				}
			}
		}
		client.conn.Outgoing <- stratum.StratumMessage{
			MessageID: msg.Id(),
			Result:    success,
		}
		client.SendNewJob()
	default:
		logging.Warnf("Received unknown message [%s]\n", msg.RemoteMethod)
	}
}
