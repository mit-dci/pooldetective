package routes

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type CoinPoolDetailResult struct {
	Observers []CoinPoolDetailObserver `json:"observers"`
}

type CoinPoolDetailObserver struct {
	ID                int64     `json:"id"`
	Location          string    `json:"location"`
	PoolServer        string    `json:"poolServer"`
	LastJobReceived   time.Time `json:"lastJobReceived"`
	LastShareFound    time.Time `json:"lastShareFound"`
	CurrentDifficulty float64   `json:"currentDifficulty"`
}

func coinPoolHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	coinID := params["coinID"]
	poolID := params["poolID"]

	rows, err := db.Query("SELECT po.id, l.name, po.stratum_host, po.last_job_received, po.last_share_found, po.stratum_difficulty FROM pool_observers po left join locations l on l.id=po.location_id WHERE po.coin_id=$1 AND po.pool_id=$2 AND po.disabled=false", coinID, poolID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	observers := []CoinPoolDetailObserver{}
	for rows.Next() {
		observer := CoinPoolDetailObserver{}
		err := rows.Scan(&observer.ID, &observer.Location, &observer.PoolServer, &observer.LastJobReceived, &observer.LastShareFound, &observer.CurrentDifficulty)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		observers = append(observers, observer)
	}
	writeJson(w, CoinPoolDetailResult{Observers: observers})
}

func algorithmPoolHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	algoID := params["algorithmID"]
	poolID := params["poolID"]

	rows, err := db.Query("SELECT po.id, l.name, po.stratum_host, po.last_job_received, po.last_share_found, po.stratum_difficulty FROM pool_observers po left join locations l on l.id=po.location_id WHERE po.algorithm_id=$1 AND po.pool_id=$2 AND po.disabled=false", algoID, poolID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	observers := []CoinPoolDetailObserver{}
	for rows.Next() {
		observer := CoinPoolDetailObserver{}
		err := rows.Scan(&observer.ID, &observer.Location, &observer.PoolServer, &observer.LastJobReceived, &observer.LastShareFound, &observer.CurrentDifficulty)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		observers = append(observers, observer)
	}
	writeJson(w, CoinPoolDetailResult{Observers: observers})
}
