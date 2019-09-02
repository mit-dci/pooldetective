package routes

import (
	"net/http"
	"sync"
	"time"
)

type CoinResult struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

var coinsHandlerCache []CoinResult
var coinsHandlerCacheLock sync.Mutex = sync.Mutex{}
var coinsHandlerCacheLastBuilt time.Time = time.Now().Add(-24 * time.Hour)

func coinsHandler(w http.ResponseWriter, r *http.Request) {
	if time.Now().Sub(coinsHandlerCacheLastBuilt).Minutes() > 15 {
		coinsHandlerCacheLock.Lock()
		if time.Now().Sub(coinsHandlerCacheLastBuilt).Minutes() > 15 {
			rows, err := db.Query("SELECT id, name FROM coins")
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			result := []CoinResult{}
			for rows.Next() {
				var coinID int64
				var coinName string
				err := rows.Scan(&coinID, &coinName)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				result = append(result, CoinResult{ID: coinID, Name: coinName})
			}
			coinsHandlerCache = result
			coinsHandlerCacheLastBuilt = time.Now()
		}
		coinsHandlerCacheLock.Unlock()
	}

	writeJson(w, coinsHandlerCache)
}
