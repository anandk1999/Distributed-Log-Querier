package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run log_generator.go <machineNumber>")
		os.Exit(1)
	}
	machine := os.Args[1]
	filename := fmt.Sprintf("machine.%s.log", machine)

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("failed to create log file %s: %v\n", filename, err)
		os.Exit(1)
	}
	defer file.Close()

	rand.Seed(time.Now().UnixNano())

	// --- Known patterns ---

	// Rare pattern: only put this in machine 0
	if machine == "0" {
		fmt.Fprintln(file, "ALERT: RARETOKEN found on machine 0")
	}

	// Frequent pattern: put this many times on every machine
	for i := 0; i < 50; i++ {
		fmt.Fprintf(file, "INFO: HEARTBEAT OK machine=%s cycle=%d\n", machine, i)
	}

	// Somewhat Frequent pattern: only in even-numbered machines
	if machineInt := int(machine[0] - '0'); machineInt%2 == 0 {
		for i := 0; i < 5; i++ {
			fmt.Fprintf(file, "ERROR: something failed on machine %s at iter=%d\n", machine, i)
		}
	}

	// --- Random filler lines ---
	randomMsgs := []string{
		"DEBUG: connection established",
		"INFO: user logged in",
		"TRACE: value updated",
		"DEBUG: cache refreshed",
		"INFO: job completed",
	}
	for i := 0; i < 200; i++ {
		msg := randomMsgs[rand.Intn(len(randomMsgs))]
		fmt.Fprintf(file, "%s\n", msg)
	}

	fmt.Printf("Generated %s\n", filename)
}
