#!/bin/bash
BITCOIN_DIR=/root/.imgcash
BITCOIN_CONF=${BITCOIN_DIR}/imgcash.conf

# If config doesn't exist, initialize with sane defaults for running a
# non-mining node.
if [ ! -e "${BITCOIN_CONF}" ]; then
  tee -a >${BITCOIN_CONF} <<EOF

server=1
rpcuser=pooldetective
rpcpassword=pooldetective
rpcclienttimeout=30
rpcallowip=0.0.0.0/0
rpcport=6898
rpcbind=0.0.0.0
printtoconsole=1
disablewallet=1
txindex=1
dbcache=512
prune=0
addnode=23.101.61.34:6888
addnode=111.220.94.247:6888
addnode=43.226.40.56:6888
addnode=217.61.104.163:6888
addnode=167.114.159.30:6888
addnode=149.56.115.240:6888
addnode=108.61.188.68:6888
EOF
fi

if [ $# -eq 0 ]; then
  exec imgcashd -datadir=${BITCOIN_DIR} -conf=${BITCOIN_CONF}
else
  exec "$@"
fi
