#!/bin/bash
BITCOIN_DIR=/root/.aricoin
BITCOIN_CONF=${BITCOIN_DIR}/aricoin.conf

# If config doesn't exist, initialize with sane defaults for running a
# non-mining node.
if [ ! -e "${BITCOIN_CONF}" ]; then
  tee -a >${BITCOIN_CONF} <<EOF

server=1
rpcuser=pooldetective
rpcpassword=pooldetective2
rpcclienttimeout=30
rpcallowip=0.0.0.0/0
rpcport=13555
rpcbind=0.0.0.0
printtoconsole=1
disablewallet=1
EOF
fi

if [ $# -eq 0 ]; then
  exec aricoind -datadir=${BITCOIN_DIR} -conf=${BITCOIN_CONF}
else
  exec "$@"
fi
