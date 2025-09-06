#!/bin/bash

# Load remote username
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
[ -f "$SCRIPT_DIR/.env" ] && source "$SCRIPT_DIR/.env"
REMOTE_USER="${REMOTE_USER:-saik2}"

# Ask for file to transfer
read -p "Enter file name: " file

# Check if file exists
if [[ ! -f "$file" ]]; then
  echo "Error: File '$file' not found."
  exit 1
fi

# Confirm file with user
read -p "Continue? (Y/N): " confirm && [[ $confirm =~ ^[Yy](es)?$ ]] || exit 1

HOSTS_FILE="../hosts.txt"
# Loop through hosts
for HOST in $(cat $HOSTS_FILE); do
  (
    echo ">>> Transferring $file to $HOST"
  scp "$file" "$REMOTE_USER@$HOST:~/"

    # If it's a .sh file, give execute permissions on the VM
    if [[ "$file" == *.sh ]]; then
      echo ">>> Setting execute permission on $HOST"
  ssh "$REMOTE_USER@$HOST" "chmod +x \$HOME/$(basename "$file")"
    fi
  ) &
done
wait