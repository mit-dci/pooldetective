package routes

import (
	"net/http"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

var coinsHandlerCache []*CoinResult = []*CoinResult{}

type CoinResult struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	Ticker           string    `json:"ticker"`
	BestHash         string    `json:"bestHash"`
	BestHashObserved time.Time `json:"bestHashObserved"`
}

func UpdateCoins() {
	for {
		rows, err := db.Query("SELECT id, name, ticker, besthash, besthashobserved FROM coins")
		if err == nil {
			for rows.Next() {
				var coinID int64
				var coinName string
				var coinTicker string
				var bestHash []byte
				var bestHashObserved time.Time
				err := rows.Scan(&coinID, &coinName, &coinTicker, &bestHash, &bestHashObserved)
				if err != nil {
					break
				}

				bestHashString := ""
				h, err := chainhash.NewHash(bestHash)
				if err == nil {
					bestHashString = h.String()
				}

				idx := -1
				for i, c := range coinsHandlerCache {
					if c.ID == coinID {
						idx = i
						break
					}
				}
				if idx == -1 {
					coinsHandlerCache = append(coinsHandlerCache, &CoinResult{ID: coinID, Name: coinName, Ticker: coinTicker, BestHash: bestHashString, BestHashObserved: bestHashObserved})
				} else {
					if coinsHandlerCache[idx].BestHash != bestHashString {
						coinsHandlerCache[idx].BestHash = bestHashString
						coinsHandlerCache[idx].BestHashObserved = bestHashObserved
						publishToWebsockets(websocketMessage{Type: "c", Msg: coinsHandlerCache[idx]})
					}
				}
			}
		}
		time.Sleep(time.Second * 2)
	}
}

func coinsHandler(w http.ResponseWriter, r *http.Request) {

	writeJson(w, coinsHandlerCache)
}
