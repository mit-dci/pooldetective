package routes

import (
	"math/big"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/mit-dci/pooldetective/util"
)

type ReorgResult struct {
	ID                 int64     `json:"id"`
	RemovedBlocks      int64     `json:"removedBlocks"`
	AddedBlocks        int64     `json:"addedBlocks"`
	ForkBlock          string    `json:"forkBlock"`
	DoubleSpentOutputs int       `json:"doubleSpentOutputs"`
	CoinName           string    `json:"coinName"`
	CoinID             int       `json:"coinID"`
	CoinTicker         string    `json:"coinTicker"`
	BudishCost         float64   `json:"budishCost"`
	NiceHashCost       float64   `json:"niceHashCost"`
	AddedWork          int64     `json:"addedWork"`
	RemovedWork        int64     `json:"removedWork"`
	CoinPrice          float64   `json:"coinPrice"`
	CoinsInAddedBlocks float64   `json:"coinsInAddedBlocks"`
	Occurred           time.Time `json:"occurred"`
	ForkBlockHeight    int       `json:"forkBlockHeight"`
}

var reorgsHandlerCache []ReorgResult
var reorgsHandlerCacheLock sync.Mutex = sync.Mutex{}
var reorgsHandlerCacheLastBuilt time.Time = time.Now().Add(-24 * time.Hour)

func reorgsHandler(w http.ResponseWriter, r *http.Request) {
	start := 0
	startString := r.URL.Query().Get("start")
	start, _ = strconv.Atoi(startString)

	minBlocksRemoved := 0
	minBlocksRemovedString := r.URL.Query().Get("minRemoved")
	minBlocksRemoved, _ = strconv.Atoi(minBlocksRemovedString)

	onlyWithDoubleSpends := r.URL.Query().Get("onlyWithDoubleSpends") == "1"

	coinTicker := r.URL.Query().Get("ticker")

	if time.Now().Sub(reorgsHandlerCacheLastBuilt).Minutes() > 60 {
		reorgsHandlerCacheLock.Lock()
		if time.Now().Sub(reorgsHandlerCacheLastBuilt).Minutes() > 60 {
			result := make([]ReorgResult, 0)
			rows, err := db.Query(`SELECT 
										r.id, 
										cardinality(r.added_blocks) as added_blocks, 
										cardinality(r.removed_blocks) as removed_blocks, 
										r.fork_block_hash, 
										COALESCE(ds.cnt,0) as double_spent_outputs, 
										c.name, 
										c.id, 
										trim(c.ticker), 
										r.observed, 
										r.fork_total_chainwork, 
										r.removed_total_chainwork, 
										r.added_total_chainwork, 
										r.total_generated_coins, 
										COALESCE(r.bitcoin_price,0), 
										COALESCE(r.nicehash_price,0), 
											COALESCE(r.currency_price,0),
										COALESCE((SELECT nicehash_marketfactor FROM algorithms WHERE id=c.algorithm_id), 1) as market_factor,
										COALESCE(r.fork_block_height,-1)
									FROM 
										reorgs r 
										LEFT JOIN coins c on c.id=r.coin_id 
										LEFT JOIN (SELECT reorg_id, count(*) as cnt FROM doublespends GROUP BY reorg_id) ds on ds.reorg_id=r.id 
									WHERE 
										r.analyzed=true 
										AND r.failed=false 
									ORDER BY 
										r.observed DESC`)
			if err != nil {
				http.Error(w, err.Error(), 500)
				reorgsHandlerCacheLock.Unlock()
				return
			}

			defer rows.Close()
			for rows.Next() {
				var r ReorgResult

				var forkTotalChainWork []byte
				var addedTotalChainWork []byte
				var removedTotalChainWork []byte
				var addedCoins int64
				var nicehashPrice float64
				var bitcoinPrice float64
				var forkBlockHash []byte
				var marketFactor int64
				err := rows.Scan(
					&r.ID,
					&r.AddedBlocks,
					&r.RemovedBlocks,
					&forkBlockHash,
					&r.DoubleSpentOutputs,
					&r.CoinName,
					&r.CoinID,
					&r.CoinTicker,
					&r.Occurred,
					&forkTotalChainWork,
					&removedTotalChainWork,
					&addedTotalChainWork,
					&addedCoins,
					&bitcoinPrice,
					&nicehashPrice,
					&r.CoinPrice,
					&marketFactor,
					&r.ForkBlockHeight)

				r.ForkBlock = util.HashToString(forkBlockHash)
				r.CoinsInAddedBlocks = float64(addedCoins) / float64(100000000)

				forkTotalChainWorkInt := big.NewInt(0).SetBytes(forkTotalChainWork)
				addedTotalChainWorkInt := big.NewInt(0).SetBytes(addedTotalChainWork)
				removedTotalChainWorkInt := big.NewInt(0).SetBytes(removedTotalChainWork)

				r.AddedWork = big.NewInt(0).Sub(addedTotalChainWorkInt, forkTotalChainWorkInt).Int64()
				r.RemovedWork = big.NewInt(0).Sub(removedTotalChainWorkInt, forkTotalChainWorkInt).Int64()
				r.BudishCost = r.CoinsInAddedBlocks * r.CoinPrice
				r.NiceHashCost = float64(r.AddedWork) * ((nicehashPrice / float64(100000000)) / 86400) * bitcoinPrice

				if err != nil {
					http.Error(w, err.Error(), 500)
					reorgsHandlerCacheLock.Unlock()
					return
				}
				result = append(result, r)
			}

			reorgsHandlerCacheLastBuilt = time.Now()
			reorgsHandlerCache = result
		}
		reorgsHandlerCacheLock.Unlock()

	}

	result := make([]ReorgResult, 0)
	if minBlocksRemoved == 0 {
		result = reorgsHandlerCache
	} else {
		for _, r := range reorgsHandlerCache {
			if r.RemovedBlocks >= int64(minBlocksRemoved) {
				result = append(result, r)
			}
		}
	}

	if coinTicker != "" {
		result2 := make([]ReorgResult, 0)
		for _, r := range result {
			if r.CoinTicker == coinTicker {
				result2 = append(result2, r)
			}
		}
		result = result2
	}

	if onlyWithDoubleSpends {
		result2 := make([]ReorgResult, 0)
		for _, r := range result {
			if r.DoubleSpentOutputs > 0 {
				result2 = append(result2, r)
			}
		}
		result = result2
	}

	if start > len(result) {
		writeJson(w, []ReorgResult{})
		return
	}

	end := start + 100
	if end > len(result) {
		end = len(result)
	}

	writeJson(w, result[start:end])
}
