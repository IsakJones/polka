#! /usr/bin/env bash
set -euo pipefail

# Script prepares all binaries and files. First argument is number of nodes.

# Make list of services
declare -a services=("receiver" "balancer" "cache" "generator" "settler")

# Place env directories and files
for service in ${services[@]}
do
    # The log and the .env files go in a different place for receiver
    if [ $service != "receiver" ]
    then
        mkdir $service/env
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

    popd # $POLKA # Return to original directory

    echo "Built binary in $service/bin"
done