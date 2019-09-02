#!/bin/bash
BITCOIN_DIR=/root/.euno
BITCOIN_CONF=${BITCOIN_DIR}/euno.conf

# If config doesn't exist, initialize with sane defaults for running a
# non-mining node.
if [ ! -e "${BITCOIN_CONF}" ]; then
  tee -a >${BITCOIN_CONF} <<EOF

server=1
rpcuser=pooldetective
rpcpassword=pooldetective2
rpcclienttimeout=30
rpcallowip=*.*.*.*
rpcport=9332
rpcbind=0.0.0.0
printtoconsole=1
disablewallet=1
txindex=1
dbcache=512
prune=0
nosmsg=1
EOF
fi

if [ $# -eq 0 ]; then
  exec eunod -datadir=${BITCOIN_DIR} -conf=${BITCOIN_CONF}
else
  exec "$@"
fi
