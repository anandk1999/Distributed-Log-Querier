#!/bin/bash

HOSTS_FILE="../hosts.txt"
for HOST in $(cat $HOSTS_FILE); do
  (
    echo ">>> Pulling new code from repo on $HOST"
    ssh saik2@"$HOST" "cd mp1-g02 && git pull"
  ) &
done
wait