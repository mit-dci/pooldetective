package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/mit-dci/pooldetective/logging"
	"github.com/mit-dci/pooldetective/util"
)

var db *sql.DB

type NicehashPrice struct {
	Time          time.Time
	AvailableHash float64
	Price         float64
}

type NicehashAlgorithmsResponse struct {
	MiningAlgorithms []NicehashAlgorithm `json:"miningAlgorithms"`
}

type NicehashAlgorithm struct {
	Algorithm    string `json:"algorithm"`
	MarketFactor string `json:"marketFactor"`
}

func main() {
	logging.SetLogLevel(util.GetLoglevelFromEnv())

	var err error
	db, err = sql.Open("postgres", os.Getenv("PGSQL_CONNECTION"))
	if err != nil {
		logging.Fatal(err)
	}

	var algos NicehashAlgorithmsResponse

	err = GetJson("https://api2.nicehash.com/main/api/v2/mining/algorithms", &algos)
	if err != nil {
		logging.Fatal(err)
	}
	for _, a := range algos.MiningAlgorithms {
		marketFactor, err := strconv.Atoi(a.MarketFactor)
		if err != nil {
			logging.Fatal(err)
		}
		_, err = db.Exec("UPDATE algorithms SET nicehash_marketfactor=$1 WHERE nicehash_algoid=$2", marketFactor, a.Algorithm)
		if err != nil {
			logging.Fatal(err)
		}
	}

	var prices [][]interface{}

	for {
		rows, err := db.Query("SELECT id, nicehash_algoid FROM algorithms WHERE nicehash_algoid is not null")
		if err != nil {
			logging.Fatal(err)
		}

		for rows.Next() {
			var algoID int
			var nicehashID string
			err = rows.Scan(&algoID, &nicehashID)
			if err != nil {
				logging.Fatal(err)
			}
			err = GetJson(fmt.Sprintf("https://api2.nicehash.com/main/api/v2/public/algo/history?algorithm=%s", nicehashID), &prices)
			if err != nil {
				logging.Fatal(err)
			}

			sort.Slice(prices, func(i, j int) bool {
				if prices[i][0].(float64) < prices[j][0].(float64) {
					return false
				}
				return true
			})

			nhPrices := make([]NicehashPrice, 0)

			var maxTime time.Time
			err := db.QueryRow("SELECT coalesce(max(time),'2019-01-01') FROM nicehash_pricehistory WHERE algorithm_id=$1", algoID).Scan(&maxTime)
			if err != nil {
				logging.Fatal(err)
			}

			for _, p := range prices {
				pr := NicehashPrice{
					Time:          time.Unix(int64(p[0].(float64)), 0),
					AvailableHash: p[1].(float64),
					Price:         p[2].(float64),
				}
				if pr.Time.Before(maxTime) {
					break
				}

				nhPrices = append(nhPrices, pr)
			}

			tx, err := db.Begin()
			if err != nil {
				panic(err)
			}

			stmt, err := tx.Prepare(pq.CopyIn("nicehash_pricehistory", "algorithm_id", "time", "available_hash", "price"))
			if err != nil {
				panic(err)
			}

			errors := 0
			for _, p := range nhPrices {
				_, err := stmt.Exec(algoID, p.Time, p.AvailableHash, p.Price)
				if err != nil {
					errors++
				}
			}

			_, err = stmt.Exec()
			if err != nil {
				panic(err)
			}

			err = stmt.Close()
			if err != nil {
				panic(err)
			}

			err = tx.Commit()
			if err != nil {
				panic(err)
			}

			fmt.Printf("Inserted %d nicehash prices - %d errors\n", len(nhPrices), errors)
		}
		time.Sleep(time.Hour * 2)
	}

}

var jsonClient = &http.Client{Timeout: 60 * time.Second}

func GetJson(url string, target interface{}) error {
	r, err := jsonClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}
