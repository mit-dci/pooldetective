# Data model of Pool Detective Postgres database

## algorithms

Table containing the different proof-of-work algorithms that are in use by coins tracked by the PoolDetective

| field          | type          | description                                                                             |
| -------------- | ------------- |---------------------------------------------------------------------------------------- |
| id | int    |  ID of the algorithm (Primary Key) |
| name | varchar(20) | Name of the algorithm |


## block_coinbase_merkleproofs

Table containing the merkle proofs for the coinbase transaction as seen in full blocks observed by our full nodes. Used for matching jobs (that also contain these merkle proofs) to blocks

| field          | type          | description                                                                             |
| -------------- | ------------- |---------------------------------------------------------------------------------------- |
| id | bigint | Unique ID (Primary Key) |
| block_id | bigint | Foreign key to blocks.id, refers to the block this is the coinbase merkleproof for |
| coinbase_merklebranches | []bytea | Merkle proof for the coinbase transaction in the referred block |

## block_coinbase_outputs

Table containing the output scripts and values of blocks' coinbase transactions. Used for matching jobs to blocks

| field          | type          | description                                                                             |
| -------------- | ------------- |---------------------------------------------------------------------------------------- |
| id | bigint | Unique ID (Primary Key) |
| block_id | bigint | Foreign key to blocks.id, refers to the block this is a coinbase output for |
| output_script | bytea | Script in the transaction's output |
| value | bigint | Value of the output in satoshi |
| output_index | int | The index of the output in the coinbase transaction |

## block_observations 

Table containing the observed blocks on the peer-to-peer network. These are observed by listening for `inv` messages from peers on the network, which indicate that that peer is telling our observer that they have a particular block. We record the peer and the locally observed timestamp along with the hash of the block (which is inserted into the `blocks` table and referred here)

| field          | type          | description                                                                             |
| -------------- | ------------- |---------------------------------------------------------------------------------------- |
| id | bigint | Unique ID (Primary Key) |
| block_id | bigint | Foreign key to blocks.id, refers to the block we observed |
| location_id | int | Foreign key to locations.id, refers to where the observer is running |
| observed | timestamp | Locally observed timestamp in UTC zone for the `inv` message |
| peer_ip | inet | The IP address of the peer we got the `inv` message from |
| peer_port | int | The port of the peer we got the `inv` message from |
| frommongo | bool | Old field used for importing the data from the previously used mongo DB |

## blocks

Table containing block header data from blocks both validated by our full node (indicated by the record having a `height` that is not NULL) or by one of the block observers

| field          | type          | description                                                                             |
| -------------- | ------------- |---------------------------------------------------------------------------------------- |
| id | bigint | Unique ID (Primary Key) |
| coin_id | int | Foreign key to `coins.id`, identifying which coin(chain) this block belongs to |
| block_hash | bytea | Hash of the block |
| previous_block_hash | bytea | Hash of the block prior to this block in the blockchain |
| height | int | The height of the block in the blockchain |
| merkle_root | bytea | The merkle root of all transaction hashes included in this block |
| timestamp | int | The unix epoch of when this block was found (controlled by the miners) |

## coins

Table containing all coins tracked by the pooldetective


| field          | type          | description                                                                             |
| -------------- | ------------- |---------------------------------------------------------------------------------------- |
| id | bigint | Unique ID (Primary Key) |
| name | varchar(64) | Name of the coin |
| ticker | varchar(5) | Shorthand ticker symbol for the coin | 
| algorithm_id | int | Foreign key to `algorithms.id` indicating the PoW algorithm used by the coin |


## job_coinbase_outputs

Table containing the output scripts and values of jobs' coinbase transactions. Used for matching jobs to blocks

| field          | type          | description                                                                             |
| -------------- | ------------- |---------------------------------------------------------------------------------------- |
| id | bigint | Unique ID (Primary Key) |
| job_id | bigint | Foreign key to jobs.id, refers to the job this is a coinbase output for |
| output_script | bytea | Script in the transaction's output |
| value | bigint | Value of the output in satoshi |
| output_index | int | The index of the output in the coinbase transaction |

## jobs 

Table containing the job data received over stratum. For more info on stratum's job fields (`mining.nofity message`), see [this wiki](https://en.bitcoin.it/wiki/Stratum_mining_protocol#mining.notify)

| field          | type          | description                                                                             |
| -------------- | ------------- |---------------------------------------------------------------------------------------- |
| id | bigint | Unique ID (Primary Key) |
| observed | timestamp | Timestamp indicating when the job was received (local timestamp) | 
| pool_observer_id | int | Foreign key to `pool_observers.id` indiciating which stratum connection we received the job over |
| pool_job_id | bytea | The ID the job has at the pool |
| previous_block_hash | bytea | The blockhash the job is supposed to build on top of |
| previous_block_id | int | A convenience field to store/cache the lookup from block_hash to block_id |
| generation_transaction_part_1 | bytea | The first part of the coinbase transaction included in the job |
| generation_transaction_part_2 | bytea | The second part of the coinbase transaction included in the job |
| merkle_branches | bytea | The merkle branches for the coinbase transaction included in the job | 
| block_version | bytea  | the block version field included in the job |
| difficulty_bits | bytea | The difficulty bits field included in the job | 
| clean_jobs | boolean | The clean_jobs field included in the job | 
| timestamp | bigint | The timestamp field included in the job |
| frommongo | bool | Old field used for importing the data from the previously used mongo DB |
| tempid | bigint | Unused |
| next_job_id | bigint | Convenience field to cache the next job received from the same pool_observer |
| reserved | bytea | Field used in stratum messages from ZEC and BTG |
| time_spent_msec | bigint | Convenience field to cache the time between this job and next_job_id |

## locations

Table containing the different locations where PoolDetective software is running

| field          | type          | description                                                                             |
| -------------- | ------------- |---------------------------------------------------------------------------------------- |
| id | bigint | Unique ID (Primary Key) |
| name | varchar(32) | Name of the location | 
| reload_coordinator | bool | Update this field to `true` to force the coordinator in this location to reload |

## pool_observer_events

Table logging the events of pool_observers such as connect/disconnect/login success/login failure

| field          | type          | description                                                                             |
| -------------- | ------------- |---------------------------------------------------------------------------------------- |
| id | bigint | Unique ID (Primary Key) |
| pool_observer_id | bigint | Foreign key to `pool_observers.id` |
| event | int | Numeric indicator for the type of event that occurred |
| timestamp | timestamp | When did the event occur |


## pool_observers

Table containing the stratum connections used by pool detective from various locations to observe the pool work

| field          | type          | description                                                                             |
| -------------- | ------------- |---------------------------------------------------------------------------------------- |
| id | bigint | Unique ID (Primary Key) |
| frommongo | bool | Old field used for importing the data from the previously used mongo DB |
| disabled | bool | Indicates if this observer is currently disabled |
| location_id | int | Foreign key to locations.id, refers to where the pool_observer is running |
| pool_id | int | Foreign key to pools.id, refering to the pool that's being monitored |
| coin_id | int | Foreign key to coins.id, refering to the coin that's being monitored | 
| stratum_host | varchar() | The host we're connecting to |
| stratum_port | varchar() | The port we're connecting to |
| stratum_username | varchar() | The username being used to connect |
| stratum_password | varchar() | The password being used to connect |
| stratum_difficulty | float | The current difficulty for this stratum connection |
| stratum_extranonce1 | bytea | The current extranonce1 used for the stratum connection | 
| stratum_extranonce2size | int | The current extranonce2_size used for this stratum connection |
| last_job_received | timestamp | Time when the last job was received from this pool connection |
| last_share_found | timestamp | Time when the last share was found for this pool connection |
| last_job_id | bigint | Foreign key to jobs.id for the last job received from this pool connection |
| last_share_id | bigint | Foreign key to shares.id for the last share found for this pool connection |
| stratum_protocol | int | 0 = normal, 1 = ZEC/BTG |
| stratum_target | bytea | for stratum protocol 1, uses target in stead of difficulty |

## pools 

Table containing the different pools we're monitoring 

| field          | type          | description                                                                             |
| -------------- | ------------- |---------------------------------------------------------------------------------------- |
| id | bigint | Unique ID (Primary Key) |
| name | varchar() | Name of the pool |

## shares 

Table containing the shares we found using mining 

| field          | type          | description                                                                             |
| -------------- | ------------- |---------------------------------------------------------------------------------------- |
| id | bigint | Unique ID (Primary Key) |
| job_id | bigint | Foreign key to jobs.id refering to the job this share was found for |
| extranonce2 | bytea | The value of extranonce2 for this share |
| timestamp | int | The value of timestamp for this share | 
| nonce | bigint | The value of the nonce for this share | 
| found | timestamp | The timestamp of when the share was found |
| submitted | timestamp | The timestamp of when the share was submitted to the pool |
| responsereceived | timestamp | The timestamp of when the response to submitting the share was received |
| accepted | bool | Wether the share was accepted by the server |
| details | varchar() | Any details about the share's submission (error code in case of failure) |
| stale | bool | Wether the share was found to be stale and not submitted to the upstream pool | 
| additional_solution_data | bytea | Additional share data required by BTG/ZEC |
| frommongo | bool | Old field used for importing the data from the previously used mongo DB |
| tempid | bigint | Unused |

## stratum_servers

Table containing the active stratum servers that are being ran in order for miners to do pool work

| field          | type          | description                                                                             |
| -------------- | ------------- |---------------------------------------------------------------------------------------- |
| id | bigint | Unique ID (Primary Key) |
| algorithm_id | int | Foreign key to `algorithms.id` indicating the PoW algorithm used by the stratum server |
| stratum_protocol | int | 0 = normal, 1 = ZEC/BTG |
| location_id | int | Foreign key to locations.id, refers to where the stratum server should be running |
| port | int | The port on which to run the stratum server |


