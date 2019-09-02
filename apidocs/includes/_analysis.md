# Analysis

## Public endpoints:

* Get wrong work from yesterday - [`/api/public/wrongwork/yesterday`](#get-wrong-work-that-we-discovered-yesterday)
* Get unresolved work from yesterday - [`/api/public/unresolvedwork/yesterday`](#get-unresolved-work-that-we-discovered-yesterday)

## Get Wrong Work that we discovered on a specific date

```shell
curl "https://pooldetective.org/api/wrongwork/2020-09-01" \
  -H "X-Api-Key: myapikey"
```

> The above command returns JSON structured like this:

```json
[
  {
    "observedOn": "2020-09-10T00:00:00Z",
    "poolId": 2,
    "expectedCoinId": 3,
    "expectedCoinName": "Bitcoin Cash",
    "stratumHost": "stratum.antpool.com",
    "location": "Netherlands",
    "wrongCoinId": 1,
    "wrongCoinName": "Bitcoin",
    "totalJobs": 2581,
    "wrongJobs": 169,
    "totalTimeMs": 86408050,
    "wrongTimeMs": 77097
  },
  {
    "observedOn": "2020-09-10T00:00:00Z",
    "poolId": 7,
    "expectedCoinId": 1,
    "expectedCoinName": "Bitcoin",
    "stratumHost": "stratum.btc.top",
    "location": "Netherlands",
    "wrongCoinId": 2,
    "wrongCoinName": "Bitcoin SV",
    "totalJobs": 3163,
    "wrongJobs": 13,
    "totalTimeMs": 86407440,
    "wrongTimeMs": 3
  }
]
```

This endpoint retrieves the wrong work being given by pools. When we connect to a pool, we expect the work to be for coin `expectedCoinId`. But if we receive other work (that we can resolve to another blockchain), we are classifying this as wrong. This endpoint will show the coin we actually got work for in `wrongCoinId`. It will also show the share of wrong work in both number of jobs (we got `wrongJobs` jobs of the wrong coin, and `totalJobs` in total from the pool during the entire day yesterday) as well as in time. We calculate the time for a single job to be the time we received it until we received the next one from the same pool connection. `totalTimeMs` is the total time spent on jobs during the day (usually this is 86400 seconds in 24 hours). `wrongTimeMs` is the time we would be able to spend working on the wrong jobs if we switched to them instantly, measured in milliseconds.

### HTTP Request

`GET https://pooldetective.org/api/wrongwork/<DATE>`

### URL Parameters

Parameter | Description
--------- | -----------
DATE | The date you want the data from in `YYYY-MM-DD` format

## Get Unresolved work that we discovered on a specific date

```shell
curl "https://pooldetective.org/api/unresolvedwork/2020-09-01" \
  -H "X-Api-Key: myapikey"
```

> The above command returns JSON structured like this:

```json
[
  {
    "observedOn": "2020-09-10T00:00:00Z",
    "poolId": 2,
    "expectedCoinId": 8,
    "expectedCoinName": "Dash",
    "stratumHost": "stratum-dash.antpool.com",
    "location": "Netherlands",
    "wrongCoinId": -1,
    "wrongCoinName": "No Coin",
    "totalJobs": 2284,
    "wrongJobs": 2,
    "totalTimeMs": 86383770,
    "wrongTimeMs": 100008
  },
  {
    "observedOn": "2020-09-10T00:00:00Z",
    "poolId": 3,
    "expectedCoinId": 8,
    "expectedCoinName": "Dash",
    "stratumHost": "dash.f2pool.com",
    "location": "Netherlands",
    "wrongCoinId": -1,
    "wrongCoinName": "No Coin",
    "totalJobs": 11911,
    "wrongJobs": 1,
    "totalTimeMs": 86392261,
    "wrongTimeMs": 2328
  }
]
```

This endpoint retrieves the unresolved work being given by pools. When we connect to a pool, we expect the work to be for coin `expectedCoinId`. But if we receive work we cannot match to any of our known blocks, we call this unresolved. Unresolved work is expected to generally be work on blocks that were orphaned and that our full nodes never received, so we don't know about their existence. It could also mean that this is work done on blocks from another chain we do not happen to monitor. It will also show the share of unresolved work in both number of jobs (we got `wrongJobs` jobs we couldn't resolve, and `totalJobs` in total from the pool during the entire day yesterday) as well as in time. We calculate the time for a single job to be the time we received it until we received the next one from the same pool connection. `totalTimeMs` is the total time spent on jobs during the day (usually this is 86400 seconds in 24 hours). `wrongTimeMs` is the time we would be able to spend working on the unresolved jobs if we switched to them instantly, measured in milliseconds.

### HTTP Request

`GET https://pooldetective.org/api/unresolvedwork/<DATE>`

### URL Parameters

Parameter | Description
--------- | -----------
DATE | The date you want the data from in `YYYY-MM-DD` format


