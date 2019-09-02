package routes

import (
	"net/http"
	"sync"
	"time"
)

type AlgorithmResult struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

var algorithmsHandlerCache []AlgorithmResult
var algorithmsHandlerCacheLock sync.Mutex = sync.Mutex{}
var algorithmsHandlerCacheLastBuilt time.Time = time.Now().Add(-24 * time.Hour)

func algorithmsHandler(w http.ResponseWriter, r *http.Request) {
	if time.Now().Sub(algorithmsHandlerCacheLastBuilt).Minutes() > 15 {
		algorithmsHandlerCacheLock.Lock()
		if time.Now().Sub(algorithmsHandlerCacheLastBuilt).Minutes() > 15 {
			rows, err := db.Query("SELECT id, name FROM algorithms")
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			result := []AlgorithmResult{}
			for rows.Next() {
				var algoID int64
				var algoName string
				err := rows.Scan(&algoID, &algoName)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				result = append(result, AlgorithmResult{ID: algoID, Name: algoName})
			}
			algorithmsHandlerCache = result
			algorithmsHandlerCacheLastBuilt = time.Now()
		}
		algorithmsHandlerCacheLock.Unlock()
	}
	writeJson(w, algorithmsHandlerCache)
}
