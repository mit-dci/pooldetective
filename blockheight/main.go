package main

import (
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
	"github.com/mit-dci/pooldetective/logging"
	"github.com/mit-dci/pooldetective/util"
)

var db *sql.DB

func main() {
	var err error
	logging.SetLogLevel(util.GetLoglevelFromEnv())

	connStr := os.Getenv("PGSQL_CONNECTION")
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	rows, err := db.Query(`SELECT j.id, encode(generation_transaction_part_1, 'hex') || repeat('00',length(po.stratum_extranonce1)+po.stratum_extranonce2size) || encode(generation_transaction_part_2, 'hex') FROM jobs j LEFT JOIN pool_observers po on po.id=j.pool_observer_id LEFT JOIN pools p on p.id=po.pool_id LEFT JOIN coins c on c.id=po.coin_id WHERE previous_block_id =-1 AND j.observed < now() - interval '30 minutes' AND length(generation_transaction_part_1) > 0 AND j.probable_blockheight IS NULL ORDER BY id desc`)
	if err != nil {
		panic(err)
	}
	i := 0
	for rows.Next() {
		var jobID int64
		var coinbaseTxHex sql.NullString
		err = rows.Scan(&jobID, &coinbaseTxHex)
		if err != nil {
			panic(err)
		}

		bh := -1
		if coinbaseTxHex.Valid {
			ffidx := strings.Index(coinbaseTxHex.String, "ffffffff")
			//fmt.Printf("ffidx for job %d is %d\n", jobID, ffidx)
			coinbaseScript := coinbaseTxHex.String[ffidx+10:]
			//fmt.Printf("coinbaseScript for job %d is %s\n", jobID, coinbaseScript)
			blockHeightLength, _ := strconv.Atoi(coinbaseScript[0:2])
			//fmt.Printf("blockHeightLength for job %d is %d\n", jobID, blockHeightLength)
			blockHeightBytes, _ := hex.DecodeString(coinbaseScript[2 : 2+(blockHeightLength*2)])
			//fmt.Printf("blockHeightBytes for job %d is %x\n", jobID, blockHeightBytes)
			blockHeightLE := make([]byte, 4)
			copy(blockHeightLE, blockHeightBytes)
			//fmt.Printf("blockHeightLE for job %d is %x\n", jobID, blockHeightLE)
			bh = int(binary.LittleEndian.Uint32(blockHeightLE))
			//fmt.Printf("Block height for job %d is %d\n", jobID, bh)
		}
		db.Exec("UPDATE jobs SET probable_blockheight=$1 where id=$2", bh, jobID)
		i++
		if i%100 == 0 {
			fmt.Printf("\rUpdated %d jobs", i)
		}
	}

}
