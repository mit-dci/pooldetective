#!/bin/bash
BITCOIN_DIR=/root/.zcash 
BITCOIN_CONF=${BITCOIN_DIR}/zcash.conf

# If config doesn't exist, initialize with sane defaults for running a
# non-mining node.
if [ ! -e "${BITCOIN_CONF}" ]; then

  zcash-fetch-params

  tee -a >${BITCOIN_CONF} <<EOF

server=1
rpcuser=pooldetective
rpcpassword=pooldetective
rpcclienttimeout=30
rpcallowip=0.0.0.0/0
rpcport=8232
rpcbind=0.0.0.0
printtoconsole=1
disablewallet=1
txindex=1
dbcache=512
prune=0
EOF
fi

if [ $# -eq 0 ]; then
  exec zcashd -datadir=${BITCOIN_DIR} -conf=${BITCOIN_CONF}
else
  exec "$@"
fi
