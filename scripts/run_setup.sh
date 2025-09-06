#!/bin/bash

HOSTS_FILE="../hosts.txt"

for HOST in $(cat "$HOSTS_FILE"); do
  (
    echo ">>> Running setup script on $HOST"
    ssh saik2@"$HOST" "nohup ./setup.sh > setup.log 2>&1 &" </dev/null
  ) &
done

wait   # wait for all SSH sessions to be started