#!/bin/bash
FILE="$PWD/futurocoind"
if [ ! -f "$FILE" ]; then
     git clone --depth=1 https://github.com/futuro-coin/Futuro-binaries
     cd Futuro-binaries
     tar -xvf futurocoincore-1.1.0-ubuntu18_04-x86_64.tar.xz
     mv futurocoincore-1.1.0-ubuntu18_04-x86_64/futurocoin* ..
     rm ../futurocoin-qt
     cd ..
     rm -rf Futuro-binaries
fi
