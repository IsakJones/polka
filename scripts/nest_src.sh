#! /usr/bin/bash
set -euo pipefail

# nests contents of a folder in a src/ directory and makes a bin/ directory
nest() {
    echo $(pwd)

    mv $1 src/

    echo $(ls src)

    mkdir $1/
    mkdir $1/src
    mkdir $1/bin

    mv src/ $1/
}

unnest() {
    echo $(pwd)

    mv $1/src/src $1/source
    rm -r $1/src
    mv $1/source $1/src
}

makeenv() {
    mkdir $1/env
    mv $1/src/*.env $1/env
}


declare -a servicesToNest=("processor" "balancer" "cache" "generator")

for service in ${servicesToNest[@]}
do
    makeenv ../$service
done

