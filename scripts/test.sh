#! /usr/bin/bash


echo $1
VAR=$1

for i in $(eval echo "{0..$1}")
do
    NEWVAR=$((VAR+i))

    if [[ $i -gt 0 ]]
    then
        echo "Hello, Isak!"
    fi

    echo "Hello, ${i}"
done
