#! /usr/bin/bash
set -euo pipefail

QB="UPDATE banks SET balance=0;"
QA="UPDATE accounts SET balance=0;"


psql -h 127.0.0.1 -U polka -d payments -c "$QB $QA"