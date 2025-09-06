#!/bin/bash

# File containing list of VM IPs or hostnames (one per line)
HOSTS_FILE="../hosts.txt"

# SSH key to use
SSH_PUBLIC_KEY="$HOME/.ssh/cs_425_ssh.pub"
SSH_PRIVATE_KEY="$HOME/.ssh/cs_425_ssh"

# Loop through hosts
for HOST in $(cat $HOSTS_FILE); do
  (
    echo ">>> Transferring SSH keys to $HOST"
    ssh-copy-id -i "$HOME/.ssh/cs_425_ssh.pub" "saik2@$HOST"
    ssh "saik2@$HOST" "mkdir -p ~/.ssh && chmod 700 ~/.ssh && \
      echo '
  Host gitlab.engr.illinois.edu
      IdentityFile ~/.ssh/cs_425_ssh
  ' >> ~/.ssh/config && chmod 600 ~/.ssh/config"
  ) &
done

