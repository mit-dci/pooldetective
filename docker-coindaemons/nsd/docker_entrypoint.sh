#!/bin/bash
BITCOIN_DIR=/root/.nasdacoin
BITCOIN_CONF=${BITCOIN_DIR}/nasdacoin.conf

# If config doesn't exist, initialize with sane defaults for running a
# non-mining node.
if [ ! -e "${BITCOIN_CONF}" ]; then
  tee -a >${BITCOIN_CONF} <<EOF

server=1
rpcuser=pooldetective
rpcpassword=pooldetective
rpcclienttimeout=30
rpcallowip=*.*.*.*
rpcport=13555
rpcbind=0.0.0.0
printtoconsole=1
disablewallet=1
EOF
fi

if [ $# -eq 0 ]; then
  exec Nasdacoind -datadir=${BITCOIN_DIR} -conf=${BITCOIN_CONF}
else
  exec "$@"
fi
