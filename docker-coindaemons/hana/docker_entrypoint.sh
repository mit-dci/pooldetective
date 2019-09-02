#!/bin/bash
COIN=hanacoin
COIN_RPCPORT=9502
COIN_DIR=/root/.${COIN}
COIN_CONF=${COIN_DIR}/${COIN}.conf

# If config doesn't exist, initialize with sane defaults for running a
# non-mining node.
if [ ! -e "${COIN_CONF}" ]; then
  tee -a >${COIN_CONF} <<EOF

server=1
rpcuser=pooldetective
rpcpassword=pooldetective
rpcclienttimeout=30
rpcallowip=0.0.0.0/0
rpcport=${COIN_RPCPORT}
rpcbind=0.0.0.0
printtoconsole=1
disablewallet=1
txindex=1
testnet=0
dbcache=512
prune=0
EOF
fi

if [ $# -eq 0 ]; then
  exec "${COIN}d" -datadir=${COIN_DIR} -conf=${COIN_CONF}
else
  exec "$@"
fi
