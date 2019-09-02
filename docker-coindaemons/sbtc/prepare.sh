#!/bin/bash
FILE="$PWD/sbtcd"
FILENAME="sbtc-0.17.7-ubuntu16_04.tar.xz"
DOWNLOAD_URL="https://github.com/superbitcoin/SuperBitcoin/releases/download/v0.17.7/$FILENAME"

if [ ! -f "$FILE" ]; then
     mkdir download
     cd download
     wget $DOWNLOAD_URL
     tar -xvf $FILENAME 
     rm -rf $FILENAME 
     rm -rf sbtc-qt
     mv sbtc-0.17.7-ubuntu16_04/sbtcd ..
     mv sbtc-0.17.7-ubuntu16_04/sbtc-cli ..
     mv sbtc-0.17.7-ubuntu16_04/log.conf ..
     
     cd ..
     rm -rf download
fi
