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

# Start tmux session
tmux new-session -d -s polka 'zsh'

# Make node windows
for i in $(eval echo "{0..$NODES}")
do
    if [[ $i -gt 0 ]]
    then
        tmux new-window 
    fi
    tmux rename-window "receiver$i"
    tmux send-keys "cd ${POLKA}/receiver/node$i; ./receiver" 'C-m'
done

# Make window for cache, balancer, and generator
tmux new-window 
tmux rename-window "control"
tmux send-keys "cd ${POLKA}/cache; ./bin/cache" 'C-m'

# Make pane for balancer
tmux split-window -h
tmux send-keys "cd ${POLKA}/balancer; ./bin/balancer" 'C-m'

# Make generator window
tmux split-window -v
tmux send-keys "cd ${POLKA}/generator" "C-m"

tmux -2 attach-session -t polka
