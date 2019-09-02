package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/mit-dci/pooldetective/blockobserver/wire"
	"github.com/mit-dci/pooldetective/logging"
	"github.com/mit-dci/pooldetective/util"
)

var db *sql.DB
var batchChan chan []Job
var wg sync.WaitGroup
var poolObservers map[int]PoolObserver

type PoolObserver struct {
	ID        int
	CoinID    int
	NonceSize int
}

type UsedScript struct {
	PoolID    int
	Script    []byte
	FirstSeen time.Time
	LastSeen  time.Time
	Address   string
}

func main() {
	var err error
	logging.SetLogLevel(util.GetLoglevelFromEnv())
	wg = sync.WaitGroup{}
	batchChan = make(chan []Job, 5)
	connStr := os.Getenv("PGSQL_CONNECTION")
	db, err = sql.Open("postgres", connStr)
	poolObservers = map[int]PoolObserver{}
	if err != nil {
		panic(err)
	}

	// Fetch Pool Observer Data
	rows, err := db.Query(`SELECT id, coin_id, COALESCE(stratum_extranonce2size,4) + length(COALESCE(stratum_extranonce1,decode('0000','hex'))) as extranonce_size FROM pool_observers`)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		po := PoolObserver{}
		err = rows.Scan(&po.ID, &po.CoinID, &po.NonceSize)
		if err != nil {
			panic(err)
		}
		poolObservers[po.ID] = po
	}

	go func() {
		wg.Add(1)
		processBatches()
		wg.Done()
	}()

	for {
		i, err := processJobCoinbaseData()

		if err != nil {
			panic(err)
		}
		if i < 10000 {
			break
		}
	}
	close(batchChan)
	wg.Wait()

	usedScripts := make([]UsedScript, 0)

	logging.Debugf("Processing data...")

	rows, err = db.Query(`SELECT 
					p.id,
					jco.output_script,
					min(j.observed) as first_seen,
					max(j.observed) as last_seen
				FROM 
					job_coinbase_outputs jco
					LEFT JOIN jobs j on j.id=jco.job_id
					LEFT JOIN pool_observers po on po.id=j.pool_observer_id
					LEFT JOIN pools p on p.id=po.pool_id
				WHERE
					jco.value > 0 AND p.name != 'NiceHash'
				GROUP BY p.id, jco.output_script`)

	if err != nil {
		panic(err)
	}

	i := 0
	for rows.Next() {
		i++
		s := UsedScript{}
		rows.Scan(&s.PoolID, &s.Script, &s.FirstSeen, &s.LastSeen)
		s.Address = "Invalid"
		usedScripts = append(usedScripts, s)
	}

	logging.Debugf("Parsing scripts")
	for i := range usedScripts {
		s, err := txscript.ParsePkScript(usedScripts[i].Script)
		if err == nil {
			addr, err := s.Address(&chaincfg.MainNetParams)
			if err == nil {
				usedScripts[i].Address = addr.String()
			}
		}
	}

	logging.Debugf("Beginning TX")
	// Process Data
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	_, err = tx.Exec(`CREATE TABLE public.analysis_pool_output_scripts_new
	(
		pool_id int,
		output_script bytea,
		address varchar(50),
		first_seen timestamp without time zone,
		last_seen timestamp without time zone
	)`)
	if err != nil {
		panic(err)
	}

	logging.Debugf("Inserting new scripts and addresses")
	stmt, err := tx.Prepare("INSERT INTO analysis_pool_output_scripts_new(pool_id, output_script, address, first_seen, last_seen) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		panic(err)
	}
	for _, s := range usedScripts {
		_, err = stmt.Exec(s.PoolID, s.Script, s.Address, s.FirstSeen, s.LastSeen)
		if err != nil {
			panic(err)
		}
	}
	logging.Debugf("Committing new data")
	err = tx.Commit()
	if err != nil {
		panic(err)
	}

	logging.Debugf("Making new data active")

	tx, err = db.Begin()
	if err != nil {
		panic(err)
	}

	_, err = tx.Exec(`ALTER TABLE analysis_pool_output_scripts RENAME TO analysis_pool_output_scripts_old`)
	if err != nil {
		panic(err)
	}
	_, err = tx.Exec(`ALTER TABLE analysis_pool_output_scripts_new	RENAME TO analysis_pool_output_scripts`)
	if err != nil {
		panic(err)
	}
	_, err = tx.Exec(`DROP TABLE analysis_pool_output_scripts_old`)
	if err != nil {
		panic(err)
	}
	err = tx.Commit()
	if err != nil {
		panic(err)
	}
}

type Job struct {
	jobID          int64
	poolObserverID int
	genTX1         []byte
	genTX2         []byte
}

func processBatches() {
	for jobs := range batchChan {
		tx, err := db.Begin()
		if err != nil {
			panic(err)
		}
		stmt, err := tx.Prepare(pq.CopyIn("job_coinbase_outputs", "job_id", "output_script", "value", "output_index"))
		if err != nil {
			panic(err)
		}

		errors := 0
		for _, j := range jobs {
			po := poolObservers[j.poolObserverID]
			err = processJob(stmt, j.jobID, po.NonceSize, j.genTX1, j.genTX2)
			if err != nil {
				errors++
			}
		}

		_, err = stmt.Exec()
		if err != nil {
			panic(err)
		}

		err = stmt.Close()
		if err != nil {
			panic(err)
		}

		err = tx.Commit()
		if err != nil {
			panic(err)
		}
		log.Printf("Processed %d jobs, %d failed\n", len(jobs)-errors, errors)
	}
}

func getPoolObserverIDsForCoins(coins []int) []interface{} {
	result := make([]interface{}, 0)
	for _, po := range poolObservers {
		for _, c := range coins {
			if po.CoinID == c {
				result = append(result, po.ID)
				break
			}
		}
	}
	return result
}

func processJobCoinbaseData() (int, error) {

	log.Printf("Reading jobs")
	po := getPoolObserverIDsForCoins([]int{1}) // Only Bitcoin

	sql := `SELECT 
				j.id,
				j.pool_observer_id,
				j.generation_transaction_part_1,
				j.generation_transaction_part_2
			FROM 
				jobs j 
			WHERE 
				j.id > COALESCE((SELECT MAX(job_id) FROM job_coinbase_outputs), 0)	
				AND j.pool_observer_id = any($1)
			ORDER BY j.id`

	rows, err := db.Query(sql, pq.Array(po))
	if err != nil {
		return 0, err
	}
	jobs := make([]Job, 0)
	for rows.Next() {
		j := Job{}
		err = rows.Scan(&j.jobID, &j.poolObserverID, &j.genTX1, &j.genTX2)
		if err != nil {
			return 0, err
		}
		jobs = append(jobs, j)
		if len(jobs) >= 100000 {
			batchChan <- jobs
			jobs = make([]Job, 0)
		}

	}
	batchChan <- jobs
	log.Printf("Read jobs, waiting till processing is done\n")
	for len(batchChan) > 0 {
		time.Sleep(time.Second)
	}

	return 0, nil
}

func processJob(stmt *sql.Stmt, jobID int64, nonceSize int, genTX1, genTX2 []byte) error {
	genTx := createGenTX(nonceSize, genTX1, genTX2)
	tx := wire.NewMsgTx()
	err := tx.Deserialize(bytes.NewReader(genTx))
	if err != nil {
		possibleTxs := make([]*wire.MsgTx, 0)
		// Try to mitigate by using extranonce data ranging from nonceSize-4 - nonceSize+4
		for i := 1; i < nonceSize*2; i++ {
			genTx = createGenTX(i, genTX1, genTX2)
			tx2 := wire.NewMsgTx()
			err = tx2.Deserialize(bytes.NewReader(genTx))
			if err == nil && len(tx2.TxIn) > 0 && len(tx2.TxOut) > 0 {
				for _, txo := range tx2.TxOut {
					if txo.Value > 55 {
						// This can't be a coinbase output - sanity check for misaligned but valid parsing
						err = fmt.Errorf("Coinbase output has out of range value")
						break
					}
				}
				if err == nil {
					possibleTxs = append(possibleTxs, tx2)
				}
			}
		}

		if len(possibleTxs) == 0 {
			stmt.Exec(jobID, []byte{0x00}, 0, 0)
			return nil // just skip!
		} else if len(possibleTxs) > 1 {
			for _, tx2 := range possibleTxs {
				var buf bytes.Buffer
				tx2.Serialize(&buf)
				logging.Debugf("%x", buf.Bytes())
				buf.Reset()
			}
			panic("Two potential transactions found")
		} else {
			tx = possibleTxs[0]
		}

	}

	for i, txo := range tx.TxOut {
		script := make([]byte, len(txo.PkScript))
		copy(script, txo.PkScript)
		_, err := stmt.Exec(jobID, script, txo.Value, i)
		if err != nil {
			return err
		}
	}

	return nil
}

func createGenTX(nonceSize int, genTX1, genTX2 []byte) []byte {
	genTx := make([]byte, nonceSize+len(genTX1)+len(genTX2))
	copy(genTx[:], genTX1)
	copy(genTx[len(genTX1)+nonceSize:], genTX2)
	return genTx
}
