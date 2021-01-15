#!/bin/bash

BUILD="api apidocs frontend dbqueries coordinator coordinatorhost blockfetcher blockobserver blockobserverhost stratumclient stratumclienthost stratumserver pubsubhost"
echo "Argument: [$1]"

if [ ! -z "$1" ]; then
    BUILD="$1"
fi

echo "Building: [$BUILD]"

if [ -z "$PREFIX" ]; then
    PREFIX="pooldetective"
fi

TAG_PREFIX=$PREFIX
if [ ! -z "$DOCKER_REGISTRY" ]; then
    TAG_PREFIX="$DOCKER_REGISTRY/$PREFIX"
fi

docker build . -f "Dockerfile.base" -t "$TAG_PREFIX-base"

for prod in $BUILD 
do
    echo "Building $prod"
    docker build --build-arg baseimage="$TAG_PREFIX-base" . -f "Dockerfile.$prod" -t "$TAG_PREFIX-$prod"
    if [ ! -z "$DOCKER_REGISTRY" ]; then
        docker push "$TAG_PREFIX-$prod"
    fi
done