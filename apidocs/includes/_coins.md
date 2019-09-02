# Coins

## Public endpoints:

* Get all coins - [`/api/public/coins`](#get-all-coins)
* Get pools for a coin - [`/api/public/coins/<ID>/pools`](#get-the-pools-that-we-are-mining-a-specific-coin-on)

## Get a specific coin

```shell
curl "https://pooldetective.org/api/coins/2" \
  -H "X-Api-Key: myapikey"
```

> The above command returns JSON structured like this:

```json
{
  "id": 2,
  "minHeight": 0,
  "maxHeight": 627866,
  "numBlocks": 627874,
  "tipHash": "0000000000000000012977bd1bf57513e8891c482bc9ba09d6b511219583842d",
  "algorithmName": "SHA-256",
  "algorithmID": 1
}
```

This endpoint retrieves details for a specific coin

<aside class="notice">
This API requires authentication with an API Key
</aside>

### HTTP Request

`GET https://pooldetective.org/api/coins/<ID>`

### URL Parameters

Parameter | Description
--------- | -----------
ID | The ID of the coin to retrieve



## Get the details about a coins we are mining on a pool

```shell
curl "https://pooldetective.org/api/coins/1/pools/2" \
  -H "X-Api-Key: myapikey"
```

> The above command returns JSON structured like this:

```json
{
  "observers": [
    {
      "id": 17,
      "location": "Singapore",
      "poolServer": "stratum.antpool.com",
      "lastJobReceived": "2020-03-24T11:19:39.705958Z",
      "lastShareFound": "2020-03-24T11:30:58.915025Z",
      "currentDifficulty": 512
    },
    {
      "id": 10,
      "location": "Netherlands",
      "poolServer": "stratum.antpool.com",
      "lastJobReceived": "2020-03-26T14:20:10.953866Z",
      "lastShareFound": "2020-03-26T14:20:06.191896Z",
      "currentDifficulty": 512
    },
    {
      "id": 24,
      "location": "United States West",
      "poolServer": "stratum.antpool.com",
      "lastJobReceived": "2020-03-26T14:20:05.633823Z",
      "lastShareFound": "2020-03-26T14:20:17.824093Z",
      "currentDifficulty": 512
    },
    {
      "id": 3,
      "location": "Australia",
      "poolServer": "stratum.antpool.com",
      "lastJobReceived": "2020-03-26T14:20:06.063283Z",
      "lastShareFound": "2020-03-26T14:20:19.883102Z",
      "currentDifficulty": 512
    },
    {
      "id": 31,
      "location": "United States East",
      "poolServer": "stratum.antpool.com",
      "lastJobReceived": "2020-03-25T06:25:28.84478Z",
      "lastShareFound": "2020-03-25T06:40:26.719632Z",
      "currentDifficulty": 512
    }
  ]
}
```

This endpoint retrieves details about how we are currently mining the given coin at the given pool. It returns details about the last time we got a job from the pool, and the last time we found a valid proof-of-work share. It also returns the current mining difficulty on this pool.

### HTTP Request

`GET https://pooldetective.org/api/coins/<CoinID>/pools/<PoolID>`

### URL Parameters

Parameter | Description
--------- | -----------
CoinID | The ID of the coin to retrieve details for
PoolID | The ID of the pool we want to retrieve mining details for about mining coin CoinID

