#! /usr/bin/env bash

# Make list of services
declare -a services=("settler" "balancer" "cache" "generator")

# Kill tmux session, if any
tmux kill-session -t polka

# Kill services
for service in ${services[@]}
do
    for pid in $(pgrep "^polka$service$")
    do
        kill -SIGINT $pid
    done
done

# Kill receivers
for pid in $(pgrep "^polkareceiver[0-9]+$")
do
    kill -SIGINT $pid
done

echo "Killed all processes"

# Remove node directories
rm -r -f receiver/node*

echo "Removed node directories"

# Remove env subdirectories
for service in ${services[@]}
do
    rm -r -f $service/env
done

echo "Removed env directories"
