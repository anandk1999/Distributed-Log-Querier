#!/bin/bash

LOG_DIR="/Users/ASUS1/Desktop/Fall 2025 Semester (UIUC)/Distributed Systems/Machine Problems/MP1/MP1 Demo Data FA22"
CLIENT_OUTPUT_FILE="../client_output.txt"
PATTERN='[0-9]{2}/Aug/[0-9]{4}'
PATTERN='AppleWebKit/5320'
PATTERN='138.171.13.249'

read -p "Enter the log type (e.g., demo or unit): " LOG_TYPE

echo "Running client.go and saving output to $CLIENT_OUTPUT_FILE..."
echo `pwd`
echo "grep -c $PATTERN \n $LOG_TYPE" | go run ../client/client.go > "$CLIENT_OUTPUT_FILE"

if [[ $? -ne 0 ]]; then
    echo "❌ Failed to run client.go. Exiting."
    exit 1
fi

echo "✅ client.go executed successfully."
echo

echo "Comparing client output to local grep counts..."
echo "Using pattern: $PATTERN"
echo

local_filenames=()
local_counts=()

for log_file in "$LOG_DIR"/*.log; do
    if [[ -f "$log_file" ]]; then
        filename=$(basename "$log_file")
        count=$(grep -c "$PATTERN" "$log_file")
        echo $count
        local_filenames+=("$filename")
        local_counts+=("$count")
    fi
done



echo "Report:"
echo "-------------------"
mismatch_count=0
total_checked=0

while IFS= read -r response_line && IFS= read -r data_line; do
    # Extract log filename and remote count from the second line
    if [[ "$data_line" =~ \./(vm[0-9]+\.log):([0-9]+) ]]; then
        machine="${BASH_REMATCH[1]}"
        remote_count="${BASH_REMATCH[2]}"

        # Find the corresponding local count
        local_count=""
        for i in "${!local_filenames[@]}"; do
            if [[ "${local_filenames[$i]}" == "$machine" ]]; then
                local_count="${local_counts[$i]}"
                break
            fi
        done

        if [[ -n "$local_count" ]]; then
            ((total_checked++))
            if [[ "$remote_count" -ne "$local_count" ]]; then
                echo "❌ Mismatch in $machine → Remote: $remote_count, Local: $local_count"
                ((mismatch_count++))
            else
                echo "✅ Match in $machine → Count: $remote_count"
            fi
        fi
    fi
done < "$CLIENT_OUTPUT_FILE"

