#!/bin/bash
BITCOIN_DIR=/root/.bitcoinz 
BITCOIN_CONF=${BITCOIN_DIR}/bitcoinz.conf

# If config doesn't exist, initialize with sane defaults for running a
# non-mining node.
if [ ! -e ${BITCOIN_CONF} ]; then
  fetch-params.sh
  tee -a >${BITCOIN_CONF} <<EOF

server=1
rpcuser=pooldetective
rpcpassword=pooldetective
rpcclienttimeout=30
rpcallowip=0.0.0.0/0
rpcport=1979
rpcbind=0.0.0.0
printtoconsole=1
disablewallet=1
txindex=0
testnet=0
dbcache=512
prune=1
addnode=btcz.kovach.biz
addnode=seed.btcz.life
addnode=bzseed.secnode.tk
addnode=btzseed.blockhub.info
addnode=btczseed.1ds.us
addnode=104.131.83.192
addnode=52.207.253.9
addnode=54.174.68.212
addnode=155.138.144.191:1989
addnode=107.191.52.3:1989
addnode=207.246.92.95:1989
addnode=144.202.51.98:1989
addnode=45.63.119.108:1989
addnode=45.77.61.21:1989
addnode=54.87.127.115:1989
addnode=3.92.242.49:1989
addnode=3.89.28.227:1989
addnode=54.165.70.17:1989
addnode=155.138.236.125:1989
EOF
fi


if [ $# -eq 0 ]; then
  exec bitcoinzd
else
  exec "$@"
fi
