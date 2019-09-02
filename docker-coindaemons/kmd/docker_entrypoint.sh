#!/bin/bash
BITCOIN_DIR=/root/.komodo 

# If config doesn't exist, initialize with sane defaults for running a
# non-mining node.
if [ ! -e "/root/.komodo/komodo.conf" ]; then
  fetch-params.sh
  touch /root/.komodo/komodo.conf
fi


if [ $# -eq 0 ]; then
  exec komodod -printtoconsole -server=1
else
  exec "$@"
fi
