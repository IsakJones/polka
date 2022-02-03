#! /usr/bin/bash
set -euo pipefail

POLKA=~/code/projects/polka

# builds a go binary in a directory named $1
# remember to execute from polka directory!
build() {
    cd $1/src # Go to source file

    go build .
    mv src ../bin/$1

    cd $POLKA # Return to original directory

    echo "Built binary in $1/bin"
}

declare -a servicesToNest=("receiver" "balancer" "cache" "generator")

for service in ${servicesToNest[@]}
do
    build $service
done