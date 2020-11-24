package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/mit-dci/pooldetective/logging"
	"github.com/mit-dci/pooldetective/util"
)

var db *sql.DB

type Coin struct {
	ID          int64
	CoinGeckoID string
}

type CoinGeckoResponse struct {
	MarketData CoinGeckoMarketDataResponse `json:"market_data"`
}

type CoinGeckoMarketDataResponse struct {
	CurrentPrice CoinGeckoMarketDataCurrentPriceResponse `json:"current_price"`
}

type CoinGeckoMarketDataCurrentPriceResponse struct {
	USD float64 `json:"usd"`
}

func main() {
	logging.SetLogLevel(util.GetLoglevelFromEnv())

	var err error
	db, err = sql.Open("postgres", os.Getenv("PGSQL_CONNECTION"))
	if err != nil {
		logging.Fatal(err)
	}

	var coins []Coin

	rows, err := db.Query("SELECT id, coingecko_id FROM coins WHERE coingecko_id is not null")
	if err != nil {
		logging.Fatal(err)
	}

	for rows.Next() {
		coin := Coin{}

		err = rows.Scan(&coin.ID, &coin.CoinGeckoID)
		if err != nil {
			logging.Fatal(err)
		}
		coins = append(coins, coin)
	}

	for {
		for _, c := range coins {
			var maxDate time.Time
			err := db.QueryRow("SELECT COALESCE(max(time),'2019-06-30') FROM coin_pricehistory WHERE coin_id=$1", c.ID).Scan(&maxDate)
			if err != nil {
				logging.Fatal(err)
			}
			errors := 0
			maxDate = maxDate.Add(time.Hour * 24)
			for date := maxDate; date.Before(time.Now()); date = date.Add(time.Hour * 24) {
				var resp CoinGeckoResponse
				err = GetJson(fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s/history?date=%s&localization=false", c.CoinGeckoID, date.Format("02-01-2006")), &resp)
				if err != nil {
					errors++
					if errors > 5 {
						logging.Fatal(err)
					}
					date = date.Add(time.Hour * 24)
					time.Sleep(time.Second * 5)
					continue

				}
				errors = 0

				_, err = db.Exec("INSERT INTO coin_pricehistory(coin_id, time, price_usd) VALUES ($1, $2, $3)", c.ID, date, resp.MarketData.CurrentPrice.USD)
				if err != nil {
					logging.Fatal(err)
				}

				time.Sleep(time.Millisecond * 250)
			}
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
