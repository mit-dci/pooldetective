package routes

import (
	"net/http"
	"sync"
	"time"
)

type PoolResult struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

var poolsHandlerCache []PoolResult
var poolsHandlerCacheLock sync.Mutex = sync.Mutex{}
var poolsHandlerCacheLastBuilt time.Time = time.Now().Add(-24 * time.Hour)

func poolsHandler(w http.ResponseWriter, r *http.Request) {
	if time.Now().Sub(poolsHandlerCacheLastBuilt).Minutes() > 15 {
		poolsHandlerCacheLock.Lock()
		if time.Now().Sub(poolsHandlerCacheLastBuilt).Minutes() > 15 {
			rows, err := db.Query("SELECT id, name FROM pools WHERE id IN (SELECT pool_id FROM pool_observers WHERE disabled=false)")
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			results := []PoolResult{}
			for rows.Next() {
				var res PoolResult
				err := rows.Scan(&res.ID, &res.Name)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				results = append(results, res)
			}
			poolsHandlerCache = results
			poolsHandlerCacheLastBuilt = time.Now()
		}
		poolsHandlerCacheLock.Unlock()
	}
	writeJson(w, poolsHandlerCache)
}
