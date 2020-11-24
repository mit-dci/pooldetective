package routes

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/mit-dci/pooldetective/logging"
)

type WrongWorkResult struct {
	ObservedOn    time.Time `json:"observedOn"`
	PoolID        int       `json:"poolId"`
	CoinID        int       `json:"expectedCoinId"`
	CoinName      string    `json:"expectedCoinName"`
	ObserverID    int       `json:"poolObserverID"`
	StratumHost   string    `json:"stratumHost"`
	LocationID    int       `json:"locationID"`
	LocationName  string    `json:"location"`
	WrongCoinID   int       `json:"wrongCoinId"`
	WrongCoinName string    `json:"wrongCoinName"`
	TotalJobs     int64     `json:"totalJobs"`
	WrongJobs     int64     `json:"wrongJobs"`
	TotalTime     int64     `json:"totalTimeMs"`
	WrongTime     int64     `json:"wrongTimeMs"`
}

var wrongWorkYesterdayHandlerCache []WrongWorkResult
var wrongWorkYesterdayHandlerCacheLock sync.Mutex = sync.Mutex{}
var wrongWorkYesterdayHandlerCacheLastBuilt time.Time = time.Now().Add(-24 * time.Hour)

func wrongWorkYesterdayHandler(w http.ResponseWriter, r *http.Request) {
	if wrongWorkYesterdayHandlerCacheLastBuilt.Day() != time.Now().Day() {
		wrongWorkYesterdayHandlerCacheLock.Lock()
		if wrongWorkYesterdayHandlerCacheLastBuilt.Day() != time.Now().Day() {
			results, err := wrongWorkOnDay(time.Now().Add(-24*time.Hour), false)
			if err != nil {
				logging.Errorf("Error: %s", err.Error())
				http.Error(w, "Internal server error", 500)
				wrongWorkYesterdayHandlerCacheLock.Unlock()
				return
			}
			wrongWorkYesterdayHandlerCache = results
			wrongWorkYesterdayHandlerCacheLastBuilt = time.Now()
		}
		wrongWorkYesterdayHandlerCacheLock.Unlock()
	}

	writeJson(w, wrongWorkYesterdayHandlerCache)
}

var wrongWorkLastWeekHandlerCache []WrongWorkResult
var wrongWorkLastWeekHandlerCacheLock sync.Mutex = sync.Mutex{}
var wrongWorkLastWeekHandlerCacheLastBuilt time.Time = time.Now().Add(-24 * time.Hour)

func wrongWorkLastWeekHandler(w http.ResponseWriter, r *http.Request) {
	if wrongWorkLastWeekHandlerCacheLastBuilt.Day() != time.Now().Day() {
		wrongWorkLastWeekHandlerCacheLock.Lock()
		if wrongWorkLastWeekHandlerCacheLastBuilt.Day() != time.Now().Day() {
			results := []WrongWorkResult{}
			t := time.Now().Add(-24 * 7 * time.Hour)
			for t.Day() != time.Now().Day() {
				resultsForDay, err := wrongWorkOnDay(t, false)
				if err != nil {
					logging.Errorf("Error: %s", err.Error())
					http.Error(w, "Internal server error", 500)
					wrongWorkYesterdayHandlerCacheLock.Unlock()
					return
				}
				results = append(results, resultsForDay...)
				t = t.Add(24 * time.Hour)
			}
			wrongWorkLastWeekHandlerCache = results
			wrongWorkLastWeekHandlerCacheLastBuilt = time.Now()
		}
		wrongWorkLastWeekHandlerCacheLock.Unlock()
	}

	writeJson(w, wrongWorkLastWeekHandlerCache)
}

var wrongWorkAllHandlerCache []WrongWorkResult
var wrongWorkAllHandlerCacheLock sync.Mutex = sync.Mutex{}
var wrongWorkAllHandlerCacheLastBuilt time.Time = time.Now().Add(-24 * time.Hour)

func wrongWorkAllHandler(w http.ResponseWriter, r *http.Request) {
	if wrongWorkAllHandlerCacheLastBuilt.Day() != time.Now().Day() {
		wrongWorkAllHandlerCacheLock.Lock()
		if wrongWorkAllHandlerCacheLastBuilt.Day() != time.Now().Day() {
			results := []WrongWorkResult{}
			t := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
			for t.Before(time.Now()) {
				resultsForDay, err := wrongWorkOnDay(t, false)
				if err != nil {
					logging.Errorf("Error: %s", err.Error())
					http.Error(w, "Internal server error", 500)
					wrongWorkAllHandlerCacheLock.Unlock()
					return
				}
				results = append(results, resultsForDay...)
				t = t.Add(24 * time.Hour)
			}
			wrongWorkAllHandlerCache = results
			wrongWorkAllHandlerCacheLastBuilt = time.Now()
		}
		wrongWorkAllHandlerCacheLock.Unlock()
	}

	writeJson(w, wrongWorkAllHandlerCache)
}

func wrongWorkOnDateHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	dateString := params["date"]

	date, err := time.Parse("2006-01-02", dateString)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid date: %s", dateString), 500)
		return
	}

	results, err := wrongWorkOnDay(date, false)
	if err != nil {
		logging.Errorf("Error: %s", err.Error())
		http.Error(w, "Internal server error", 500)
		return
	}
	writeJson(w, results)
}

var unresolvedWorkYesterdayHandlerCache []WrongWorkResult
var unresolvedWorkYesterdayHandlerCacheLock sync.Mutex = sync.Mutex{}
var unresolvedWorkYesterdayHandlerCacheLastBuilt time.Time = time.Now().Add(-24 * time.Hour)

func unresolvedWorkYesterdayHandler(w http.ResponseWriter, r *http.Request) {
	if unresolvedWorkYesterdayHandlerCacheLastBuilt.Day() != time.Now().Day() {
		unresolvedWorkYesterdayHandlerCacheLock.Lock()
		if unresolvedWorkYesterdayHandlerCacheLastBuilt.Day() != time.Now().Day() {
			results, err := wrongWorkOnDay(time.Now().Add(-24*time.Hour), true)
			if err != nil {
				logging.Errorf("Error: %s", err.Error())
				http.Error(w, "Internal server error", 500)
				unresolvedWorkYesterdayHandlerCacheLock.Unlock()
				return
			}
			unresolvedWorkYesterdayHandlerCache = results
			unresolvedWorkYesterdayHandlerCacheLastBuilt = time.Now()
		}
		unresolvedWorkYesterdayHandlerCacheLock.Unlock()
	}

	writeJson(w, unresolvedWorkYesterdayHandlerCache)
}

func unresolvedWorkOnDateHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	dateString := params["date"]

	date, err := time.Parse("2006-01-02", dateString)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid date: %s", dateString), 500)
		return
	}

	results, err := wrongWorkOnDay(date, true)
	if err != nil {
		logging.Errorf("Error: %s", err.Error())
		http.Error(w, "Internal server error", 500)
		return
	}
	writeJson(w, results)
}

func wrongWorkOnDay(date time.Time, unresolved bool) ([]WrongWorkResult, error) {
	results := []WrongWorkResult{}
	query := `SELECT 
					awk.observed_on,
					awk.pool_id, 
					l.name, 
					l.id,
					po.id,
					po.stratum_host,
					awk.expected_coin_id,
					ec.name,
					awk.got_coin_id,
					gc.name,
					awk.total_jobs,
					awk.wrong_jobs,
					awk.total_time_msec,
					awk.wrong_time_msec
				FROM 
					analysis_wrong_work_daily awk
					LEFT JOIN pool_observers po ON po.id=awk.pool_observer_id
					LEFT JOIN locations l on l.id=awk.location_id
					LEFT JOIN coins ec ON ec.id=awk.expected_coin_id
					LEFT JOIN coins gc ON gc.id=awk.got_coin_id
				WHERE 
					awk.observed_on = $1::date
					AND awk.got_coin_id `
	if unresolved {
		query += "="
	} else {
		query += "!="
	}
	query += " -1"

	rows, err := db.Query(query, date)
	if err != nil {
		return results, err
	}
	for rows.Next() {
		var res WrongWorkResult
		err := rows.Scan(&res.ObservedOn, &res.PoolID, &res.LocationName, &res.LocationID, &res.ObserverID, &res.StratumHost, &res.CoinID, &res.CoinName, &res.WrongCoinID, &res.WrongCoinName, &res.TotalJobs, &res.WrongJobs, &res.TotalTime, &res.WrongTime)
		if err != nil {
			return results, err
		}
		results = append(results, res)
	}
	return results, nil
}
