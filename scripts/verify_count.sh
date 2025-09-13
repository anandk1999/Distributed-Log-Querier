#!/bin/bash

LOG_DIR="/Users/ASUS1/Desktop/Fall 2025 Semester (UIUC)/Distributed Systems/Machine Problems/MP1/MP1 Demo Data FA22"
CLIENT_OUTPUT_FILE="../client_output.txt"
PATTERN='[0-9]{2}/Aug/[0-9]{4}'

read -p "Enter the log type (e.g., demo or unit): " LOG_TYPE

echo "Running client.go and saving output to $CLIENT_OUTPUT_FILE..."
go run ../client/client.go > "$CLIENT_OUTPUT_FILE" <<EOF
$PATTERN
$LOG_TYPE
EOF

if [[ $? -ne 0 ]]; then
    echo "❌ Failed to run client.go. Exiting."
    exit 1
fi

echo "✅ client.go executed successfully."
echo

echo "Comparing client output to local grep counts..."
echo "Using pattern: $PATTERN"
echo

declare -A local_counts

for log_file in "$LOG_DIR"/*.log; do
    if [[ -f "$log_file" ]]; then
        filename=$(basename "$log_file")
        count=$(grep -c -E "$PATTERN" "$log_file")
        local_counts["$filename"]=$count
    fi
done

echo "Report:"
echo "-------------------"
while IFS= read -r line; do
    if [[ "$line" =~ \[(.*)from\s+\./(vm[0-9]+\.log)\] response:\s*([0-9]+) ]]; then
        machine="${BASH_REMATCH[2]}"
        remote_count="${BASH_REMATCH[3]}"
        local_count="${local_counts[$machine]}"

        if [[ "$remote_count" -ne "$local_count" ]]; then
            echo "❌ Mismatch in $machine → Remote: $remote_count, Local: $local_count"
        else
            echo "✅ Match in $machine → Count: $remote_count"
        fi
    fi
done < "$CLIENT_OUTPUT_FILE"
