package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	dcrchainhash "github.com/decred/dcrd/chaincfg/chainhash"
	dcrrpcclient "github.com/decred/dcrd/rpcclient"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	rpcclient "github.com/mit-dci/pooldetective/blockfetcher/go-bitcoin-core-rpc"
	"github.com/mit-dci/pooldetective/logging"
	"github.com/mit-dci/pooldetective/util"
)

type NSDRPC struct {
	rpc *rpcclient.Client
}

var db *sql.DB

func bulk() {
	bulkHeaders := map[string]BlockHeader{}
	bulkBatchSize := 250000
	if os.Getenv("BULKBATCHSIZE") != "" {
		bulkBatchSize, _ = strconv.Atoi(os.Getenv("BULKBATCHSIZE"))
	}
	c := os.Getenv("BULK")
	cid, err := getCoinId(c)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}
	if cid == -1 {
		logging.Debugf("Adding coin %s\n", c)
		cid, err = addCoin(c)
		if err != nil {
			panic(err)
		}
	}

	logging.Debugf("Coin %s is ID %d\n", c, cid)

	rpc, err := coinInit(c, cid)
	if err != nil {
		panic(err)
	}
	defer shutdown(rpc)

	tips, err := getChainTips(rpc)
	if err != nil {
		panic(err)
	}

	fetchChan := make(chan *chainhash.Hash, 10000)

	best, err := getBestBlockHash(rpc)
	if err != nil {
		panic(err)
	}
	fetchChan <- best
	for _, t := range tips {
		fetchChan <- t
	}

	log.Printf("Querying already known blocks")

	rows, err := db.Query("SELECT block_hash FROM blocks WHERE coin_id=$1", cid)
	if err != nil {
		panic(err)
	}
	knownBlocks := map[string]bool{}
	var h []byte
	for rows.Next() {
		err = rows.Scan(&h)
		if err != nil {
			panic(err)
		}

		hash, _ := chainhash.NewHash(h)
		knownBlocks[hash.String()] = true
	}

	log.Printf("Found %d already known blocks", len(knownBlocks))
	log.Printf("Querying previous_block_hash to determine unconnected blocks")

	rows, err = db.Query("SELECT previous_block_hash FROM blocks WHERE coin_id=$1", cid)
	if err != nil {
		panic(err)
	}

	disconnectedBlocks := 0

	for rows.Next() {
		err = rows.Scan(&h)
		if err != nil {
			panic(err)
		}

		hash, _ := chainhash.NewHash(h)

		_, alreadyKnown := knownBlocks[hash.String()]
		if !alreadyKnown {
			fetchChan <- hash
			disconnectedBlocks++
		}
	}

	log.Printf("Found %d disconnected blocks, added to the fetchChan", disconnectedBlocks)

	lastUpdate := time.Now()
	for t := range fetchChan {

		header, err := getBlockHeader(rpc, t)
		if err != nil {
			logging.Errorf("Error getting blockheader: %v", err)
			if !strings.Contains(err.Error(), "Block not found") {
				time.Sleep(time.Second * 1)
				fetchChan <- t
			}
			continue
		}

		hash, _ := chainhash.NewHash(header.BlockHash)
		bulkHeaders[hash.String()] = header

		prevHash, _ := chainhash.NewHash(header.PreviousBlockHash)
		_, ok := bulkHeaders[prevHash.String()]
		if !ok && prevHash.String() != "0000000000000000000000000000000000000000000000000000000000000000" {
			_, alreadyKnown := knownBlocks[prevHash.String()]
			if !alreadyKnown {
				fetchChan <- prevHash
			}
		}
		if len(fetchChan) == 0 {
			break
		}
		if len(bulkHeaders)%bulkBatchSize == 0 {

			for b := range knownBlocks {
				_, ok := bulkHeaders[b]
				if ok {
					delete(bulkHeaders, b)
				}
			}

			if len(bulkHeaders) > 0 {
				logging.Debugf("Inserting %d headers in bulk", len(bulkHeaders))
				errors := bulkInsertHeaders(cid, bulkHeaders)
				logging.Debugf("Done! - %d errors", errors)

				for k := range bulkHeaders {
					knownBlocks[k] = true
				}

				bulkHeaders = map[string]BlockHeader{}
			}
		}

		if time.Now().Sub(lastUpdate).Seconds() > 5 {
			logging.Debugf("Read %d headers - fetchChan length: %d", len(bulkHeaders), len(fetchChan))
			lastUpdate = time.Now()
		}

	}

	for b := range knownBlocks {
		_, ok := bulkHeaders[b]
		if ok {
			delete(bulkHeaders, b)
		}
	}

	logging.Debugf("Final - Inserting %d headers in bulk", len(bulkHeaders))
	errors := bulkInsertHeaders(cid, bulkHeaders)
	logging.Debugf("Done! - %d errors", errors)
}

func bulkInsertHeaders(cid int, bulkHeaders map[string]BlockHeader) int {
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	stmt, err := tx.Prepare(pq.CopyIn("blocks", "coin_id", "block_hash", "previous_block_hash", "merkle_root", "height", "timestamp"))
	if err != nil {
		panic(err)
	}

	errors := 0
	for _, header := range bulkHeaders {
		_, err := stmt.Exec(cid, header.BlockHash, header.PreviousBlockHash, header.MerkleRoot, header.Height, header.Time)
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
	return errors
}

func main() {
	var err error
	logging.SetLogLevel(util.GetLoglevelFromEnv())

	connStr := os.Getenv("PGSQL_CONNECTION")
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	if os.Getenv("BULK") != "" {
		bulk()
		return
	}

	unconnectedBlocks, err := getUnconnectedBlocks()
	if err != nil {
		panic(err)
	}

	coins := strings.Split(os.Getenv("COINS"), "-")
	for _, c := range coins {
		cid, err := getCoinId(c)
		if err != nil && err != sql.ErrNoRows {
			panic(err)
		}

		if cid == -1 {
			logging.Debugf("Adding coin %s\n", c)
			cid, err = addCoin(c)
			if err != nil {
				panic(err)
			}
		}
		logging.Debugf("Coin %s is ID %d\n", c, cid)

		unconnectedBlocksForCoin, ok := unconnectedBlocks[cid]
		if !ok {
			unconnectedBlocksForCoin = make([]*chainhash.Hash, 0)
		}
		rpc, err := coinInit(c, cid)
		if err != nil {
			panic(err)
		}
		defer shutdown(rpc)

		startHeight := 0
		startHeightParam := os.Getenv(fmt.Sprintf("%s_COINBASE_STARTHEIGHT", c))
		if startHeightParam != "" {
			startHeight, _ = strconv.Atoi(startHeightParam)
		}

		go func(ticker string, coinId int, r interface{}) {
			tips, err := getChainTips(r)
			if err == nil {
				logging.Debugf("Adding %d hashes from getchaintips for coin %s", len(tips), ticker)
				unconnectedBlocksForCoin = append(unconnectedBlocksForCoin, tips...)
			} else {
				logging.Debugf("Error fetching chaintips: %s", err.Error())
			}
			logging.Debugf("Starting blockfetcher...")
			blockFetcher(coinId, r, unconnectedBlocksForCoin)
		}(c, cid, rpc)

		// For now, skip coinbase fetching for Decred. Sort it out later.
		btcrpc, ok := rpc.(*rpcclient.Client)
		if ok {
			go coinbaseFetcher(cid, startHeight, btcrpc)
		}
	}
	for {
		time.Sleep(time.Second * 10)
	}
}

func shutdown(rpc interface{}) {
	switch r := rpc.(type) {
	case *rpcclient.Client:
		r.Shutdown()
	case *dcrrpcclient.Client:
		r.Shutdown()
	}
}

func getUnconnectedBlocks() (map[int][]*chainhash.Hash, error) {
	return getUnconnectedBlocksForCoins([]interface{}{})
}

func getUnconnectedBlocksForCoins(coins []interface{}) (map[int][]*chainhash.Hash, error) {
	query := `select 
	b1.coin_id, b1.previous_block_hash 
	from 
	blocks b1 left join blocks b2 
	on 
	b2.block_hash=b1.previous_block_hash 
	and b2.coin_id=b1.coin_id 
	and b2.height is not null
	where 
	b2.block_hash is null 
	and b1.height > 0`

	if len(coins) == 1 {
		query += " and b1.coin_id = $1"
	} else if len(coins) > 1 {
		query += " and b1.coin_id in ("
		for i := range coins {
			if i > 0 {
				query += ","
			}
			query += fmt.Sprintf("$%d", i+1)
		}
		query += ")"
	}

	rows, err := db.Query(query, coins...)

	if err != nil {
		return nil, err
	}

	ret := map[int][]*chainhash.Hash{}
	for rows.Next() {
		var b []byte
		var coinID int
		err = rows.Scan(&coinID, &b)
		if err != nil {
			return nil, err
		}
		hash, err := chainhash.NewHash(b)
		if err != nil {
			return nil, err
		}

		existingArray, ok := ret[coinID]
		if !ok {
			existingArray = make([]*chainhash.Hash, 0)
		}

		existingArray = append(existingArray, hash)
		ret[coinID] = existingArray
	}

	return ret, nil

}

func coinInit(coin string, coinid int) (interface{}, error) {
	if coin == "DCR" {
		certs, err := ioutil.ReadFile(os.Getenv(fmt.Sprintf("%s_RPCCERT", coin)))
		if err != nil {
			log.Fatal(err)
		}
		connCfg := &dcrrpcclient.ConnConfig{
			Host:         os.Getenv(fmt.Sprintf("%s_RPCHOST", coin)),
			Endpoint:     "ws",
			User:         os.Getenv(fmt.Sprintf("%s_RPCUSER", coin)),
			Pass:         os.Getenv(fmt.Sprintf("%s_RPCPASS", coin)),
			Certificates: certs,
		}
		logging.Debugf("RPC Server: %s for coin %d", connCfg.Host, coinid)
		return dcrrpcclient.New(connCfg, nil)
	}
	if coin == "NSD" || coin == "EUNO" {
		connCfg := &rpcclient.ConnConfig{
			Host: os.Getenv(fmt.Sprintf("%s_RPCHOST", coin)),
			User: os.Getenv(fmt.Sprintf("%s_RPCUSER", coin)),
			Pass: os.Getenv(fmt.Sprintf("%s_RPCPASS", coin)),
		}
		logging.Debugf("RPC Server: %s for coin %d", connCfg.Host, coinid)
		rpc, err := rpcclient.New(connCfg)
		if err != nil {
			return nil, err
		}
		return &NSDRPC{rpc: rpc}, nil
	}
	connCfg := &rpcclient.ConnConfig{
		Host: os.Getenv(fmt.Sprintf("%s_RPCHOST", coin)),
		User: os.Getenv(fmt.Sprintf("%s_RPCUSER", coin)),
		Pass: os.Getenv(fmt.Sprintf("%s_RPCPASS", coin)),
	}
	logging.Debugf("RPC Server: %s for coin %d", connCfg.Host, coinid)
	return rpcclient.New(connCfg)

}

type GetChainTipsResult struct {
	Hash   string `json:"hash"`
	Status string `json:"status"`
}

func getChainTips(rpc interface{}) ([]*chainhash.Hash, error) {
	switch r := rpc.(type) {
	case *rpcclient.Client:
		res, err := r.RawRequest("getchaintips", []json.RawMessage{})
		if err != nil {
			return nil, err
		}

		j, _ := res.MarshalJSON()
		var chainTips []GetChainTipsResult
		err = json.Unmarshal(j, &chainTips)
		if err != nil {
			return nil, err
		}

		result := make([]*chainhash.Hash, 0)

		for _, ct := range chainTips {
			if ct.Status != "invalid" {
				h, err := chainhash.NewHashFromStr(ct.Hash)
				if err != nil {
					return nil, err
				}
				result = append(result, h)
			}
		}

		best, err := r.GetBestBlockHash()
		if err != nil {
			return nil, err
		}
		result = append(result, best)
		return result, nil
	case *dcrrpcclient.Client:
		res, err := r.GetChainTips()
		if err != nil {
			return nil, err
		}
		result := make([]*chainhash.Hash, 0)

		for _, ct := range res {
			if ct.Status != "invalid" {
				logging.Debugf("Appending chain tip %s at height %d\n", ct.Hash, ct.Height)

				h, err := chainhash.NewHashFromStr(ct.Hash)
				if err != nil {
					return nil, err
				}
				result = append(result, h)
			}
		}
		best, err := r.GetBestBlockHash()
		if err != nil {
			return nil, err
		}

		h, _ := chainhash.NewHashFromStr(best.String())
		result = append(result, h)
		return result, nil
	default:

		return make([]*chainhash.Hash, 0), nil
	}
}

func getBestBlockHash(rpc interface{}) (*chainhash.Hash, error) {
	switch r := rpc.(type) {
	case *NSDRPC:
		return r.rpc.GetBestBlockHash()
	case *rpcclient.Client:
		return r.GetBestBlockHash()
	case *dcrrpcclient.Client:
		h, err := r.GetBestBlockHash()
		if err != nil {
			return nil, err
		}
		return chainhash.NewHash(h.CloneBytes())
	}
	return nil, errors.New("Unrecognized RPC Client")
}

func blockFetcher(coinid int, rpc interface{}, unconnected []*chainhash.Hash) {
	logging.Debugf("Got %d unconnected blocks, so creating a channel of size %d", len(unconnected), len(unconnected)+100)
	archiveChannel := make(chan *chainhash.Hash, len(unconnected)+100)

	go archiveBlock(rpc, coinid, archiveChannel)

	time.Sleep(time.Second * 10)

	for _, u := range unconnected {
		archiveChannel <- u
	}
	previousBest, _ := chainhash.NewHash(make([]byte, 32))

	for {
		bestHash, err := getBestBlockHash(rpc)
		if err != nil {
			logging.Debugf("Error getting best block for coin %d : %v", coinid, err)
			time.Sleep(time.Second * 5)
			continue
		}
		if !previousBest.IsEqual(bestHash) {
			archiveChannel <- bestHash
			previousBest = bestHash
		}

		for len(archiveChannel) >= 50 {
			time.Sleep(time.Second * 1)
		}

	}
}

func getCoinId(ticker string) (int, error) {
	result := int(-1)
	err := db.QueryRow("select id from coins where ticker = $1", ticker).Scan(&result)
	return result, err
}

func addCoin(ticker string) (int, error) {
	result := int(-1)
	err := db.QueryRow("INSERT INTO coins (ticker, name) VALUES ($1, $2) RETURNING id", ticker, os.Getenv(fmt.Sprintf("%s_NAME", ticker))).Scan(&result)
	return result, err
}

func getBlockID(coinid int, blockHash []byte) (int64, int, error) {
	result := int64(-1)
	height := int(-1)

	err := db.QueryRow("select id, coalesce(height, -1) from blocks where coin_id=$1 and block_hash=$2", coinid, blockHash).Scan(&result, &height)
	return result, height, err
}

func upsertBlock(blockID int64, coinID int, blockHash, previousBlockHash, merkleRoot []byte, height int, timestamp int64) error {
	_, err := db.Exec(`
				  INSERT INTO blocks(
					  coin_id, block_hash, previous_block_hash, merkle_root, height, timestamp
				  ) VALUES (
					  $1,$2,$3,$4,$5,$6
				  )
				  ON CONFLICT ON CONSTRAINT blocks_coin_id_block_hash_key
				  DO
				  UPDATE SET previous_block_hash=$3, merkle_root=$4, height=$5, timestamp=$6`,
		coinID, blockHash, previousBlockHash, merkleRoot, height, timestamp)

	return err
}

type BlockHeader struct {
	BlockHash         []byte
	PreviousBlockHash []byte
	MerkleRoot        []byte
	Height            int
	Time              int64
}

func getBlockHeader(rpc interface{}, hash *chainhash.Hash) (BlockHeader, error) {
	returnValue := BlockHeader{BlockHash: hash.CloneBytes()}
	switch r := rpc.(type) {
	case *rpcclient.Client:
		header, err := r.GetBlockHeaderVerbose(hash)
		if err != nil {
			return returnValue, err
		}
		h, _ := chainhash.NewHashFromStr(header.PreviousHash)
		returnValue.PreviousBlockHash = h.CloneBytes()

		h, _ = chainhash.NewHashFromStr(header.MerkleRoot)
		returnValue.MerkleRoot = h.CloneBytes()

		returnValue.Height = int(header.Height)
		returnValue.Time = header.Time
	case *NSDRPC:
		b, err := r.rpc.GetBlockVerboseOld(hash)
		if err != nil {
			return returnValue, err
		}
		h, _ := chainhash.NewHashFromStr(b.PreviousHash)
		returnValue.PreviousBlockHash = h.CloneBytes()

		h, _ = chainhash.NewHashFromStr(b.MerkleRoot)
		returnValue.MerkleRoot = h.CloneBytes()

		returnValue.Height = int(b.Height)
		returnValue.Time = b.Time
	case *dcrrpcclient.Client:
		dcrhash, _ := dcrchainhash.NewHash(hash.CloneBytes())
		header, err := r.GetBlockHeaderVerbose(dcrhash)
		if err != nil {
			return returnValue, err
		}

		h, _ := chainhash.NewHashFromStr(header.PreviousHash)
		returnValue.PreviousBlockHash = h.CloneBytes()

		h, _ = chainhash.NewHashFromStr(header.MerkleRoot)
		returnValue.MerkleRoot = h.CloneBytes()

		returnValue.Height = int(header.Height)
		returnValue.Time = header.Time
	default:
		return BlockHeader{}, errors.New("Unrecognized RPC Client")
	}
	return returnValue, nil
}

func archiveBlock(rpc interface{}, coinID int, channel chan *chainhash.Hash) {
	for hash := range channel {
		logging.Debugf("Archiving block %s for coin %d (pending: %d)\n", hash.String(), coinID, len(channel))
		blockID, blockHeight, err := getBlockID(coinID, hash.CloneBytes())
		if err != nil && err != sql.ErrNoRows {
			panic(err)
		}

		if blockID == -1 || blockHeight == -1 {
			logging.Debugf("Block not present, getting header from RPC...\n")

			header, err := getBlockHeader(rpc, hash)
			if err != nil {
				logging.Errorf("Error getting blockheader: %v", err)
				if !strings.Contains(err.Error(), "Block not found") {
					time.Sleep(time.Second * 1)
					channel <- hash
				}
				continue
			}

			err = upsertBlock(blockID, coinID, header.BlockHash, header.PreviousBlockHash, header.MerkleRoot, header.Height, header.Time)
			if err != nil {
				panic(err)
			}

			prevHash, _ := chainhash.NewHash(header.PreviousBlockHash)
			if prevHash.String() != "0000000000000000000000000000000000000000000000000000000000000000" {
				channel <- prevHash
			}
		} else {
			logging.Debugf("Block %s already present!\n", hash.String())

		}
	}
}
