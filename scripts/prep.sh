#! /usr/bin/env bash
set -euo pipefail

# Script prepares all binaries and files. First argument is number of nodes.

# Define base port for receiver nodes
BASEPORT=8082
# Define number of receiver nodes
NODES=$(($1-1))
# Make list of services
declare -a services=("receiver" "balancer" "cache" "generator")

# Place env directories and files
for service in ${services[@]}
do
    # The log and the .env files go in a different place for receivers
    if [ $service != "receiver" ]
    then
        mkdir $service/env
        touch $service/log.txt
        cp envs/$service.env $service/env/
    fi
done

echo "Made env directories and log files"

# Place db.env file in cache
cp envs/db.env cache/env/

# Build all binaries
for service in ${services[@]}
do
    pushd "$service" # Go to source file

    go build -o bin/polka$service ./src

    # go build .
    # mv src ../bin/polka$service

    popd # $POLKA # Return to original directory

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
    touch "receiver/node$i/log.txt"
    cp receiver/bin/polkareceiver "receiver/node$i"
    cp envs/receiver.env "receiver/node$i/"
    cp envs/db.env "receiver/node$i/"
    sed -i -e "s/${BASEPORT}/${CURPORT}/g" "receiver/node$i/receiver.env"
    echo "Prepared node $i"
done