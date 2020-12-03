package routes

import (
	"net/http"
	"sync"
	"time"

	"github.com/mit-dci/pooldetective/logging"
)

type EmptyBlockWorkResult struct {
	ObservedOn     time.Time `json:"observedOn"`
	PoolID         int       `json:"poolId"`
	CoinID         int       `json:"coinId"`
	CoinName       string    `json:"coinName"`
	ObserverID     int       `json:"poolObserverID"`
	StratumHost    string    `json:"stratumHost"`
	LocationID     int       `json:"locationID"`
	LocationName   string    `json:"location"`
	TotalJobs      int64     `json:"totalJobs"`
	EmptyBlockJobs int64     `json:"emptyBlockJobs"`
	TotalTime      int64     `json:"totalTimeMs"`
	EmptyBlockTime int64     `json:"emptyBlockTimeMs"`
}

var emptyBlockWorkAllHandlerCache []EmptyBlockWorkResult
var emptyBlockWorkAllHandlerCacheLock sync.Mutex = sync.Mutex{}
var emptyBlockWorkAllHandlerCacheLastBuilt time.Time = time.Now().Add(-24 * time.Hour)

func emptyBlockWorkAllHandler(w http.ResponseWriter, r *http.Request) {
	if emptyBlockWorkAllHandlerCacheLastBuilt.Day() != time.Now().Day() {
		emptyBlockWorkAllHandlerCacheLock.Lock()
		if emptyBlockWorkAllHandlerCacheLastBuilt.Day() != time.Now().Day() {
			results := []EmptyBlockWorkResult{}
			t := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
			for t.Before(time.Now()) {
				resultsForDay, err := emptyBlockWorkOnDay(t)
				if err != nil {
					logging.Errorf("Error: %s", err.Error())
					http.Error(w, "Internal server error", 500)
					wrongWorkAllHandlerCacheLock.Unlock()
					return
				}
				results = append(results, resultsForDay...)
				t = t.Add(24 * time.Hour)
			}
			emptyBlockWorkAllHandlerCache = results
			emptyBlockWorkAllHandlerCacheLastBuilt = time.Now()
		}
		emptyBlockWorkAllHandlerCacheLock.Unlock()
	}

	writeJson(w, emptyBlockWorkAllHandlerCache)
}

func emptyBlockWorkOnDay(date time.Time) ([]EmptyBlockWorkResult, error) {
	results := []EmptyBlockWorkResult{}
	query := `SELECT 
					ebwk.observed_on,
					ebwk.pool_id, 
					l.name, 
					l.id,
					po.id,
					po.stratum_host,
					ebwk.coin_id,
					ec.name,
					ebwk.total_jobs,
					ebwk.empty_block_work_jobs,
					ebwk.total_time_msec,
					ebwk.empty_block_work_time_msec
				FROM 
					analysis_empty_block_work_daily ebwk
					LEFT JOIN pool_observers po ON po.id=ebwk.pool_observer_id
					LEFT JOIN locations l on l.id=ebwk.location_id
					LEFT JOIN coins ec ON ec.id=ebwk.coin_id
				WHERE 
					ebwk.observed_on = $1::date`

	rows, err := db.Query(query, date)
	if err != nil {
		return results, err
	}
	for rows.Next() {
		var res EmptyBlockWorkResult
		err := rows.Scan(&res.ObservedOn, &res.PoolID, &res.LocationName, &res.LocationID, &res.ObserverID, &res.StratumHost, &res.CoinID, &res.CoinName, &res.TotalJobs, &res.EmptyBlockJobs, &res.TotalTime, &res.EmptyBlockTime)
		if err != nil {
			return results, err
		}
		results = append(results, res)
	}
	return results, nil
}
