#! /usr/bin/bash
set -euo pipefail

# Script starts all processes.
# Must download 'multitail' for the terminal.

# Set project root directory
POLKA=$(pwd)
# Make list of services
declare -a services=("balancer" "cache")

# Get change ulimits flag
u_flag=''
while getopts 'abf:v' flag; do
  case "${flag}" in
    u) u_flag='true' ;;
  esac
done

# Start tmux session
tmux new-session -d -s polka "zsh"

# Make node windows
i=0
for node in $(find -name "node*")
do
    cd "$POLKA/$node"
    tmux rename-window "receiver$i"
    tmux send-keys "cd ${POLKA}/receiver/node$i; ./polkareceiver" "C-m"
    tmux new-window
    i=$((i+1))
done
cd $POLKA

# Make window for cache, balancer, and generator
tmux rename-window "control"
tmux send-keys "cd ${POLKA}/cache; ./bin/polkacache" "C-m"

# Make pane for balancer
tmux split-window -h
tmux send-keys "cd ${POLKA}/balancer; ./bin/polkabalancer" "C-m"

# Make generator window
tmux split-window -v
tmux send-keys "cd ${POLKA}/generator" "C-m"
tmux send-keys "echo 'Run ./bin/polkabalancer -w=1000 -t=1000 to send 1000 transactions with 1000 workers.'" "C-m"

if [ u_flag = true ]
then
    # Increase ulimits for balancer and cache
    for $service in ${services[@]}
    do
        prlimit --pid $(pgrep "^polka$service$") --nofile=10000:10000
    done

    # Increase ulimits for receivers
    for node in $(find -name "node*")
    do
        cd "$POLKA/$node"

        tmux rename-window "receiver$i"
        tmux send-keys "cd ${POLKA}/receiver/node$i; ./polkareceiver" "C-m"
        tmux new-window
        i=$((i+1))
    done
fi

tmux -2 attach-session -t polka