package routes

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

var poolCoinsHandlerCache map[string][]CoinResult = map[string][]CoinResult{}
var poolCoinsHandlerCacheLock sync.Mutex = sync.Mutex{}
var poolCoinsHandlerCacheLastBuilt map[string]time.Time = map[string]time.Time{}

func poolCoinsHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	poolID := params["poolID"]

	t, ok := poolCoinsHandlerCacheLastBuilt[poolID]
	if !ok || time.Now().Sub(t).Minutes() > 15 {
		poolCoinsHandlerCacheLock.Lock()
		t, ok = poolCoinsHandlerCacheLastBuilt[poolID]
		if !ok || time.Now().Sub(t).Minutes() > 15 {
			rows, err := db.Query("SELECT id, name FROM coins WHERE id IN (SELECT coin_id FROM pool_observers WHERE pool_id=$1 AND disabled=false)", poolID)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			results := []CoinResult{}
			for rows.Next() {
				var res CoinResult
				err := rows.Scan(&res.ID, &res.Name)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				results = append(results, res)
			}
			poolCoinsHandlerCache[poolID] = results
			poolCoinsHandlerCacheLastBuilt[poolID] = time.Now()
		}
		poolCoinsHandlerCacheLock.Unlock()
	}
	writeJson(w, poolCoinsHandlerCache[poolID])
}
