#! /usr/bin/bash

# Make list of services
declare -a services=("receiver" "balancer" "cache" "generator")

# Kill jobs
for service in ${services[@]}
do
    for pid in $(pgrep "^polka$service$")
    do
        kill -9 $pid
    done
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
