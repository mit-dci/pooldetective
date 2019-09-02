# Pool Detective

Pool Detective is a system we built at the [Digital Currency Initiative](https://dci.mit.edu/), and are currently running, to monitor the behavior of mining pools that operate on Proof-of-Work cryptocurrencies such as Bitcoin, Litecoin and others. Mining pools have ultimate control over the work that constituent miners process and therefore their (mis)behavior can have large consequences for the security of Proof-of-Work networks. We're conducting this research because we think it's important to perform detailed monitoring and analyze the behavior of pools, and no one else is doing that up to this level of detail.

More on PoolDetective can be seen on [our website](https://pooldetective.org/) - we're also exposing an [API](https://pooldetective.org/api/docs)

## Running PoolDetective yourself

If you want to, you can run PoolDetective yourself by the use of this open-source repository. There's a few moving parts to it, but it's made pretty simple to run using [Docker](https://www.docker.com/).

### Prerequisites

Before you can run PoolDetective, you'll have to install Docker. This is only tested on Ubuntu. Other linux flavors with Docker should work just fine. MacOS or Windows with Docker might not. 

In this guide, we're assuming you're running this on Linux (Ubuntu).

### Clone repository

Once you have Docker installed, clone this repository somewhere:

```
git clone https://github.com/mit-dci/pooldetective
```

### Simple setup script

From the root of the repository, run `docker-setup.sh` - this will setup a pre-made demo environment including:

* Will build all of the Pool Detective Docker containers
* Will build dockerized coin daemons for Bitcoin and Litecoin
* Will start the Postgres database
* Will import the Pool Detective database schema
* Will import basic demo data including two pool observers for a Bitcoin and Litecoin pool
* Will start all components of the PoolDetective, including the frontend and api 

### Adding more pools, coins, algorithms

The [Demo Data SQL Script](database/demodata.sql) is an example of how to add more coins, algorithms and pools to the database. For full reference, refer to the [Data model](database/data-model.md). Once you have added more pools/coins/etc, restart the `coordinator` Docker container to force reloading the configuration.

### Adding more coin daemons

If you want to run more of the coin daemons yourself, allowing you to import block data into the Pool Detective database, you can build the coin daemon using the `build.sh` script in the `docker-coindaemons` folder. For instance, building Bitcoin Cash would mean running `build.sh bch` from that folder. 

Then, look at the `coind-btc` definition in the [`localdocker/docker-compose.yml`](localdocker/docker-compose.yml) file to understand how to run additional coins, and map the data directory of the coin to a folder on the host to preserve block data in the full node.

Adjust the environment variables for the `blockfetcher` container as well to add your new coindaemon ensuring it will be indexed by the blockfetcher as well.

### Advanced use cases

With some knowledge of Docker and the `docker-compose.yml` file in the `localdocker` folder, you should be able to figure out how to integrate PoolDetective into an existing docker setup or a more production-ready set up.

# Questions, comments, ideas

Feel free to leave your feedback by [creating a new issue in this repository](https://github.com/mit-dci/pooldetective/issues/new) 

