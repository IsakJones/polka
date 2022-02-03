#! /usr/bin/bash
set -euo pipefail

# Script starts application. First argument is number of nodes.

# Define polka directory variable
POLKA=~/code/projects/polka
# Define base port for receiver nodes
BASEPORT=8082
# Define number of receiver nodes
NODES=$(($1-1))

# Set number of nodes in balances.env
sed -i -e "s/NODENUM=0/NODENUM=$1/g" balancer/env/balancer.env

for i in $(eval echo "{0..$NODES}")
do
    CURPORT=$((BASEPORT+i))
    # Add address to node
    echo "NODEADDRESS$i=http://localhost:$CURPORT" >> balancer/env/balancer.env
    # Make node directory and include .env files
    mkdir "receiver/node$i"
    cp receiver/bin/receiver "receiver/node$i"
    cp receiver/env/db.env "receiver/node$i/"
    cp receiver/env/receiver.env "receiver/node$i/"
    sed -i -e "s/${BASEPORT}/${CURPORT}/g" "receiver/node$i/receiver.env"
done
