#!/bin/bash
BITCOIN_DIR=/root/.vertcoin
BITCOIN_CONF=${BITCOIN_DIR}/vertcoin.conf

# If config doesn't exist, initialize with sane defaults for running a
# non-mining node.
if [ ! -e "${BITCOIN_CONF}" ]; then
  tee -a >${BITCOIN_CONF} <<EOF

server=1
rpcuser=pooldetective
rpcpassword=pooldetective
rpcclienttimeout=30
rpcallowip=0.0.0.0/0
rpcport=5888
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
  exec vertcoind -datadir=${BITCOIN_DIR} -conf=${BITCOIN_CONF}
else
  exec "$@"
fi
