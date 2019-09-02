#!/bin/bash
BITCOIN_DIR=/root/.imagecoin
BITCOIN_CONF=${BITCOIN_DIR}/imagecoin.conf

# If config doesn't exist, initialize with sane defaults for running a
# non-mining node.
if [ ! -e "${BITCOIN_CONF}" ]; then
  tee -a >${BITCOIN_CONF} <<EOF

server=1
rpcuser=pooldetective
rpcpassword=pooldetective
rpcclienttimeout=30
rpcallowip=0.0.0.0/0
rpcport=9337
rpcbind=0.0.0.0
printtoconsole=1
disablewallet=1
txindex=1
dbcache=512
prune=0
addnode=23.101.61.34:6998
addnode=51.77.144.203:6998
addnode=5.189.162.110:6998
addnode=94.16.122.165:6998
addnode=79.135.200.25:6998
addnode=54.38.158.185:6998
addnode=37.187.127.238:6998
EOF
fi

if [ $# -eq 0 ]; then
  exec ImageCoind -datadir=${BITCOIN_DIR} -conf=${BITCOIN_CONF}
else
  exec "$@"
fi
