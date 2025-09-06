#!/bin/bash

HOSTS_FILE="../hosts.txt"

TARGETS=("go" "go.mod" "*.log" "mp1-g02" "setup.log" "setup.sh" "tmp")

for HOST in $(cat "$HOSTS_FILE"); do
  (
    echo ">>> Cleaning specified files on $HOST"
    ssh saik2@"$HOST" "
      for item in ${TARGETS[@]}; do
        if [ -e /home/saik2/\$item ]; then
          echo \"Removing /home/saik2/\$item\"
          rm -rf /home/saik2/\$item
        fi
      done
    "
  ) &
done
wait