#!/bin/bash

set -e 

trap "echo Something failed while doing the setup. View docker-setup.log for more details." ERR

if [ -d "localdocker/data" ]; then
    printf "\n\n========\n!!! WARNING\n=======\n\nThere is already a data folder in the localdocker subfolder, which indicates you have already set up the local docker environment. If you want to do a re-setup - ensure the environment is stopped (run [docker-compose down] when in the localdocker folder) and then remove the data folder (you will require root privileges for this). After you've cleaned the data folder up, re-run docker-setup.sh\n\n"
    exit 1
fi

rm -f docker-setup.log

echo "========================================"
echo "Building PoolDetective Docker Containers"
echo "========================================"
echo ""
echo "Now building the docker containers for all the PoolDetective components. This might take a while"
echo ""
./docker-build.sh 2>&1 1>>docker-setup.log

echo "========================================"
echo "Building Coin Daemon Docker Containers  "
echo "========================================"
echo ""
echo "Now building the docker containers for two coin daemons"
echo ""

cd docker-coindaemons
./build.sh "btc ltc" 2>&1 1>>../docker-setup.log
cd ..

echo "==========================================================="
echo "Starting database server and importing schema and demo data"
echo "==========================================================="
echo ""
echo "Starting up just the database server so we can import the PoolDetective schema and demo data"
echo ""

cd localdocker
mkdir -p data/database 
cp ../database/schema.sql data/database/schema.sql
cp ../database/demodata.sql data/database/demodata.sql

docker-compose up -d postgres
sleep 10

echo ""
echo "Importing schema and demo data"
echo ""

docker-compose exec postgres psql -h localhost -U postgres -d postgres -f /tmp/database/schema.sql 2>&1 1>>../docker-setup.log
docker-compose exec postgres psql -h localhost -U postgres -d postgres -f /tmp/database/demodata.sql 2>&1 1>>../docker-setup.log

echo "============================================="
echo "Starting system"
echo "============================================="
echo ""
echo "Now starting the entire system"
echo ""

docker-compose up -d

POSTGRES_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' localdocker_postgres_1)
FRONTEND_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' localdocker_frontend_1)

echo ""
echo "The system is up and running."
echo ""
echo "You can connect to the PostgreSQL database on:"
printf "Host: $POSTGRES_IP\nPort: 5432\nUser: postgres\nPass: pooldetective\nDatabase: postgres\n\n"
echo ""
echo "You can visit the frontend on: http://$FRONTEND_IP/ and the API Docs on http://$FRONTEND_IP/api/docs"
echo ""
echo "You can use the API on http://$FRONTEND_IP/api with API Key [demo] (without the brackets)"
echo ""
echo "You can start stratum mining with a SHA-256 (Bitcoin) miner to port 42651 and a Scrypt (Litecoin) miner to port 42652 - any credential will be accepted"
echo ""
echo ""

