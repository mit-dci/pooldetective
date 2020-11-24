package routes

import (
	"net/http"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type PoolResult struct {
	ID            int64                 `json:"id"`
	Name          string                `json:"name"`
	CoinID        int64                 `json:"coinId"`
	CoinName      string                `json:"coinName"`
	PoolObservers []*PoolObserverResult `json:"observers"`
}

type PoolObserverResult struct {
	ID              int64     `json:"id"`
	StratumHost     string    `json:"stratumHost"`
	StratumPort     int64     `json:"stratumPort"`
	LocationID      int64     `json:"locationId"`
	LocationName    string    `json:"locationName"`
	LastJobReceived time.Time `json:"lastJobReceived"`
	LastJobID       int64     `json:"lastJobID"`
	LastJobPrevHash string    `json:"lastJobPrevHash"`
}

var poolsHandlerCache []*PoolResult
var poolsHandlerCacheLock sync.Mutex = sync.Mutex{}
var poolsHandlerCacheLastBuilt time.Time = time.Now().Add(-24 * time.Hour)

func UpdatePools() {
	for {
		rows, err := db.Query("SELECT po.id, j.id, j.previous_block_hash, po.last_job_received FROM pool_observers po LEFT JOIN jobs j on j.id=po.last_job_id WHERE disabled=false")
		if err == nil {
			for rows.Next() {
				var id int64
				var lastJobID int64
				var lastJobPrevHash []byte
				var lastJobReceived time.Time

				rows.Scan(&id, &lastJobID, &lastJobPrevHash, &lastJobReceived)

				for i := range poolsHandlerCache {
					for j := range poolsHandlerCache[i].PoolObservers {
						if poolsHandlerCache[i].PoolObservers[j].ID == id && poolsHandlerCache[i].PoolObservers[j].LastJobID != lastJobID {
							poolsHandlerCache[i].PoolObservers[j].LastJobID = lastJobID
							poolsHandlerCache[i].PoolObservers[j].LastJobReceived = lastJobReceived
							ph, err := chainhash.NewHash(lastJobPrevHash)
							if err == nil {
								poolsHandlerCache[i].PoolObservers[j].LastJobPrevHash = ph.String()
								publishToWebsockets(websocketMessage{Type: "p", Msg: poolsHandlerCache[i].PoolObservers[j]})
							}
						}
					}
				}

			}
		}
		time.Sleep(time.Second * 2)
	}
}

func poolsHandler(w http.ResponseWriter, r *http.Request) {
	if time.Now().Sub(poolsHandlerCacheLastBuilt).Minutes() > 15 {
		poolsHandlerCacheLock.Lock()
		if time.Now().Sub(poolsHandlerCacheLastBuilt).Minutes() > 15 {
			err := CachePoolData()
			if err != nil {
				http.Error(w, err.Error(), 500)
				poolsHandlerCacheLock.Unlock()
				return
			}
		}
		poolsHandlerCacheLock.Unlock()
	}
	writeJson(w, poolsHandlerCache)
}

func CachePoolData() error {
	rows, err := db.Query("SELECT p.id, p.name, c.id, c.name, l.id, l.name, po.id, po.stratum_host, po.stratum_port, po.last_job_received, j.id, j.previous_block_hash FROM pools p left join pool_observers po on (po.pool_id=p.id AND (po.disabled=false or po.disabled is null)) left join coins c on c.id=po.coin_id left join jobs j on j.id=po.last_job_id left join locations l on l.id=po.location_id WHERE c.id IS NOT NULL ORDER BY p.name, c.name")
	if err != nil {
		return err
	}
	results := []*PoolResult{}
	for rows.Next() {

		poolObserverResult := PoolObserverResult{}
		poolResult := PoolResult{PoolObservers: []*PoolObserverResult{}}
		lastJobPrevHash := []byte{}

		err := rows.Scan(
			&poolResult.ID,
			&poolResult.Name,
			&poolResult.CoinID,
			&poolResult.CoinName,
			&poolObserverResult.LocationID,
			&poolObserverResult.LocationName,
			&poolObserverResult.ID,
			&poolObserverResult.StratumHost,
			&poolObserverResult.StratumPort,
			&poolObserverResult.LastJobReceived,
			&poolObserverResult.LastJobID,
			&lastJobPrevHash,
		)

		ph, err := chainhash.NewHash(lastJobPrevHash)
		if err == nil {
			poolObserverResult.LastJobPrevHash = ph.String()
		}

		if err != nil {
			return err
		}

		idx := -1
		for i, p := range results {
			if p.ID == poolResult.ID && p.CoinID == poolResult.CoinID {
				idx = i
				break
			}
		}
		if idx == -1 {
			results = append(results, &poolResult)
			idx = len(results) - 1
		}

		results[idx].PoolObservers = append(results[idx].PoolObservers, &poolObserverResult)
	}
	poolsHandlerCache = results
	poolsHandlerCacheLastBuilt = time.Now()
	return nil
}
