package routes

import (
	"net/http"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gorilla/mux"
)

type CoinDetailResult struct {
	MinHeight  int64                 `json:"minHeight"`
	MaxHeight  int64                 `json:"maxHeight"`
	NumBlocks  int64                 `json:"numBlocks"`
	TipHash    string                `json:"tipHash"`
	Algorithms []CoinDetailAlgorithm `json:"algorithms"`
}

type CoinDetailAlgorithm struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
}

func coinHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	coinID := params["coinID"]
	var result CoinDetailResult
	err := db.QueryRow("SELECT min(height), max(height), count(*) from blocks b WHERE b.coin_id=$1 AND b.height IS NOT NULL", coinID).Scan(&result.MinHeight, &result.MaxHeight, &result.NumBlocks)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var bestTip []byte
	err = db.QueryRow("SELECT block_hash from blocks b WHERE b.coin_id=$1 AND b.height IS NOT NULL ORDER BY b.height DESC LIMIT 1", coinID).Scan(&bestTip)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	h, err := chainhash.NewHash(bestTip)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	result.TipHash = h.String()

	rows, err := db.Query("SELECT COALESCE(a.id,-1), COALESCE(a.name,'') from coin_algorithm ca left join algorithms a on a.id=ca.algorithm_id where ca.coin_id=$1", coinID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	result.Algorithms = make([]CoinDetailAlgorithm, 0)
	for rows.Next() {
		a := CoinDetailAlgorithm{}
		err = rows.Scan(&a.ID, &a.Name)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		result.Algorithms = append(result.Algorithms, a)
	}

	writeJson(w, result)
}
