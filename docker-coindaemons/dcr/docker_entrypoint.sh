#!/bin/bash

cd /home/decred
echo "We are currently in: $PWD"
CONTENTS=$(ls)
echo "Contents of this directory is: $CONTENTS"
./dcrd