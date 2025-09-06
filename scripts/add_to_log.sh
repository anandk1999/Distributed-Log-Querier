#!/bin/bash

HOSTS_FILE="../hosts.txt"

read -r -d '' TEXT_TO_WRITE <<'EOF'
hi
hello
bye
goodbye
Hi
HI
HELLO
BYE
GOODBYE
EOF

# Loop through all hosts
for HOST in $(cat $HOSTS_FILE); do
  if [ -n "$HOST" ]; then
    echo ">>> Writing to log file in $HOST"
    ssh saik2@"$HOST" "
      host=\$(hostname)
      num=\${host#fa25-cs425-02}
      num=\${num%%.*}
      num=\$((10#\$num))
      cat >> ./mp1-g02/machine.\$num.log <<EOT
$TEXT_TO_WRITE
EOT
    "
  fi
done < "$HOSTS_FILE"
