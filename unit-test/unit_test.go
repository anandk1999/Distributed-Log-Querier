package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func runClient(pattern string) string {
	cmd := exec.Command("go", "run", "client.go")
	stdin := bytes.NewBufferString("grep -c " + pattern + "\n")
	cmd.Stdin = stdin

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error running client: %v\n", err)
	}
	return string(output)
}

func checkPatternInOutput(pattern string, output string, expectedMachines []string) bool {
	for _, m := range expectedMachines {
		if !strings.Contains(output, m) {
			fmt.Printf("Expected machine %s missing for pattern %s\n", m, pattern)
			return false
		}
	}

	// Check no unexpected machines contain the pattern
	for i := 1; i < 11; i++ {
		machine := fmt.Sprintf("machine.%d", i)
		found := strings.Contains(output, machine) && strings.Contains(output, pattern)
		expected := contains(expectedMachines, machine)
		if found != expected {
			fmt.Printf("Unexpected result: machine %s, pattern %s, found=%v, expected=%v\n", machine, pattern, found, expected)
			return false
		}
	}

	return true
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func main() {
	// Define patterns and expected machines
	patterns := map[string][]string{
		"RARETOKEN": {"machine.1"},                                                                                                                       // rare pattern appears only on machine 0
		"ERROR":     {"machine.10", "machine.2", "machine.4", "machine.6", "machine.8"},                                                                  // medium pattern on even machines
		"HEARTBEAT": {"machine.10", "machine.1", "machine.2", "machine.3", "machine.4", "machine.5", "machine.6", "machine.7", "machine.8", "machine.9"}, // frequent pattern on all
	}

	allPassed := true
	for pattern, expectedMachines := range patterns {
		output := runClient(pattern)
		fmt.Printf("Testing pattern: %s\n", pattern)
		if !checkPatternInOutput(pattern, output, expectedMachines) {
			allPassed = false
			fmt.Printf("Test failed for pattern: %s\n", pattern)
		} else {
			fmt.Printf("Test passed for pattern: %s\n", pattern)
		}
	}

	if allPassed {
		fmt.Println("All tests passed!")
	} else {
		fmt.Println("Some tests failed!")
	}
}
