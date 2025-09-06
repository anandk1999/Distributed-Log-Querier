#!/bin/bash

# File containing list of VM IPs or hostnames (one per line)
HOSTS_FILE="../hosts.txt"

# SSH key to use
SSH_PUBLIC_KEY="$HOME/.ssh/cs_425_ssh.pub"
SSH_PRIVATE_KEY="$HOME/.ssh/cs_425_ssh"

# Loop through hosts
for HOST in $(cat $HOSTS_FILE); do
  (  
    echo ">>> Setting up $HOST"
    scp $SSH_PRIVATE_KEY saik2@$HOST:/home/saik2/.ssh/
    scp $SSH_PUBLIC_KEY saik2@$HOST:/home/saik2/.ssh/
  ) &
done
wait