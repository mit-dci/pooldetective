# blockfetcher

This project contains the blockfetcher, which is responsible for connecting to full nodes over RPC, and retrieving block header data and storing it in the database. It also fetches full blocks to analyze the coinbase data (merkle proof and coinbase outputs) and insert that into the database (see `coinbase.go`).