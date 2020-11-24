package routes

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"github.com/mit-dci/pooldetective/util"
)

var reorgsCache = sync.Map{}

type ReorgDetailResult struct {
	ID            int                         `json:"id"`
	RemovedBlocks []string                    `json:"removedBlocks"`
	AddedBlocks   []string                    `json:"addedBlocks"`
	DoubleSpends  [][2]DoubleSpendTransaction `json:"doubleSpends"`
}

type DoubleSpendTransaction struct {
	BlockHash []byte                         `json:"-"`
	BlockID   string                         `json:"blockHash"`
	TxHash    []byte                         `json:"-"`
	TxID      string                         `json:"txHash"`
	TxIns     []DoubleSpendTransactionInput  `json:"in"`
	TxOuts    []DoubleSpendTransactionOutput `json:"out"`
}

type DoubleSpendTransactionInput struct {
	Idx         int    `json:"idx"`
	DoubleSpent bool   `json:"doubleSpent"`
	PrevOutTxID string `json:"prevoutTxID"`
	PrevOutIdx  int    `json:"prevoutIdx"`
}

type DoubleSpendTransactionOutput struct {
	Idx     int    `json:"idx"`
	Value   int64  `json:"value"`
	Address string `json:"address"`
}

func reorgHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	reorgIDString := params["reorgID"]
	reorgID, _ := strconv.Atoi(reorgIDString)

	existingData, ok := reorgsCache.Load(reorgID)
	if ok {
		writeJson(w, existingData)
		return
	}

	var addedBlocks pq.ByteaArray
	var removedBlocks pq.ByteaArray

	err := db.QueryRow("SELECT added_blocks, removed_blocks FROM reorgs WHERE id=$1", reorgID).Scan(&addedBlocks, &removedBlocks)

	doubleSpends := [][2]DoubleSpendTransaction{}
	doubleSpentOutpoints := []string{}
	rows, err := db.Query(`SELECT outpoint_txid, outpoint_idx FROM doublespends WHERE reorg_id=$1`, reorgID)
	for rows.Next() {
		var outpointTxID []byte
		var outpointIdx int
		err = rows.Scan(&outpointTxID, &outpointIdx)
		doubleSpentOutpoints = append(doubleSpentOutpoints, fmt.Sprintf("%x-%d", outpointTxID, outpointIdx))
	}

	rows, err = db.Query(`SELECT DISTINCT block_1, block_2, tx_1, tx_2 FROM doublespends WHERE reorg_id=$1`, reorgID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	for rows.Next() {
		ds := [2]DoubleSpendTransaction{DoubleSpendTransaction{}, DoubleSpendTransaction{}}

		var blockHash1 []byte
		var blockHash2 []byte
		var txHash1 []byte
		var txHash2 []byte

		var err = rows.Scan(&blockHash1, &blockHash2, &txHash1, &txHash2)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		ds[0].BlockHash = blockHash1
		ds[0].TxHash = txHash1
		ds[0].BlockID = util.HashToString(blockHash1)
		ds[0].TxID = util.HashToString(txHash1)

		ds[1].BlockHash = blockHash2
		ds[1].TxHash = txHash2
		ds[1].BlockID = util.HashToString(blockHash2)
		ds[1].TxID = util.HashToString(txHash2)

		doubleSpends = append(doubleSpends, ds)
	}

	for i := range doubleSpends {
		for j := 0; j < 2; j++ {
			rows, err = db.Query(`SELECT idx, prevout_txhash, prevout_idx FROM doublespend_txins WHERE txid=$1`, doubleSpends[i][j].TxHash)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			doubleSpends[i][j].TxIns = make([]DoubleSpendTransactionInput, 0)
			for rows.Next() {
				var prevOutTxID []byte
				var prevOutIdx int
				var idx int

				err = rows.Scan(&idx, &prevOutTxID, &prevOutIdx)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}

				txi := DoubleSpendTransactionInput{Idx: idx, PrevOutIdx: prevOutIdx, PrevOutTxID: util.HashToString(prevOutTxID), DoubleSpent: false}
				txiPrevOut := fmt.Sprintf("%x-%d", prevOutTxID, prevOutIdx)
				for _, op := range doubleSpentOutpoints {
					if op == txiPrevOut {
						txi.DoubleSpent = true
						break
					}
				}
				doubleSpends[i][j].TxIns = append(doubleSpends[i][j].TxIns, txi)
			}

			rows, err = db.Query(`SELECT idx, value, script FROM doublespend_txouts WHERE txid=$1`, doubleSpends[i][j].TxHash)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			doubleSpends[i][j].TxOuts = make([]DoubleSpendTransactionOutput, 0)
			for rows.Next() {
				var script []byte
				var value int64
				var idx int

				err = rows.Scan(&idx, &value, &script)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}

				txo := DoubleSpendTransactionOutput{Idx: idx, Value: value, Address: fmt.Sprintf("%x", script)}
				doubleSpends[i][j].TxOuts = append(doubleSpends[i][j].TxOuts, txo)
			}
		}
	}

	rdr := ReorgDetailResult{
		ID:            reorgID,
		RemovedBlocks: make([]string, len(removedBlocks)),
		AddedBlocks:   make([]string, len(addedBlocks)),
		DoubleSpends:  doubleSpends,
	}

	for i, b := range removedBlocks {
		rdr.RemovedBlocks[i] = util.HashToString(b)
	}
	for i, b := range addedBlocks {
		rdr.AddedBlocks[i] = util.HashToString(b)
	}

	reorgsCache.Store(reorgID, rdr)
	writeJson(w, rdr)
}
