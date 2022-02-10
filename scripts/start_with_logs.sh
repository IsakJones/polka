#! /usr/bin/bash
set -euo pipefail

# Script starts all processes with logs.
# Must download 'multitail' for the terminal.

# Make list of services
declare -a services=("balancer" "cache")

# start services
for service in ${services[@]}
do
    # Start process in the background
    pushd $service
    ./bin/polka$service > log.txt 2>&1 &
    popd
done

echo "Started cache and balancer"

# start nodes
for node in $(find -name "node*")
do
    pushd $node
    ./polkareceiver > log.txt 2>&1 &
    popd
done

echo "Started receiver nodes"

# Display logs with multitail
multitail -s 2 -sn 1,2 cache/log.txt \
                       balancer/log.txt \
                       receiver/node0/log.txt
