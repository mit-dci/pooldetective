package main

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/btcsuite/btcd/wire"
	dcrwire "github.com/decred/dcrd/wire"
	"github.com/lib/pq"
	"github.com/mit-dci/pooldetective/logging"
	"github.com/mit-dci/pooldetective/util"
	"github.com/mit-dci/pooldetective/util/zcash"
	xvgwire "github.com/vergecurrency/xvgd/wire"
)

var db *sql.DB
var nullHash []byte
var coins map[int]string

var (
	ErrCannotReadBlock = errors.New("Cannot read block")
)

const MaxSolutionSize = 9999
const VERSION_AUXPOW = (1 << 8)
const VERSION_ALGO_EQUIHASH = (5 << 9)
const VERSION_ALGO_ZHASH = (23 << 9)
const VERSION_ALGO_EH192 = (46 << 9)
const VERSION_ALGO_MARS = (47 << 9)

func main() {
	logging.SetLogLevel(util.GetLoglevelFromEnv())

	var err error
	db, err = sql.Open("postgres", os.Getenv("PGSQL_CONNECTION"))
	if err != nil {
		logging.Fatal(err)
	}

	coins = map[int]string{}

	rows, err := db.Query("SELECT id, name FROM coins")
	if err != nil {
		logging.Fatal(err)
	}
	for rows.Next() {
		var coinID int
		var coinName string

		rows.Scan(&coinID, &coinName)
		coins[coinID] = coinName
	}

	for {
		rows, err := db.Query("SELECT id, coin_id, removed_blocks, added_blocks FROM reorgs WHERE analyzed=false OR analyzed IS NULL")
		if err != nil {
			logging.Fatal(err)
		}
		i := 0
		for rows.Next() {
			i++
			var removedBlocks pq.ByteaArray
			var addedBlocks pq.ByteaArray
			var reorgID int
			var coinID int
			err := rows.Scan(&reorgID, &coinID, &removedBlocks, &addedBlocks)
			if err != nil {
				logging.Fatal(err)
			}

			err = processReorg(reorgID, coinID, removedBlocks, addedBlocks)
			if err == ErrCannotReadBlock {
				// Block missing - mark processed and blocksmissing for later fix
				_, err = db.Exec("UPDATE reorgs SET failed=true, blocksmissing=true, analyzed=true, failurereason=NULL WHERE ID=$1", reorgID)
				if err != nil {
					logging.Fatal(err)
				}
			} else if err != nil {

				if strings.HasPrefix(err.Error(), "Error while reading transaction: readScript: transaction input signature script is larger than the max allowed size") {
					err = errors.New("Error while reading transaction: readScript: transaction input signature script is larger than the max allowed size")
				}
				if strings.HasPrefix(err.Error(), "Error while reading transaction: readScript: transaction output public key script is larger than the max allowed size") {
					err = errors.New("Error while reading transaction: readScript: transaction output public key script is larger than the max allowed size")
				}

				_, err = db.Exec("UPDATE reorgs SET failed=true, blocksmissing=false, analyzed=true, failurereason=$2 WHERE ID=$1", reorgID, err.Error())
				if err != nil {
					logging.Fatal(err)
				}
			} else {
				_, err = db.Exec("UPDATE reorgs SET analyzed=true, blocksmissing=false, failed=false, failurereason=NULL WHERE ID=$1", reorgID)
				if err != nil {
					logging.Fatal(err)
				}
			}
			fmt.Printf("\rProcessed %d reorgs...", i)
		}
		_, err = db.Exec("UPDATE reorgs SET currency_price=COALESCE((SELECT price_usd FROM coin_pricehistory WHERE coin_id=reorgs.coin_id AND time < reorgs.observed ORDER BY time desc limit 1),0) WHERE currency_price IS NULL or currency_price=0")
		if err != nil {
			logging.Error(err)
		}
		_, err = db.Exec("UPDATE reorgs SET bitcoin_price=COALESCE((SELECT price_usd FROM coin_pricehistory WHERE coin_id=1 AND time < reorgs.observed ORDER BY time desc limit 1),0) WHERE bitcoin_price IS NULL or bitcoin_price=0")
		if err != nil {
			logging.Error(err)
		}
		_, err = db.Exec("UPDATE reorgs SET nicehash_price=COALESCE((SELECT price FROM nicehash_pricehistory WHERE algorithm_id=(SELECT algorithm_id FROM coins WHERE id=reorgs.coin_id) AND time < reorgs.observed ORDER BY time desc limit 1),0) WHERE nicehash_price IS NULL or nicehash_price=0")
		if err != nil {
			logging.Error(err)
		}
		time.Sleep(time.Second * 30)
	}
}

type Block struct {
	Hash         []byte
	Transactions []Transaction
}

type Transaction struct {
	Hash    []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

type TxOutput struct {
	Value  int64
	Script []byte
}

type TxInput struct {
	PrevOutTxHash []byte
	PrevOutIndex  uint32
}

type SpentInput struct {
	TxIn      TxInput
	TxID      []byte
	BlockHash []byte
}

func processReorg(reorgID int, coinID int, removedBlocks, addedBlocks [][]byte) error {
	var err error
	removed := make([]Block, len(removedBlocks))
	added := make([]Block, len(addedBlocks))

	for i, b := range removedBlocks {
		removed[i], err = ReadBlock(coinID, b)
		if err != nil {
			return err
		}
	}

	for i, b := range addedBlocks {
		added[i], err = ReadBlock(coinID, b)
		if err != nil {
			return err
		}
	}

	inputsInRemovedBlocks := make([]SpentInput, 0)
	inputsInAddedBlocks := make([]SpentInput, 0)

	for _, rem := range removed {
		for _, remTx := range rem.Transactions {
			for _, remTxIn := range remTx.Inputs {
				if !bytes.Equal(remTxIn.PrevOutTxHash, nullHash) {
					spentInput := SpentInput{
						TxIn:      remTxIn,
						TxID:      remTx.Hash,
						BlockHash: rem.Hash,
					}
					inputsInRemovedBlocks = append(inputsInRemovedBlocks, spentInput)
				}
			}
		}
	}

	totalGenerated := int64(0)
	for _, add := range added {
		for _, addTx := range add.Transactions {

			if len(addTx.Inputs) > 0 && bytes.Equal(addTx.Inputs[0].PrevOutTxHash, nullHash) {
				for _, o := range addTx.Outputs {
					totalGenerated += o.Value
				}
			} else {
				for _, addTxIn := range addTx.Inputs {
					if !bytes.Equal(addTxIn.PrevOutTxHash, nullHash) {
						spentInput := SpentInput{
							TxIn:      addTxIn,
							TxID:      addTx.Hash,
							BlockHash: add.Hash,
						}
						inputsInAddedBlocks = append(inputsInAddedBlocks, spentInput)
					}
				}
			}
		}
	}

	_, err = db.Exec("UPDATE reorgs SET total_generated_coins=$1 WHERE id=$2", totalGenerated, reorgID)
	if err != nil {
		fmt.Printf("Could not register total_generated_coins in database: %s\n", err.Error())
	}

	// Find double spent outputs
	for _, addedInputSpend := range inputsInAddedBlocks {
		for _, removedInputSpend := range inputsInRemovedBlocks {
			if OutPointsEqual(addedInputSpend.TxIn, removedInputSpend.TxIn) {
				if !bytes.Equal(addedInputSpend.TxID, removedInputSpend.TxID) {

					removedTx := Transaction{}
					addedTx := Transaction{}

					for _, ab := range added {
						for _, a := range ab.Transactions {
							if bytes.Equal(a.Hash, addedInputSpend.TxID) {
								addedTx = a
								break
							}
						}
					}

					for _, rb := range added {
						for _, r := range rb.Transactions {
							if bytes.Equal(r.Hash, removedInputSpend.TxID) {
								removedTx = r
								break
							}
						}
					}

					err := SaveDoublespendTx(addedTx)
					if err != nil {
						fmt.Printf("Could not register double spend transaction 1 in database: %s\n", err.Error())
					}
					err = SaveDoublespendTx(removedTx)
					if err != nil {
						fmt.Printf("Could not register double spend transaction 2 in database: %s\n", err.Error())
					}

					_, err = db.Exec("INSERT INTO doublespends(reorg_id, block_1, block_2, tx_1, tx_2, outpoint_txid, outpoint_idx) VALUES ($1,$2,$3,$4,$5,$6,$7)", reorgID, addedInputSpend.BlockHash, removedInputSpend.BlockHash, addedInputSpend.TxID, removedInputSpend.TxID, addedInputSpend.TxIn.PrevOutTxHash, addedInputSpend.TxIn.PrevOutIndex)
					if err != nil {
						fmt.Printf("Could not register double spend in database: %s\n", err.Error())
					}
				}
			}
		}
	}
	return nil

}

func SaveDoublespendTx(tx Transaction) error {
	count := -1
	err := db.QueryRow("SELECT count(*) FROM doublespend_txins WHERE txid=$1", tx.Hash).Scan(&count)
	if err == nil && count > 0 {
		return nil
	}

	dbtx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := dbtx.Prepare(pq.CopyIn("doublespend_txins", "txid", "idx", "prevout_txhash", "prevout_idx"))
	if err != nil {
		return err
	}

	for i, in := range tx.Inputs {
		_, err := stmt.Exec(tx.Hash, i, in.PrevOutTxHash, in.PrevOutIndex)
		if err != nil {
			return err
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	err = dbtx.Commit()
	if err != nil {
		return err
	}

	dbtx, err = db.Begin()
	if err != nil {
		return err
	}

	stmt, err = dbtx.Prepare(pq.CopyIn("doublespend_txouts", "txid", "idx", "value", "script"))
	if err != nil {
		return err
	}

	for i, out := range tx.Outputs {
		_, err := stmt.Exec(tx.Hash, i, out.Value, out.Script)
		if err != nil {
			return err
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	err = dbtx.Commit()
	return err
}

func OutPointsEqual(o1, o2 TxInput) bool {
	if o1.PrevOutIndex != o2.PrevOutIndex {
		return false
	}
	if !bytes.Equal(o1.PrevOutTxHash, o2.PrevOutTxHash) {
		return false
	}
	return true
}

func ReadBlockHeader(b *Block, coinId int, r io.Reader) error {

	// We're not really reading it since we don't care what's in it -
	// the block header data is already in our database. We just need to skip enough
	// bytes to get to the TX count.

	switch coinId {
	case 30, 34, 11, 29, 35, 10, 37, 32, 28: // ZEC and forks
		b := make([]byte, 140)
		util.ReadElement(r, &b)
		_, err := util.ReadVarBytes(r, 0, 1344, "Solution")
		if err != nil {
			return err
		}
	case 17: // GTO
		var nVersion uint32
		err := util.ReadElement(r, &nVersion)
		if err != nil {
			return err
		}

		if nVersion&VERSION_AUXPOW != 0 {
			return fmt.Errorf("Cannot process AUXPOW blocks (yet)")
		}
		// If Equihash based read equihash header, otherwise default
		if nVersion&VERSION_ALGO_EQUIHASH != 0 || nVersion&VERSION_ALGO_ZHASH != 0 || nVersion&VERSION_ALGO_EH192 != 0 || nVersion&VERSION_ALGO_MARS != 0 {
			b := make([]byte, 136)
			util.ReadElement(r, &b)
			_, err = util.ReadVarBytes(r, 0, 1344, "Solution")
			if err != nil {
				return err
			}
		} else {
			b := make([]byte, 76)
			util.ReadElement(r, &b)
		}
	default:
		b := make([]byte, 80)
		util.ReadElement(r, &b)
	}
	txCount, err := wire.ReadVarInt(r, 0)
	if err != nil {
		return err
	}
	if txCount == 0 {
		return fmt.Errorf("Block cannot have zero transactions")
	}
	//fmt.Printf("Coin: %s (%d) - Block %x - TXCount: %d", coins[coinId], coinId, b.Hash, txCount)

	b.Transactions = make([]Transaction, txCount)
	return nil
}

// ReadOldTransaction reads a transaction from way back, which included the nTime after version.
func ReadOldTransaction(r io.Reader) (Transaction, error) {
	btx := Transaction{}

	var version uint32
	err := util.ReadElement(r, &version)
	if err != nil {
		return btx, err
	}

	var nTime uint32
	err = util.ReadElement(r, &nTime)
	if err != nil {
		return btx, err
	}

	count, err := wire.ReadVarInt(r, 0)
	if err != nil {
		return btx, err
	}

	// Deserialize the inputs.
	txIns := make([]wire.TxIn, count)
	for i := uint64(0); i < count; i++ {
		// The pointer is set now in case a script buffer is borrowed
		// and needs to be returned to the pool on error.
		ti := &txIns[i]
		err = util.ReadTxIn(r, ti)
		if err != nil {
			return btx, err
		}
	}

	count, err = wire.ReadVarInt(r, 0)
	if err != nil {
		return btx, err
	}

	// Deserialize the outputs.
	btx.Outputs = make([]TxOutput, count)
	to := &wire.TxOut{}
	for i := uint64(0); i < count; i++ {
		err = util.ReadTxOut(r, to)
		if err != nil {
			return btx, err
		}
		btx.Outputs[i] = TxOutput{
			Value:  to.Value,
			Script: to.PkScript,
		}
	}

	// Read locktime
	err = util.ReadElement(r, &nTime)
	if err != nil {
		return btx, err
	}

	// Todo btx.Hash
	btx.Inputs = make([]TxInput, len(txIns))
	for i, ip := range txIns {
		btx.Inputs[i] = TxInput{
			PrevOutTxHash: ip.PreviousOutPoint.Hash.CloneBytes(),
			PrevOutIndex:  ip.PreviousOutPoint.Index,
		}
	}
	return btx, nil
}

func ReadTransaction(coinId int, r io.Reader) (Transaction, error) {
	var err error
	btx := Transaction{}

	switch coinId {
	case 30, 34, 11, 29, 35, 37, 32, 28: // ZCash 'n forks
		tx := zcash.Transaction{}
		_, err = tx.ReadFrom(r)
		if err != nil {
			return btx, err
		}
		txh := tx.TxHash()
		btx.Inputs = make([]TxInput, len(tx.Inputs))
		for i, ip := range tx.Inputs {
			btx.Inputs[i] = TxInput{
				PrevOutTxHash: ip.PreviousOutPoint.Hash.CloneBytes(),
				PrevOutIndex:  ip.PreviousOutPoint.Index,
			}
		}
		btx.Outputs = make([]TxOutput, len(tx.Outputs))
		for i, op := range tx.Outputs {
			btx.Outputs[i] = TxOutput{
				Value:  op.Value,
				Script: op.ScriptPubKey,
			}
		}

		btx.Hash = txh.CloneBytes()
	case 33: // XVG
		tx := xvgwire.NewMsgTx(1)
		err = tx.Deserialize(r)
		if err != nil {
			return btx, err
		}
		txh := tx.TxHash()
		btx.Inputs = make([]TxInput, len(tx.TxIn))
		for i, ip := range tx.TxIn {
			btx.Inputs[i] = TxInput{
				PrevOutTxHash: ip.PreviousOutPoint.Hash.CloneBytes(),
				PrevOutIndex:  ip.PreviousOutPoint.Index,
			}
		}
		btx.Outputs = make([]TxOutput, len(tx.TxOut))
		for i, op := range tx.TxOut {
			btx.Outputs[i] = TxOutput{
				Value:  op.Value,
				Script: op.PkScript,
			}
		}
		btx.Hash = txh.CloneBytes()
	case 9: // DCR
		tx := dcrwire.NewMsgTx()
		err = tx.Deserialize(r)
		if err != nil {
			return btx, err
		}
		txh := tx.TxHash()
		btx.Inputs = make([]TxInput, len(tx.TxIn))
		for i, ip := range tx.TxIn {
			btx.Inputs[i] = TxInput{
				PrevOutTxHash: ip.PreviousOutPoint.Hash.CloneBytes(),
				PrevOutIndex:  ip.PreviousOutPoint.Index,
			}
		}
		btx.Outputs = make([]TxOutput, len(tx.TxOut))
		for i, op := range tx.TxOut {
			btx.Outputs[i] = TxOutput{
				Value:  op.Value,
				Script: op.PkScript,
			}
		}
		btx.Hash = txh.CloneBytes()
	case 42, 43: // NSD, EUNO - Old format
		btx, err = ReadOldTransaction(r)
		if err != nil {
			return btx, err
		}
	case 13: // DOGE
		tx := wire.NewMsgTx(1)

		err = tx.DeserializeNoWitness(r)
		if err != nil {
			return btx, err
		}
		txh := tx.TxHash()
		btx.Inputs = make([]TxInput, len(tx.TxIn))
		for i, ip := range tx.TxIn {
			btx.Inputs[i] = TxInput{
				PrevOutTxHash: ip.PreviousOutPoint.Hash.CloneBytes(),
				PrevOutIndex:  ip.PreviousOutPoint.Index,
			}
		}
		btx.Outputs = make([]TxOutput, len(tx.TxOut))
		for i, op := range tx.TxOut {
			btx.Outputs[i] = TxOutput{
				Value:  op.Value,
				Script: op.PkScript,
			}
		}
		btx.Hash = txh.CloneBytes()
	default:
		tx := wire.NewMsgTx(1)

		err = tx.Deserialize(r)
		if err != nil {
			return btx, err
		}
		txh := tx.TxHash()
		btx.Inputs = make([]TxInput, len(tx.TxIn))
		for i, ip := range tx.TxIn {
			btx.Inputs[i] = TxInput{
				PrevOutTxHash: ip.PreviousOutPoint.Hash.CloneBytes(),
				PrevOutIndex:  ip.PreviousOutPoint.Index,
			}
		}
		btx.Outputs = make([]TxOutput, len(tx.TxOut))
		for i, op := range tx.TxOut {
			btx.Outputs[i] = TxOutput{
				Value:  op.Value,
				Script: op.PkScript,
			}
		}
		btx.Hash = txh.CloneBytes()
	}
	return btx, nil
}

func ReadBlock(coinId int, hash []byte) (Block, error) {
	blockFile := fmt.Sprintf("%s/%x.blk", os.Getenv("BLOCKSDIR"), hash)

	block := Block{}
	block.Hash = hash
	f, err := os.Open(blockFile)
	if err != nil {
		return Block{}, ErrCannotReadBlock
	}
	defer f.Close()

	err = ReadBlockHeader(&block, coinId, f)
	if err != nil {
		return Block{}, fmt.Errorf("Error while reading blockheader: %s", err.Error())
	}

	for i := range block.Transactions {
		block.Transactions[i], err = ReadTransaction(coinId, f)
		if err != nil {
			return Block{}, fmt.Errorf("Error while reading transaction: %s", err.Error())
		}
	}

	/*if coinId == 10 || coinId == 34 {
		wireBlock := &btgwire.MsgBlock{}
		err = wireBlock.Deserialize(f)
		if err != nil {
			return Block{}, err
		}
		bh := wireBlock.BlockHash()
		block.Hash = bh.CloneBytes()
		block.Transactions = make([]Transaction, len(wireBlock.Transactions))
		for i, tx := range wireBlock.Transactions {
			inputs := make([]TxInput, len(tx.TxIn))
			for j, txi := range tx.TxIn {
				inputs[j] = TxInput{
					PrevOutTxHash: txi.PreviousOutPoint.Hash.CloneBytes(),
					PrevOutIndex:  txi.PreviousOutPoint.Index,
				}
			}
			h := tx.TxHash()
			block.Transactions[i] = Transaction{Inputs: inputs, Hash: h.CloneBytes()}
		}

	} else {
		wireBlock := &wire.MsgBlock{}
		if coinId == 13 || coinId == 33 { // DOGE and XVG
			err = wireBlock.DeserializeNoWitness(f)
		} else {
			err = wireBlock.Deserialize(f)
		}
		if err != nil {
			return Block{}, err
		}
		bh := wireBlock.BlockHash()
		block.Hash = bh.CloneBytes()
		block.Transactions = make([]Transaction, len(wireBlock.Transactions))
		for i, tx := range wireBlock.Transactions {
			inputs := make([]TxInput, len(tx.TxIn))
			for j, txi := range tx.TxIn {
				inputs[j] = TxInput{
					PrevOutTxHash: txi.PreviousOutPoint.Hash.CloneBytes(),
					PrevOutIndex:  txi.PreviousOutPoint.Index,
				}
			}
			h := tx.TxHash()
			block.Transactions[i] = Transaction{Inputs: inputs, Hash: h.CloneBytes()}
		}
	}*/

	return block, nil

}

func init() {
	nullHash, _ = hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
}
