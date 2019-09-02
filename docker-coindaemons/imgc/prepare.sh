#!/bin/bash
FILE="$PWD/imgcashd"
FILENAME="imgcash_linux_x86_64_70212.tar.xz"
DOWNLOAD_URL="https://github.com/mceme/ImageCash/releases/download/1.12/$FILENAME"

if [ ! -f "$FILE" ]; then
     mkdir download
     cd download
     wget $DOWNLOAD_URL
     tar -xvf $FILENAME 
     rm -rf $FILENAME 
     rm -rf imgcash-qt
     mv * ..
     cd ..
     rm -rf download
fi
