# Public API

The Public API does not require an API key. You can retrieve this data without authenticating, but you are limited to 12 requests per minute.

If you have an API key, you can call these same APIs by removing the `/public` portion of the URL and supplying your API-Key with the request. You will then not be limited to 12 requests per minute, but 120.

## Get All Coins

This endpoint retrieves all the coins the Pool Detective is tracking.

### HTTP Request

`GET https://pooldetective.org/api/public/coins`

```shell
curl "https://pooldetective.org/api/public/coins"
```

> The above command returns JSON structured like this:

```json
[
    {
        "id":1,
        "name":"Bitcoin"
    },
    {
        "id":2,
        "name":"Bitcoin SV"
    },
    {
        "id":3,
        "name":"Bitcoin Cash"
    }
]
```

## Get the pools that we are mining a specific coin on

```shell
curl "https://pooldetective.org/api/public/coins/2/pools"
```

> The above command returns JSON structured like this:

```json
[
  {
    "id": 1,
    "name": "Slushpool"
  },
  {
    "id": 2,
    "name": "Antpool"
  },
  {
    "id": 3,
    "name": "F2Pool"
  }
]
```

This endpoint retrieves which pools we are using to mine a specific coin

### HTTP Request

`GET https://pooldetective.org/api/public/coins/<ID>/pools`

### URL Parameters

Parameter | Description
--------- | -----------
ID | The ID of the coin to retrieve the pools for

## Get All Pools

```shell
curl "https://pooldetective.org/api/public/pools"
```

> The above command returns JSON structured like this:

```json
[
  {
    "id": 1,
    "name": "Slushpool"
  },
  {
    "id": 2,
    "name": "Antpool"
  },
  {
    "id": 3,
    "name": "F2Pool"
  }
]
```

This endpoint retrieves all the pools the Pool Detective is tracking.

### HTTP Request

`GET https://pooldetective.org/api/public/pools`


## Get the coins we are mining on a pool

```shell
curl "https://pooldetective.org/api/public/pools/1/coins"
```

> The above command returns JSON structured like this:

```json
[
  {
    "id": 1,
    "name": "Bitcoin"
  },
  {
    "id": 3,
    "name": "Bitcoin Cash"
  },
  {
    "id": 5,
    "name": "Litecoin"
  }
]
```

This endpoint retrieves which coins we are mining on the given pool.

### HTTP Request

`GET https://pooldetective.org/api/public/pools/<ID>/coins`

### URL Parameters

Parameter | Description
--------- | -----------
ID | The ID of the pool to retrieve the coins for

## Get Wrong Work that we discovered yesterday

```shell
curl "https://pooldetective.org/api/public/wrongwork/yesterday"
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

`GET https://pooldetective.org/api/public/wrongwork/yesterday`


## Get Unresolved work that we discovered yesterday

```shell
curl "https://pooldetective.org/api/public/unresolvedwork/yesterday"
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

`GET https://pooldetective.org/api/public/unresolvedwork/yesterday`