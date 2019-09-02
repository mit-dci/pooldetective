package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	dashwire "github.com/dashpay/godash/wire"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	rpcclient "github.com/mit-dci/pooldetective/blockfetcher/go-bitcoin-core-rpc"
	"github.com/mit-dci/pooldetective/blockobserver/wire"
	"github.com/mit-dci/pooldetective/logging"
)

type databaseBlock struct {
	id        int64
	blockHash *chainhash.Hash
}

func fillBlocksWithoutCoinbaseData(coinID, startHeight int, channel chan databaseBlock) error {
	rows, err := db.Query(`select 
	                       b.id, b.block_hash
						   from 
						   blocks b 
						   where 
						   b.coin_id=$1 
						   and b.height > $2
						   AND b.coinbase_data IS NULL
						   order by b.height
						   limit 50`,
		coinID,
		startHeight)
	if err != nil {
		return err
	}

	for rows.Next() {
		var id int64
		var b []byte
		err = rows.Scan(&id, &b)
		if err != nil {
			return err
		}
		hash, err := chainhash.NewHash(b)
		if err != nil {
			return err
		}
		channel <- databaseBlock{id: id, blockHash: hash}
	}

	return nil

}

func coinbaseFetcher(coinID, startHeight int, rpc *rpcclient.Client) {
	channel := make(chan databaseBlock, 100)
	go archiveCoinbaseBlocks(coinID, rpc, channel)
	for {
		if len(channel) == 0 {
			err := fillBlocksWithoutCoinbaseData(coinID, startHeight, channel)
			if err != nil {
				log.Fatal(err)
			}
		}
		time.Sleep(time.Second * 10)
	}
}

func insertCoinbaseData(blockID int64, merkleBranches [][]byte, outputScripts [][]byte, outputValues []int64) error {
	byteaarr, err := pq.ByteaArray(merkleBranches).Value()
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO block_coinbase_merkleproofs(block_id, coinbase_merklebranches) VALUES ($1,$2)`, blockID, byteaarr)
	if err != nil {
		return err
	}

	for i := range outputScripts {
		_, err := tx.Exec(`INSERT INTO block_coinbase_outputs(block_id, output_index, output_script, value) VALUES ($1,$2,$3,$4)`, blockID, i, outputScripts[i], outputValues[i])
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(`UPDATE blocks SET coinbase_data=true WHERE id=$1`, blockID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func archiveCoinbaseBlocks(coinID int, rpc *rpcclient.Client, channel chan databaseBlock) {
	for dbBlock := range channel {
		hash := dbBlock.blockHash
		block, err := rpc.GetBlockVerbose(hash)
		if err != nil {
			logging.Errorf("Could not fetch Block %x - Coin %d : %v", hash.CloneBytes(), coinID, err)
			//insert dummy
			insertCoinbaseData(dbBlock.id, [][]byte{}, [][]byte{}, []int64{})
			continue
		}

		log.Printf("Archiving coinbase data for coin %d - block %d", coinID, block.Height)

		txHashes := make([]*chainhash.Hash, len(block.Tx))
		for i, txHash := range block.Tx {
			txHashes[i], _ = chainhash.NewHashFromStr(txHash)
		}

		merkleRoot, err := chainhash.NewHashFromStr(block.MerkleRoot)
		if err != nil {
			log.Fatal(err)
		}

		merkles := BuildMerkleTreeStore(txHashes)
		merkleProof := NewMerkleProof(merkles, 0)
		merkleProofOK := CheckMerkleProof(merkleProof, txHashes[0], merkleRoot, 0)
		if !merkleProofOK {
			log.Fatal("Merkle proof was not OK...")
		}

		merkleProofBytes := make([][]byte, len(merkleProof))
		for i, m := range merkleProof {
			merkleProofBytes[i] = m.CloneBytes()
		}

		blockBytesHex, err := rpc.RawRequest("getblock", []json.RawMessage{[]byte(fmt.Sprintf("\"%s\"", hash.String())), []byte("false")})
		if err != nil {
			log.Fatalf("could not make raw getblock request: %s", err.Error())
		}

		blockBytes, err := hex.DecodeString(string(blockBytesHex[1 : len(blockBytesHex)-1]))
		if err != nil {
			log.Fatalf("could not decode hex from getblock request: %s", err.Error())
		}
		var outputScripts [][]byte
		var outputValues []int64
		if coinID == 8 {
			outputScripts, outputValues = GetOutputScriptsFromDashBlockBytes(blockBytes)
		} else {
			outputScripts, outputValues = GetOutputScriptsFromBlockBytes(blockBytes)
		}

		err = insertCoinbaseData(dbBlock.id, merkleProofBytes, outputScripts, outputValues)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func GetOutputScriptsFromBlockBytes(blockBytes []byte) (outputScripts [][]byte, outputValues []int64) {
	var rawBlock wire.MsgBlock

	err := rawBlock.Deserialize(bytes.NewBuffer(blockBytes))
	if err != nil {
		log.Printf("could not decode block from bytes: %s", err.Error())
		err = rawBlock.DeserializeNoWitness(bytes.NewBuffer(blockBytes))
		if err != nil {
			log.Printf("could not decode block without witness from bytes: %s -  can't recover", err.Error())
			outputScripts = make([][]byte, 0)
			outputValues = make([]int64, 0)
			return
		}
	}
	if len(rawBlock.Transactions) == 0 {
		outputScripts = make([][]byte, 0)
		outputValues = make([]int64, 0)
		return
	}
	coinbaseTx := rawBlock.Transactions[0]
	outputScripts = make([][]byte, len(coinbaseTx.TxOut))
	outputValues = make([]int64, len(coinbaseTx.TxOut))

	for i := range coinbaseTx.TxOut {
		outputScripts[i] = coinbaseTx.TxOut[i].PkScript
		outputValues[i] = coinbaseTx.TxOut[i].Value
	}

	return
}

func GetOutputScriptsFromDashBlockBytes(blockBytes []byte) (outputScripts [][]byte, outputValues []int64) {
	var rawBlock dashwire.MsgBlock

	err := rawBlock.Deserialize(bytes.NewBuffer(blockBytes))
	if err != nil {
		log.Fatalf("could not decode dash block from bytes: %s -  can't recover", err.Error())
	}
	coinbaseTx := rawBlock.Transactions[0]
	outputScripts = make([][]byte, len(coinbaseTx.TxOut))
	outputValues = make([]int64, len(coinbaseTx.TxOut))

	for i := range coinbaseTx.TxOut {
		outputScripts[i] = coinbaseTx.TxOut[i].PkScript
		outputValues[i] = coinbaseTx.TxOut[i].Value
	}

	return
}

func NewMerkleProof(merkleTree []*chainhash.Hash, idx uint64) []*chainhash.Hash {
	treeHeight := calcTreeHeight(uint64((len(merkleTree) + 1) / 2))

	proof := make([]*chainhash.Hash, treeHeight)
	for i := uint(0); i < treeHeight; i++ {
		if merkleTree[idx^1] == nil {
			// From the documentation of BuildMerkleTreeStore: "parent nodes
			// "with only a single left node are calculated by concatenating
			// the left node with itself before hashing."
			proof[i] = merkleTree[idx] // add "ourselves"
		} else {
			proof[i] = merkleTree[idx^1]
		}

		idx = (idx >> 1) | (1 << treeHeight)
	}
	return proof
}

// Check will validate a merkle proof given the hash of the element to prove (hash)
// and the expected root hash (expectedRoot). Will return true when the merkle proof
// is valid, false otherwise.
func CheckMerkleProof(proof []*chainhash.Hash, hash, expectedRoot *chainhash.Hash, hashIdx int) bool {
	treeHeight := uint(len(proof))

	for _, h := range proof {
		var newHash chainhash.Hash
		if hashIdx&1 == 1 {
			newHash = chainhash.DoubleHashH(append(h[:], hash[:]...))
		} else {
			newHash = chainhash.DoubleHashH(append(hash[:], h[:]...))
		}
		hash = &newHash
		hashIdx = (hashIdx >> 1) | (1 << treeHeight)
	}

	return bytes.Equal(hash[:], expectedRoot[:])
}

func BuildMerkleTreeStore(txHashes []*chainhash.Hash) []*chainhash.Hash {
	// Calculate how many entries are required to hold the binary merkle
	// tree as a linear array and create an array of that size.
	nextPoT := nextPowerOfTwo(len(txHashes))
	arraySize := nextPoT*2 - 1
	merkles := make([]*chainhash.Hash, arraySize)
	copy(merkles[0:], txHashes[:])

	// Start the array offset after the last transaction and adjusted to the
	// next power of two.
	offset := nextPoT
	for i := 0; i < arraySize-1; i += 2 {
		switch {
		// When there is no left child node, the parent is nil too.
		case merkles[i] == nil:
			merkles[offset] = nil

		// When there is no right child, the parent is generated by
		// hashing the concatenation of the left child with itself.
		case merkles[i+1] == nil:
			newHash := HashMerkleBranches(merkles[i], merkles[i])
			merkles[offset] = newHash

		// The normal case sets the parent node to the double sha256
		// of the concatentation of the left and right children.
		default:
			newHash := HashMerkleBranches(merkles[i], merkles[i+1])
			merkles[offset] = newHash
		}
		offset++
	}

	return merkles
}

func HashMerkleBranches(left *chainhash.Hash, right *chainhash.Hash) *chainhash.Hash {
	// Concatenate the left and right nodes.
	var hash [chainhash.HashSize * 2]byte
	copy(hash[:chainhash.HashSize], left[:])
	copy(hash[chainhash.HashSize:], right[:])

	newHash := chainhash.DoubleHashH(hash[:])
	return &newHash
}

func nextPowerOfTwo(n int) int {
	// Return the number if it's already a power of 2.
	if n&(n-1) == 0 {
		return n
	}

	// Figure out and return the next power of two.
	exponent := uint(math.Log2(float64(n))) + 1
	return 1 << exponent // 2^exponent
}

func calcTreeHeight(n uint64) (e uint) {
	for ; (1 << e) < n; e++ {
	}
	return
}
