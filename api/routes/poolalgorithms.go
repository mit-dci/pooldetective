package routes

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

var poolAlgorithmsHandlerCache map[string][]AlgorithmResult = map[string][]AlgorithmResult{}
var poolAlgorithmsHandlerCacheLock sync.Mutex = sync.Mutex{}
var poolAlgorithmsHandlerCacheLastBuilt map[string]time.Time = map[string]time.Time{}

func poolAlgorithmsHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	poolID := params["poolID"]

	t, ok := poolAlgorithmsHandlerCacheLastBuilt[poolID]
	if !ok || time.Now().Sub(t).Minutes() > 15 {
		poolAlgorithmsHandlerCacheLock.Lock()
		t, ok = poolAlgorithmsHandlerCacheLastBuilt[poolID]
		if !ok || time.Now().Sub(t).Minutes() > 15 {
			rows, err := db.Query("SELECT id, name FROM algorithms WHERE id IN (SELECT algorithm_id FROM pool_observers WHERE pool_id=$1 AND disabled=false)", poolID)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			results := []AlgorithmResult{}
			for rows.Next() {
				var res AlgorithmResult
				err := rows.Scan(&res.ID, &res.Name)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				results = append(results, res)
			}
			poolAlgorithmsHandlerCache[poolID] = results
			poolAlgorithmsHandlerCacheLastBuilt[poolID] = time.Now()
		}
		poolAlgorithmsHandlerCacheLock.Unlock()
	}
	writeJson(w, poolAlgorithmsHandlerCache[poolID])
}
