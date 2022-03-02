#! /usr/bin/env bash
set -euo pipefail

# Script prepares all binaries and files. First argument is number of nodes.

# Define base port for processor nodes
BASEPORT=8083
# Define number of processor nodes
NODES=$(($1-1))
# Make list of services
declare -a services=("processor" "balancer" "cache" "generator" "settler")

# Place env directories and files
for service in ${services[@]}
do
    # The log and the .env files go in a different place for processor
    if [ $service != "processor" ]
    then
        mkdir $service/env
        touch $service/log.txt
        cp envs/$service.env $service/env/
    fi
done

echo "Made env directories and log files"

# Place postgres.env file in cache
cp envs/postgres.env cache/env/

# Place mongo.env file in settler
cp envs/mongo.env settler/env/

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
    mkdir "processor/node$i"
    touch "processor/node$i/log.txt"
    cp "processor/bin/polkaprocessor" "processor/node$i/polkaprocessor$i"
    cp "envs/processor.env" "processor/node$i/"
    cp "envs/postgres.env" "processor/node$i/"
    sed -i -e "s/${BASEPORT}/${CURPORT}/g" "processor/node$i/processor.env"
    echo "Prepared node $i"
done