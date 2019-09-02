package main

import (
	"database/sql"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/mit-dci/pooldetective/dbqueries/queries"
	"github.com/mit-dci/pooldetective/logging"
	"github.com/mit-dci/pooldetective/util"
)

var db *sql.DB

func main() {
	var err error
	logging.SetLogLevel(util.GetLoglevelFromEnv())

	connStr := os.Getenv("PGSQL_CONNECTION")
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	runQueries()
}

func runQueries() {
	qs := queries.AllQueries()
	for {
		t := time.Now()
		for i := range qs {
			if qs[i].ShouldRunAt(t) {
				logging.Debugf("Running query %s", qs[i].Name())
				s := qs[i].SQL()
				_, err := db.Exec(s)
				if err != nil {
					logging.Errorf("Failed to execute query: %s", err.Error())
					_, err := db.Exec("ROLLBACK;")
					if err != nil {
						logging.Errorf("Could not rollback: %s", err.Error())
					}
				} else {
					qs[i].RanAt(t)
					logging.Debugf("Completed query %s", qs[i].Name())
				}

			}
		}
		time.Sleep(time.Second)
	}
}
