package main

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
)

func parseCommand(input string) (string, []string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", nil
	}
	for i, f := range parts {
		parts[i] = strings.Trim(f, `"`)
	}
	return parts[0], parts[1:]
}

func executeGrep(query string) string {
	name, args := parseCommand(query)
	args = append(args, string("./machine.1.log"))
	cmd := exec.Command(name, args...)
	fmt.Println("Outputs from grep command: ")

	// Run the command and capture its output
	output, err := cmd.Output()
	if err != nil {
		// Handle potential errors, such as the command not being found or exiting with a non-zero status
		if exitError, ok := err.(*exec.ExitError); ok {
			log.Printf("Grep exited with error: %s\nStderr: %s", exitError.Error(), exitError.Stderr)
		} else {
			log.Fatalf("Failed to run grep: %v", err)
		}
	}

	// Print the captured output
	fmt.Println(string(output))
	return string(output)
}

func handleConnection(conn net.Conn) {
	defer conn.Close() // Ensure connection is closed

	// Read data from the client
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err)
		return
	}

	query := string(buffer[:n])
	fmt.Printf("Received from client: %s\n", query)

	// Send a response back to the client
	_, err = conn.Write([]byte(executeGrep(query)))
	if err != nil {
		fmt.Println("Error writing:", err)
		return
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server listening on :8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err)
			continue
		}
		go handleConnection(conn) // Handle connection in a goroutine
	}
}
