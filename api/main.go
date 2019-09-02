package main

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/mit-dci/pooldetective/api/routes"
	"github.com/mit-dci/pooldetective/logging"
	"github.com/mit-dci/pooldetective/util"
	"github.com/rs/cors"
)

var db *sql.DB

func main() {
	logging.SetLogLevel(util.GetLoglevelFromEnv())

	var err error
	connStr := os.Getenv("PGSQL_CONNECTION")
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()

	var corsOpt = cors.New(cors.Options{
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowCredentials: true,
	})

	routes.DefineRoutes(r, db)

	srv := &http.Server{
		Handler: corsOpt.Handler(r),
		Addr:    ":8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logging.Infoln("Starting listening...")
	logging.Fatal(srv.ListenAndServe())
}
