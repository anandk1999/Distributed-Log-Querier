#!/bin/bash

HOSTS_FILE="../hosts.txt"

for HOST in $(cat "$HOSTS_FILE"); do
  (
    echo ">>> Killing processes on port 8080 at $HOST"
    ssh saik2@"$HOST" '
      PIDS=$(lsof -ti:8080)
      if [ -n "$PIDS" ]; then
        echo "Killing: $PIDS"
        kill -9 $PIDS
      else
        echo "No process running on port 8080"
      fi
    '
  ) &
done
wait