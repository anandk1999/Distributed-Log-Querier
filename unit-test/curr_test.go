package main

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func runClient_curr(pattern string) (string, error) {
	cmd := exec.Command("go", "run", "../client/client.go")
	cmd.Stdin = bytes.NewBufferString("grep -c " + pattern + "\n")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func containsAll(out string, want []string) bool {
	for _, m := range want {
		if !strings.Contains(out, m) {
			return false
		}
	}
	return true
}

func TestPatterns(t *testing.T) {
	tests := []struct {
		pattern string
		expect  []string
	}{
		{"RARETOKEN", []string{"machine.1"}},
		{"ERROR", []string{"machine.2", "machine.4", "machine.6", "machine.8", "machine.10"}},
		{"HEARTBEAT", []string{"machine.1", "machine.2", "machine.3", "machine.4", "machine.5", "machine.6", "machine.7", "machine.8", "machine.9", "machine.10"}},
	}
	for _, tc := range tests {
		out, err := runClient_curr(tc.pattern)
		if err != nil {
			t.Fatalf("runClient error for %s: %v\nOutput:\n%s", tc.pattern, err, out)
		}
		if !containsAll(out, tc.expect) {
			t.Errorf("pattern %s: missing expected machines in output:\n%s", tc.pattern, out)
		}
	}
}
