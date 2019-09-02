package routes

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mit-dci/pooldetective/api/auth"
	limiter "github.com/ulule/limiter"
	stdlib "github.com/ulule/limiter/drivers/middleware/stdlib"
	smem "github.com/ulule/limiter/drivers/store/memory"
)

var db *sql.DB

func DefineRoutes(r *mux.Router, passedDb *sql.DB) {
	db = passedDb

	// Define a limit rate to 12 requests per minute for anonymous requests
	anonRate, err := limiter.NewRateFromFormatted("12-M")
	if err != nil {
		log.Fatal(err)
		return
	}
	anonStore := smem.NewStore()
	anonLimiter := limiter.New(anonStore, anonRate)
	anonLimiterMiddleware := stdlib.NewMiddleware(anonLimiter)

	// Define a limit rate to 120 requests per minute for authenticated requests
	authRate, err := limiter.NewRateFromFormatted("120-M")
	if err != nil {
		log.Fatal(err)
		return
	}
	authStore := smem.NewStore()
	authLimiter := limiter.New(authStore, authRate)
	authLimiterMiddleware := stdlib.NewMiddleware(authLimiter)

	// Public API
	r.Handle("/public/coins", anonLimiterMiddleware.Handler(http.HandlerFunc(coinsHandler)))
	r.Handle("/public/coins/{coinID}/pools", anonLimiterMiddleware.Handler(http.HandlerFunc(coinPoolsHandler)))
	r.Handle("/public/algorithms", anonLimiterMiddleware.Handler(http.HandlerFunc(algorithmsHandler)))
	r.Handle("/public/algorithms/{algorithmID}/pools", anonLimiterMiddleware.Handler(http.HandlerFunc(algorithmPoolsHandler)))
	r.Handle("/public/pools/{poolID}/algorithms", anonLimiterMiddleware.Handler(http.HandlerFunc(poolAlgorithmsHandler)))
	r.Handle("/public/pools/{poolID}/coins", anonLimiterMiddleware.Handler(http.HandlerFunc(poolCoinsHandler)))
	r.Handle("/public/pools/{poolID}", anonLimiterMiddleware.Handler(http.HandlerFunc(poolHandler)))
	r.Handle("/public/pools", anonLimiterMiddleware.Handler(http.HandlerFunc(poolsHandler)))
	r.Handle("/public/wrongwork/yesterday", anonLimiterMiddleware.Handler(http.HandlerFunc(wrongWorkYesterdayHandler)))
	r.Handle("/public/unresolvedwork/yesterday", anonLimiterMiddleware.Handler(http.HandlerFunc(unresolvedWorkYesterdayHandler)))

	// Authenticated API

	r.Handle("/coins/{coinID}/pools/{poolID}", authLimiterMiddleware.Handler(auth.Auth(coinPoolHandler)))
	r.Handle("/coins/{coinID}/pools", authLimiterMiddleware.Handler(auth.Auth(coinPoolsHandler)))
	r.Handle("/coins/{coinID}", authLimiterMiddleware.Handler(auth.Auth(coinHandler)))
	r.Handle("/coins", authLimiterMiddleware.Handler(auth.Auth(coinsHandler)))
	r.Handle("/algorithms", authLimiterMiddleware.Handler(auth.Auth(algorithmsHandler)))
	r.Handle("/algorithms/{algorithmID}/pools", authLimiterMiddleware.Handler(auth.Auth(algorithmPoolsHandler)))
	r.Handle("/algorithms/{algorithmID}/pools/{poolID}", authLimiterMiddleware.Handler(auth.Auth(algorithmPoolHandler)))
	r.Handle("/pools/{poolID}/coins/{coinID}", authLimiterMiddleware.Handler(auth.Auth(coinPoolHandler)))
	r.Handle("/pools/{poolID}/algorithms/{algorithmID}", authLimiterMiddleware.Handler(auth.Auth(algorithmPoolHandler)))
	r.Handle("/pools/{poolID}/algorithms", authLimiterMiddleware.Handler(auth.Auth(poolAlgorithmsHandler)))
	r.Handle("/pools/{poolID}/coins", authLimiterMiddleware.Handler(auth.Auth(poolCoinsHandler)))
	r.Handle("/pools/{poolID}", authLimiterMiddleware.Handler(auth.Auth(poolHandler)))
	r.Handle("/pools", authLimiterMiddleware.Handler(auth.Auth(poolsHandler)))
	r.Handle("/wrongwork/yesterday", authLimiterMiddleware.Handler(auth.Auth(wrongWorkYesterdayHandler)))
	r.Handle("/wrongwork/date/{date}", authLimiterMiddleware.Handler(auth.Auth(wrongWorkOnDateHandler)))
	r.Handle("/unresolvedwork/yesterday", authLimiterMiddleware.Handler(auth.Auth(unresolvedWorkYesterdayHandler)))
	r.Handle("/unresolvedwork/date/{date}", authLimiterMiddleware.Handler(auth.Auth(unresolvedWorkOnDateHandler)))
}
