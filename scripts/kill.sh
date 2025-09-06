#!/bin/bash

# Load remote username
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
[ -f "$SCRIPT_DIR/.env" ] && source "$SCRIPT_DIR/.env"
REMOTE_USER="${REMOTE_USER:-saik2}"

HOSTS_FILE="../hosts.txt"

for HOST in $(cat "$HOSTS_FILE"); do
  (
    echo ">>> Killing processes on port 8080 at $HOST"
  ssh "$REMOTE_USER@$HOST" '
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