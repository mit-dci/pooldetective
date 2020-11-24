package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/ethereum/go-ethereum/common"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

type Reorg struct {
	ID            int
	Occurred      time.Time
	AddedBlocks   [][]byte
	RemovedBlocks [][]byte
	Currency      string
}

type JsonBlock struct {
	BlockID         string `json:"block_id"`
	PreviousBlockID string `json:"previous_block_id"`
}

func main() {
	rtdb, err := sql.Open("mysql", os.Getenv("MYSQL_CONNECTION"))
	if err != nil {
		panic(err)
	}
	// See "Important settings" section.
	rtdb.SetConnMaxLifetime(time.Minute * 3)
	rtdb.SetMaxOpenConns(10)
	rtdb.SetMaxIdleConns(10)

	pddb, err := sql.Open("postgres", os.Getenv("PGSQL_CONNECTION"))
	if err != nil {
		panic(err)
	}

	poolDetectiveCoins := map[string]int{}

	rows, err := pddb.Query("SELECT TRIM(ticker), id FROM coins")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var ID int
		var ticker string
		err := rows.Scan(&ticker, &ID)
		if err != nil {
			panic(err)
		}

		poolDetectiveCoins[ticker] = ID
	}
	rows.Close()

	currencies := map[int]string{}
	rows, err = rtdb.Query("SELECT id, name FROM cryptocurrency")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var ID int
		var ticker string
		err := rows.Scan(&ID, &ticker)
		if err != nil {
			panic(err)
		}

		currencies[ID] = ticker
	}
	rows.Close()

	reorgs := []Reorg{}
	rows, err = rtdb.Query("SELECT id, currency_id FROM reorg WHERE id > 10")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var ID int
		var curID int
		err := rows.Scan(&ID, &curID)
		if err != nil {
			panic(err)
		}
		reorgs = append(reorgs, Reorg{ID: ID, Currency: currencies[curID], AddedBlocks: [][]byte{}, RemovedBlocks: [][]byte{}})
	}
	rows.Close()

	for i := range reorgs {

		fmt.Printf("Querying fork block for reorg %d", reorgs[i].ID)

		pdcoin, ok := poolDetectiveCoins[reorgs[i].Currency]
		if !ok {
			continue
		}
		rows, err = rtdb.Query("SELECT block_json FROM reorg r left join block b on b.id=r.fork_block_id where r.id=?", reorgs[i].ID)
		if err != nil {
			panic(err)
		}

		forkBlockHash := []byte{}

		if rows.Next() {
			var blockJSON string
			err := rows.Scan(&blockJSON)
			if err != nil {
				panic(err)
			}
			var block JsonBlock
			json.Unmarshal([]byte(blockJSON), &block)
			forkBlockHash = parseBlockHash(block.BlockID)
		} else {
			panic("Fork block not found")
		}
		rows.Close()

		fmt.Printf("Querying added blocks for reorg %d", reorgs[i].ID)

		addedBlocks := []JsonBlock{}

		rows, err = rtdb.Query("SELECT receipt_time, block_json FROM reorg_added_blocks rb left join block b on b.id=rb.block_id where rb.reorg_id=?", reorgs[i].ID)
		if err != nil {
			panic(err)
		}
		maxReceiptTime := time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
		for rows.Next() {

			var receipt time.Time
			var blockJSON string
			err := rows.Scan(&receipt, &blockJSON)
			if err != nil {
				panic(err)
			}

			if receipt.After(maxReceiptTime) {
				maxReceiptTime = receipt
			}

			var block JsonBlock
			json.Unmarshal([]byte(blockJSON), &block)

			addedBlocks = append(addedBlocks, block)
			fmt.Printf("Queried %d added blocks", len(addedBlocks))
		}
		rows.Close()
		reorgs[i].Occurred = maxReceiptTime
		reorgs[i].AddedBlocks = sortBlocks(forkBlockHash, addedBlocks)

		removedBlocks := []JsonBlock{}

		rows, err = rtdb.Query("SELECT block_json FROM reorg_removed_blocks rb left join block b on b.id=rb.block_id where rb.reorg_id=?", reorgs[i].ID)
		if err != nil {
			panic(err)
		}
		for rows.Next() {
			var blockJSON string
			err := rows.Scan(&blockJSON)
			if err != nil {
				panic(err)
			}

			var block JsonBlock
			json.Unmarshal([]byte(blockJSON), &block)

			removedBlocks = append(removedBlocks, block)
		}
		rows.Close()
		reorgs[i].RemovedBlocks = sortBlocks(forkBlockHash, removedBlocks)

		fmt.Printf("Queried reorg:\r\nID: %d\r\nCoin: %s (PD ID: %d)\r\nObserved: %s\r\nRemoved blocks:\r\n", reorgs[i].ID, reorgs[i].Currency, pdcoin, reorgs[i].Occurred)
		for _, b := range reorgs[i].RemovedBlocks {
			h, _ := chainhash.NewHash(b)
			fmt.Printf("ETH-like: %s - BTC-like: %s\r\n", common.ToHex(b), h.String())
		}
		fmt.Printf("Added blocks:\r\n")
		for _, b := range reorgs[i].AddedBlocks {
			h, _ := chainhash.NewHash(b)
			fmt.Printf("ETH-like: %s - BTC-like: %s\r\n", common.ToHex(b), h.String())
		}
		if i > 100 {
			break
		}
	}

}

func parseBlockHash(h string) []byte {
	if strings.HasPrefix(h, "0x") {
		return common.FromHex(h)
	}

	bh, _ := chainhash.NewHashFromStr(h)
	return bh.CloneBytes()
}

func sortBlocks(lastHash []byte, blocks []JsonBlock) [][]byte {
	resultBlocks := [][]byte{}
	for len(blocks) > 0 {
		fmt.Printf("Sorting %d blocks\n", len(blocks))
		removeBlock := -1
		for ib, b := range blocks {
			previousBlockHash := parseBlockHash(b.PreviousBlockID)
			if bytes.Equal(previousBlockHash, lastHash) {
				lastHash = parseBlockHash(b.BlockID)
				resultBlocks = append(resultBlocks, lastHash)
				removeBlock = ib
				break
			}
		}

		if removeBlock == -1 {
			break
		}

		blocks[removeBlock] = blocks[len(blocks)-1]
		blocks[len(blocks)-1] = JsonBlock{}
		blocks = blocks[:len(blocks)-1]
	}
	if len(blocks) > 0 {
		panic("Could not sort blocks!")
	}

	return resultBlocks
}
