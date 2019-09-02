package routes

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

var coinPoolsHandlerCache map[string][]PoolResult = map[string][]PoolResult{}
var coinPoolsHandlerCacheLock sync.Mutex = sync.Mutex{}
var coinPoolsHandlerCacheLastBuilt map[string]time.Time = map[string]time.Time{}

func coinPoolsHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	coinID := params["coinID"]
	t, ok := coinPoolsHandlerCacheLastBuilt[coinID]
	if !ok || time.Now().Sub(t).Minutes() > 15 {
		coinPoolsHandlerCacheLock.Lock()
		t, ok = coinPoolsHandlerCacheLastBuilt[coinID]
		if !ok || time.Now().Sub(t).Minutes() > 15 {
			rows, err := db.Query("SELECT DISTINCT p.id, p.name FROM pool_observers po left join pools p on p.id=po.pool_id WHERE po.coin_id=$1 AND po.disabled=false", coinID)
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
			coinPoolsHandlerCacheLastBuilt[coinID] = time.Now()
			coinPoolsHandlerCache[coinID] = results
		}
		coinPoolsHandlerCacheLock.Unlock()
	}

	writeJson(w, coinPoolsHandlerCache[coinID])
}

var algorithmPoolsHandlerCache map[string][]PoolResult = map[string][]PoolResult{}
var algorithmPoolsHandlerCacheLock sync.Mutex = sync.Mutex{}
var algorithmPoolsHandlerCacheLastBuilt map[string]time.Time = map[string]time.Time{}

func algorithmPoolsHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	algorithmID := params["algorithmID"]

	t, ok := algorithmPoolsHandlerCacheLastBuilt[algorithmID]
	if !ok || time.Now().Sub(t).Minutes() > 15 {
		algorithmPoolsHandlerCacheLock.Lock()
		t, ok = algorithmPoolsHandlerCacheLastBuilt[algorithmID]
		if !ok || time.Now().Sub(t).Minutes() > 15 {
			rows, err := db.Query("SELECT DISTINCT p.id, p.name FROM pool_observers po left join pools p on p.id=po.pool_id WHERE po.algorithm_id=$1 AND po.disabled=false", algorithmID)
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
			algorithmPoolsHandlerCacheLastBuilt[algorithmID] = time.Now()
			algorithmPoolsHandlerCache[algorithmID] = results
		}
		algorithmPoolsHandlerCacheLock.Unlock()
	}
	writeJson(w, algorithmPoolsHandlerCache[algorithmID])
}
