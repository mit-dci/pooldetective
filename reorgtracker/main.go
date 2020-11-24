package main

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	dcrchainhash "github.com/decred/dcrd/chaincfg/chainhash"
	dcrrpcclient "github.com/decred/dcrd/rpcclient"
	"github.com/lib/pq"
	rpcclient "github.com/mit-dci/pooldetective/blockfetcher/go-bitcoin-core-rpc"
	"github.com/mit-dci/pooldetective/logging"
	"github.com/mit-dci/pooldetective/util"
)

type BlockHeader struct {
	BlockHash         []byte
	PreviousBlockHash []byte
	MerkleRoot        []byte
	Height            int
	Time              int64
	ChainWork         []byte
}

type NSDRPC struct {
	rpc *rpcclient.Client
}

var db *sql.DB
var coinID int
var baseHeight int
var currentChain []*BlockHeader
var maxDepth int

func main() {
	logging.SetLogLevel(util.GetLoglevelFromEnv())

	var err error
	db, err = sql.Open("postgres", os.Getenv("PGSQL_CONNECTION"))
	if err != nil {
		logging.Fatal(err)
	}

	c := os.Getenv("COIN")
	coinID, err = getCoinId(c)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}
	if coinID == -1 {
		logging.Debugf("Adding coin %s\n", c)
		coinID, err = addCoin(c)
		if err != nil {
			panic(err)
		}
	}

	rpc, err := coinInit(c)
	if err != nil {
		panic(err)
	}

	maxDepth, _ = strconv.Atoi(os.Getenv("MAXDEPTH"))
	if maxDepth == 0 {
		maxDepth = 100
	}

	logging.Debugf("Reorg tracker for coin %s starting, max depth %d", c, maxDepth)

	os.MkdirAll(fmt.Sprintf("%s/reorgblocks", os.Getenv("BLOCKSDIR")), 0755)
	ensureQ := make(chan *chainhash.Hash, 100)
	removeQ := make(chan *chainhash.Hash, 100)
	go ensureBlocksPresent(rpc, ensureQ)
	go removeBlocks(removeQ)

	currentChain = []*BlockHeader{}
	best, _ := chainhash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	for {
		h, err := getBestBlockHash(rpc)
		if err != nil {
			panic(err)
		}

		if h.IsEqual(best) {
			time.Sleep(time.Millisecond * 250)
			continue
		}

		writeBestHash(coinID, h)

		best = h

		bh, err := getBlockHeader(rpc, h)
		if err != nil {
			panic(err)
		}

		start, newBaseHeight, chain, err := getAttachableChain(rpc, bh, ensureQ)
		if err != nil {
			panic(err)
		}
		if start < len(currentChain) {
			observed := time.Now()
			removedBlocks := make([]*BlockHeader, len(currentChain)-start)
			copy(removedBlocks, currentChain[start:])

			currentChain = append(currentChain[:start], chain...)
			forkBlock := currentChain[start-1]
			addedBlocks := make([]*BlockHeader, len(currentChain)-start)
			copy(addedBlocks, currentChain[start:])

			reorgID, err := writeReorg(observed, forkBlock, addedBlocks, removedBlocks)
			if err != nil {
				logging.Errorf("Error registering reorg: %s", err.Error())
			} else {
				logging.Debugf("Registered reorg under ID %d", reorgID)
			}

		} else if start == 0 {
			currentChain = chain
			logging.Debugf("Initial chain loaded (tip: %x)\n", currentChain[len(currentChain)-1].BlockHash)
		} else if start == len(currentChain) {
			currentChain = append(currentChain, chain...)
			logging.Debugf("New blocks appended:\n")
			for i := start; i < len(currentChain); i++ {
				logging.Debugf("%x (Prev: %x)\n", currentChain[i].BlockHash, currentChain[i].PreviousBlockHash)
			}
			if len(currentChain) > maxDepth {
				prune := len(currentChain) - maxDepth
				for _, b := range currentChain[:prune] {
					h, _ := chainhash.NewHash(b.BlockHash)
					removeQ <- h
				}
				currentChain = currentChain[prune:]
				logging.Debugf("%d blocks pruned, currentChain length now %d\n", prune, len(currentChain))
			}

		}

		baseHeight = newBaseHeight
	}
}

func copyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func archiveReorgBlocks(hashes [][]byte) {
	for _, h := range hashes {
		blockFile := fmt.Sprintf("%s/%x.blk", os.Getenv("BLOCKSDIR"), h)
		archiveFile := fmt.Sprintf("%s/reorgblocks/%x.blk", os.Getenv("BLOCKSDIR"), h)
		_, err := copyFile(blockFile, archiveFile)
		if err != nil {
			logging.Errorf("Could not archive reorg block %x: %s\n", h, err.Error())
		} else {
			logging.Debugf("Archived reorg block %x\n", h)
		}
	}
}

func writeReorg(observed time.Time, forkBlock *BlockHeader, addedBlocks []*BlockHeader, removedBlocks []*BlockHeader) (int, error) {

	addedBlockHashes := make([][]byte, len(addedBlocks))
	removedBlockHashes := make([][]byte, len(removedBlocks))

	for i, bh := range addedBlocks {
		addedBlockHashes[i] = bh.BlockHash
	}

	for i, bh := range removedBlocks {
		removedBlockHashes[i] = bh.BlockHash
	}

	archiveBlocks := append(addedBlockHashes, removedBlockHashes...)
	// Add common ancestor
	archiveBlocks = append(archiveBlocks, addedBlocks[0].PreviousBlockHash)
	archiveReorgBlocks(archiveBlocks)

	addedBlockHashesPQ, err := pq.ByteaArray(addedBlockHashes).Value()
	if err != nil {
		return -1, err
	}
	removedBlockHashesPQ, err := pq.ByteaArray(removedBlockHashes).Value()
	if err != nil {
		return -1, err
	}

	var reorgID int
	err = db.QueryRow(`insert into reorgs(coin_id, observed, fork_block_hash, fork_block_height, removed_blocks, added_blocks, fork_total_chainwork, removed_total_chainwork, added_total_chainwork)
						values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`,
		coinID,
		observed,
		forkBlock.BlockHash,
		forkBlock.Height,
		removedBlockHashesPQ,
		addedBlockHashesPQ,
		forkBlock.ChainWork,
		removedBlocks[len(removedBlocks)-1].ChainWork,
		addedBlocks[len(addedBlocks)-1].ChainWork,
	).Scan(&reorgID)
	return reorgID, err
}

func writeBestHash(coinID int, h *chainhash.Hash) error {
	r, err := db.Exec(`update coins set besthash=$1, besthashobserved=now() where id=$2`,
		h.CloneBytes(),
		coinID,
	)
	r.Close()
	return err
}

func ensureBlockPresent(rpc interface{}, h *chainhash.Hash) error {
	blockFile := fmt.Sprintf("%s/%x.blk", os.Getenv("BLOCKSDIR"), h.CloneBytes())
	if _, err := os.Stat(blockFile); os.IsNotExist(err) {

		b, err := getRawBlock(rpc, h)
		if err != nil {
			return err
		}

		return ioutil.WriteFile(blockFile, b, 644)
	}
	return nil
}

func removeBlock(h *chainhash.Hash) {
	logging.Debugf("Pruning block %x from blocksdir\n", h.CloneBytes())
	blockFile := fmt.Sprintf("%s/%x.blk", os.Getenv("BLOCKSDIR"), h.CloneBytes())
	os.Remove(blockFile)
}

func findBlockInCurrentChain(h *BlockHeader) int {
	for i, bh := range currentChain {
		if bytes.Equal(bh.BlockHash, h.BlockHash) {
			return i
		}
	}
	return -1
}

func ensureBlocksPresent(rpc interface{}, q chan *chainhash.Hash) {
	for h := range q {
		err := ensureBlockPresent(rpc, h)
		if err != nil {
			logging.Debugf("[ERR] Could not archive block %s: %s\n", h.String(), err)
			q <- h
			time.Sleep(time.Second * 1)
		}
	}
}

func removeBlocks(q chan *chainhash.Hash) {
	for h := range q {
		removeBlock(h)
	}
}

func getAttachableChain(rpc interface{}, h *BlockHeader, ensureQ chan *chainhash.Hash) (startIndex int, newBaseHeight int, chain []*BlockHeader, err error) {
	bh, _ := chainhash.NewHash(h.BlockHash)
	ensureQ <- bh
	chain = []*BlockHeader{}
	for {
		idx := findBlockInCurrentChain(h)
		if idx != -1 {
			startIndex = idx + 1
			newBaseHeight = baseHeight
			err = nil
			return
		}
		chain = append([]*BlockHeader{h}, chain...)

		pbh, _ := chainhash.NewHash(h.PreviousBlockHash)
		h, err = getBlockHeader(rpc, pbh)
		if err != nil {
			panic(err)
		}

		ensureQ <- pbh

		if len(chain) >= maxDepth {
			startIndex = 0
			newBaseHeight = h.Height
			err = nil
			return
		}

		if len(chain)%100 == 0 {
			logging.Debugf("Finding attachable chain [%d]\n", len(chain))
		}
	}
}

func coinInit(coin string) (interface{}, error) {
	if coin == "DCR" {
		certs, err := ioutil.ReadFile(os.Getenv("RPCCERT"))
		if err != nil {
			log.Fatal(err)
		}
		connCfg := &dcrrpcclient.ConnConfig{
			Host:         os.Getenv("RPCHOST"),
			User:         os.Getenv("RPCUSER"),
			Pass:         os.Getenv("RPCPASS"),
			Endpoint:     "ws",
			Certificates: certs,
		}
		logging.Debugf("RPC Server: %s", connCfg.Host)
		return dcrrpcclient.New(connCfg, nil)
	}
	if coin == "NSD" || coin == "EUNO" {
		connCfg := &rpcclient.ConnConfig{
			Host: os.Getenv("RPCHOST"),
			User: os.Getenv("RPCUSER"),
			Pass: os.Getenv("RPCPASS"),
		}
		logging.Debugf("RPC Server: %s", connCfg.Host)
		rpc, err := rpcclient.New(connCfg)
		if err != nil {
			return nil, err
		}
		return &NSDRPC{rpc: rpc}, nil
	}

	connCfg := &rpcclient.ConnConfig{
		Host: os.Getenv("RPCHOST"),
		User: os.Getenv("RPCUSER"),
		Pass: os.Getenv("RPCPASS"),
	}
	logging.Debugf("RPC Server: %s", connCfg.Host)
	return rpcclient.New(connCfg)
}

func getBestBlockHash(rpc interface{}) (*chainhash.Hash, error) {
	switch r := rpc.(type) {
	case *NSDRPC:
		return r.rpc.GetBestBlockHash()
	case *dcrrpcclient.Client:
		h, err := r.GetBestBlockHash()
		if err != nil {
			return nil, err
		}
		return chainhash.NewHash(h.CloneBytes())
	case *rpcclient.Client:
		return r.GetBestBlockHash()
	}
	return nil, errors.New("Unrecognized RPC Client")
}

func getBlockHeader(rpc interface{}, h *chainhash.Hash) (*BlockHeader, error) {
	returnValue := BlockHeader{}
	switch r := rpc.(type) {
	case *rpcclient.Client:
		bhv, err := r.GetBlockHeaderVerbose(h)
		if err != nil {
			return nil, err
		}

		blkhsh, _ := chainhash.NewHashFromStr(bhv.Hash)
		returnValue.BlockHash = blkhsh.CloneBytes()
		mr, _ := chainhash.NewHashFromStr(bhv.MerkleRoot)
		returnValue.MerkleRoot = mr.CloneBytes()
		pbhsh, _ := chainhash.NewHashFromStr(bhv.PreviousHash)
		returnValue.PreviousBlockHash = pbhsh.CloneBytes()
		returnValue.Time = bhv.Time
		returnValue.Height = int(bhv.Height)
		returnValue.ChainWork = []byte{0x00}
		returnValue.ChainWork, _ = hex.DecodeString(bhv.ChainWork)
		return &returnValue, nil
	case *NSDRPC:
		b, err := r.rpc.GetBlockVerboseOld(h)
		if err != nil {
			return nil, err
		}
		h, _ := chainhash.NewHashFromStr(b.Hash)
		returnValue.BlockHash = h.CloneBytes()

		h, _ = chainhash.NewHashFromStr(b.PreviousHash)
		returnValue.PreviousBlockHash = h.CloneBytes()

		h, _ = chainhash.NewHashFromStr(b.MerkleRoot)
		returnValue.MerkleRoot = h.CloneBytes()

		returnValue.Height = int(b.Height)
		returnValue.Time = b.Time
		returnValue.ChainWork = []byte{0x00}
		returnValue.ChainWork, _ = hex.DecodeString(b.ChainWork)
		return &returnValue, nil
	case *dcrrpcclient.Client:
		dcrhash, _ := dcrchainhash.NewHash(h.CloneBytes())
		header, err := r.GetBlockHeaderVerbose(dcrhash)
		if err != nil {
			return nil, err
		}

		h, _ := chainhash.NewHashFromStr(header.Hash)
		returnValue.BlockHash = h.CloneBytes()

		h, _ = chainhash.NewHashFromStr(header.PreviousHash)
		returnValue.PreviousBlockHash = h.CloneBytes()

		h, _ = chainhash.NewHashFromStr(header.MerkleRoot)
		returnValue.MerkleRoot = h.CloneBytes()

		returnValue.Height = int(header.Height)
		returnValue.Time = header.Time
		returnValue.ChainWork = []byte{0x00}
		returnValue.ChainWork, _ = hex.DecodeString(header.ChainWork)
		return &returnValue, nil
	}
	return nil, errors.New("Unrecognized RPC Client")
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
	case *NSDRPC:
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
		return blk, nil
	}
	return nil, errors.New("Unrecognized RPC Client")
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
