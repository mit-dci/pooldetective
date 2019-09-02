#!/bin/bash
BITCOIN_DIR=/root/.komodo 

# If config doesn't exist, initialize with sane defaults for running a
# non-mining node.
if [ ! -e "/root/.komodo/komodo.conf" ]; then
  fetch-params.sh
fi


if [ $# -eq 0 ]; then
  exec komodod -printtoconsole -ac_name=PIRATE -ac_supply=0 -ac_reward=25600000000 -ac_halving=77777 -ac_private=1 -addnode=136.243.102.225 -server=1 -showmetrics=0
else
  exec "$@"
fi
