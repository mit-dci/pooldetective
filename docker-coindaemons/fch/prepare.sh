#!/bin/bash
FILE="$PWD/fc_miner_1.0.4/fc.io_1.0.4.tar.gz"
if [ ! -f "$FILE" ]; then
     wget https://download.sign.cash/fc_miner_docker_1.0.4.zip
     unzip fc_miner_docker_1.0.4.zip
     rm fc_miner_docker_1.0.4.zip
fi

docker load -i fc_miner_1.0.4/fc.io_1.0.4.tar.gz
docker tag fc.io:1.0.4 freecash-base

