#! /usr/bin/bash
set -euo pipefail

# Script prepares all binaries and files. First argument is number of nodes.

# Define polka directory variable
POLKA=~/code/projects/polka
# Make list of services
declare -a services=("balancer" "cache")

# start services
for service in ${services[@]}
do
    # Start process in the background
    cd $POLKA/$service
    ./bin/polka$service > log.txt 2>&1 &
done

echo "Started cache and balancer"

cd $POLKA

# start nodes
for node in $(find -name "node*")
do
    cd $POLKA/$node
    ./polkareceiver > log.txt 2>&1 &
done

echo "Started receiver nodes"

cd $POLKA

# Display logs with multitail
multitail -s 2 -sn 1,2 cache/log.txt \
                       balancer/log.txt \
                       receiver/node0/log.txt
