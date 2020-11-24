package main

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/lib/pq"
	rpcclient "github.com/mit-dci/pooldetective/blockfetcher/go-bitcoin-core-rpc"
	"github.com/mit-dci/pooldetective/logging"
	"github.com/mit-dci/pooldetective/util"
)

var pddb *sql.DB
var coinRPC map[int]interface{} = map[int]interface{}{}

func main() {
	logging.SetLogLevel(util.GetLoglevelFromEnv())

	var err error
	pddb, err = sql.Open("postgres", os.Getenv("PGSQL_CONNECTION"))
	if err != nil {
		panic(err)
	}

	rows, err := pddb.Query("select id, rpchost, rpcuser, rpcpass from coins where rpchost is not null")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var coinID int
		var rpchost string
		var rpcuser string
		var rpcpass string

		err = rows.Scan(&coinID, &rpchost, &rpcuser, &rpcpass)
		if err != nil {
			panic(err)
		}

		coinRPC[coinID], err = coinInit(rpchost, rpcuser, rpcpass)
		if err != nil {
			panic(err)
		}
	}
	rows.Close()

	rows, err = pddb.Query("select r.id, r.coin_id, fork_block_hash, added_blocks, removed_blocks FROM reorgs r WHERE r.blocksmissing=true")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var coinID int
		var reorgID int
		var forkBlockHash []byte
		var addedBlocks pq.ByteaArray
		var removedBlocks pq.ByteaArray

		err = rows.Scan(&reorgID, &coinID, &forkBlockHash, &addedBlocks, &removedBlocks)
		if err != nil {
			panic(err)
		}

		rpc, ok := coinRPC[coinID]
		if ok {
			fmt.Printf("Importing blocks for reorg %d (Coin %d)\n", reorgID, coinID)
			blocks := append(addedBlocks, removedBlocks...)
			blocks = append(blocks, forkBlockHash)

			fetchBlocks(reorgID, coinID, rpc, blocks)
		} else {
			fmt.Printf("Can't download blocks for %d: No RPC!\n", coinID)
		}
	}
	rows.Close()
}

func fetchBlocks(reorgID, coinID int, rpc interface{}, blocks [][]byte) {
	for _, b := range blocks {
		if len(b) == 32 {
			h, err := chainhash.NewHash(b)
			if err != nil {
				logging.Warnf("Could not fetch (all) blocks for reorg %d coin %d: %s\n", reorgID, coinID, err.Error())
				continue
			}

			err = ensureBlockPresent(rpc, h)
			if err != nil {
				logging.Warnf("Could not fetch (all) blocks for reorg %d coin %d: %s\n", reorgID, coinID, err.Error())
			}
		}
	}
}

func getRawBlock(rpc interface{}, h *chainhash.Hash) ([]byte, error) {
	switch r := rpc.(type) {
	case *rpcclient.Client:
		resp, err := r.RawRequest("getblock", []json.RawMessage{json.RawMessage(fmt.Sprintf("\"%s\"", h.String())), json.RawMessage("false")})
		if err != nil {
			return nil, err
		}
		var hexBlock string
		json.Unmarshal(resp, &hexBlock)
		blk, err := hex.DecodeString(hexBlock)
		if err != nil {
			return nil, err
		}
		return blk, nil
		/*case *NSDRPC:
			resp, err := r.rpc.RawRequest("getblock", []json.RawMessage{json.RawMessage(fmt.Sprintf("\"%s\"", h.String())), json.RawMessage("false")})
			if err != nil {
				return nil, err
			}
			var hexBlock string
			json.Unmarshal(resp, &hexBlock)
			blk, err := hex.DecodeString(hexBlock)
			if err != nil {
				return nil, err
			}
			return blk, nil
		case *dcrrpcclient.Client:
			resp, err := r.RawRequest("getblock", []json.RawMessage{json.RawMessage(fmt.Sprintf("\"%s\"", h.String())), json.RawMessage("false")})
			if err != nil {
				return nil, err
			}
			var hexBlock string
			json.Unmarshal(resp, &hexBlock)
			blk, err := hex.DecodeString(hexBlock)
			if err != nil {
				return nil, err
			}
			return blk, nil*/
	}
	return nil, errors.New("Unrecognized RPC Client")
}

func ensureBlockPresent(rpc interface{}, h *chainhash.Hash) error {
	blockFile := fmt.Sprintf("%s/%x.blk", os.Getenv("BLOCKSDIR"), h.CloneBytes())
	if _, err := os.Stat(blockFile); os.IsNotExist(err) {

		b, err := getRawBlock(rpc, h)
		if err != nil {
			return err
		}

		return ioutil.WriteFile(blockFile, b, 0644)
	}
	return nil
}

func coinInit(host, user, pass string) (interface{}, error) {
	connCfg := &rpcclient.ConnConfig{
		Host: host,
		User: user,
		Pass: pass,
	}
	return rpcclient.New(connCfg)
}
