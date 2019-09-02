#!/bin/bash
FILE="$PWD/bolivarcoind"
FILENAME="BolivarCoinDaemonUbuntu1804.tar.gz"
DOWNLOAD_URL="https://github.com/BOLI-Project/BolivarCoin/releases/download/v2.0.0.2/$FILENAME"

if [ ! -f "$FILE" ]; then
     mkdir download
     cd download
     wget $DOWNLOAD_URL
     tar -xvf $FILENAME 
     rm -rf $FILENAME 
     rm -rf imgcash-qt
     mv DaemonsLinuxBOLI/* ..
     cd ..
     rm -rf download
fi
