#! /usr/bin/bash
set -euo pipefail

# Restore balancer.env
sed -i "/NODEADDRESS/d" balancer/env/balancer.env
sed -i "s/NODENUM=\([0-9]\+\)/NODENUM=0/g" balancer/env/balancer.env

# Remove node directories
rm -r receiver/node*

# End all tmux sessions
tmux kill-server