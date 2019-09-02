---
title: Pool Detective by Digital Currency Initiative

language_tabs: # must be one of https://git.io/vQNgJ
  - shell: cURL

toc_footers:
  - <a href='#get-an-api-key'>Get an API Key</a>

includes:
  - public
  - coins
  - pools
  - analysis
  - errors

search: true
---

# Introduction

Welcome to the Pool Detective API! The Pool Detective is a project of the [Digital Currency Initiative](https://dci.mit.edu) at the [MIT Media Lab](https://media.mit.edu).

The Pool Detective tracks the behavior of the largest mining pools in the most popular Proof-of-Work cryptocurrencies. We compare the work sent to us by these pools against the block data observed by both official full nodes from the corresponding cryptocurrency, as well as blocks we observe being propagated using a custom observer on the peer-to-peer networks. 

## Publicly available endpoints

There are API endpoints that are public (start with `/public`) which you can freely request. Other endpoints require a valid API key.

## Rate limits

Unauthenticated requests are limited to 12 requests per minute per IP address.
Authenticated requests are limited to 120 requests per minute per IP address.

## Authentication

> To authenticate with an API key, use this code:

```shell
# With shell, you can just pass the correct header with each request
curl "api_endpoint_here"
  -H "X-Api-Key: myapikey"
```

> Make sure to replace `myapikey` with your API key.

Pool Detective uses API keys to regulate access to the API. Pool Detective expects for the API key to be included in all API requests to the server in a header that looks like the following:

`X-Api-Key: myapikey`

<aside class="notice">
You must replace <code>myapikey</code> with your personal API key.
</aside>

### Get an API Key

You can obtain an API key by sending an e-mail to `pooldetective@mit.edu`.

# Feedback, questions, ideas

Feel free to leave your feedback by [creating a new issue](https://github.com/mit-dci/pooldetective/issues/new) in the PoolDetective Github repository. This repository also contains the code underpinning the API and PoolDetective, and contains instructions on how to run it yourself.