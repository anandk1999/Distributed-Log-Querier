#!/bin/bash

# Ask for file to transfer
read -p "Enter file name: " file

# Check if file exists
if [[ ! -f "$file" ]]; then
  echo "Error: File '$file' not found."
  exit 1
fi

# Confirm file with user
read -p "Continue? (Y/N): " confirm && [[ $confirm =~ ^[Yy](es)?$ ]] || exit 1

# Loop through hosts
for HOST in $(cat hosts.txt); do
  echo ">>> Transferring $file to $HOST"
  scp "$file" saik2@"$HOST":/home/saik2/

  # If it's a .sh file, give execute permissions on the VM
  if [[ "$file" == *.sh ]]; then
    echo ">>> Setting execute permission on $HOST"
    ssh saik2@"$HOST" "chmod +x /home/saik2/$(basename "$file")"
  fi
done