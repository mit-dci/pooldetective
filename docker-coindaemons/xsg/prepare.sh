#!/bin/bash
FILE="$PWD/snowgemd"
FILENAME="snowgem-ubuntu18.04-3000458-20200807.zip"
DOWNLOAD_URL="https://github.com/Snowgem/Snowgem/releases/download/v3000458/$FILENAME"

if [ ! -f "$FILE" ]; then
     mkdir download
     cd download
     wget $DOWNLOAD_URL
     unzip $FILENAME 
     rm -rf $FILENAME 
     mv snowgemd ..
     mv snowgem-cli ..
     
     cd ..
     rm -rf download
fi
