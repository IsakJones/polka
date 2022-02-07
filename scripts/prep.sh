#! /usr/bin/bash
set -euo pipefail

# Script prepares all binaries and files. First argument is number of nodes.

# Define polka directory variable
POLKA=~/code/projects/polka
# Define base port for receiver nodes
BASEPORT=8082
# Define number of receiver nodes
NODES=$(($1-1))
# Make list of services
declare -a services=("receiver" "balancer" "cache" "generator")

# Place env directories and files
for service in ${services[@]}
do
    mkdir $service/env
    cp envs/$service.env $service/env/
done

echo "Made env directories"

# Place db.env file in cache
cp envs/db.env cache/env/

# Build all binaries
for service in ${services[@]}
do
    cd $service/src # Go to source file

    go build .
    mv src ../bin/$service

    cd $POLKA # Return to original directory

    echo "Built binary in $service/bin"
done

# Set number of nodes in balancer.env
sed -i -e "s/NODENUM=0/NODENUM=$1/g" balancer/env/balancer.env

for i in $(eval echo "{0..$NODES}")
do
    CURPORT=$((BASEPORT+i))
    # Add address to node
    echo "NODEADDRESS$i=http://localhost:$CURPORT" >> balancer/env/balancer.env
    # Make node directory and include .env files
    mkdir "receiver/node$i"
    cp receiver/bin/receiver "receiver/node$i"
    cp receiver/env/receiver.env "receiver/node$i/"
    cp envs/db.env "receiver/node$i/"
    sed -i -e "s/${BASEPORT}/${CURPORT}/g" "receiver/node$i/receiver.env"
    echo "Prepared node $i"
done