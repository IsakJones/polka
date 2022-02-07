#! /usr/bin/bash

# Make list of services
declare -a services=("receiver" "balancer" "cache" "generator")

# Remove node directories
rm -r receiver/node*

echo "Removed node directories"

# Remove env subdirectories
for service in ${services[@]}
do
    rm -r $service/env
done

echo "Removed env directories"
