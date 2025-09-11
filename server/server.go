package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
)

func parseCommand(input string) (string, []string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}

func executeGrep(query string, machine string) (string, string) {
	name, args := parseCommand(query)
	log_file := fmt.Sprintf("./machine.%s.log", machine)
	args = append(args, log_file)
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
	return string(output), log_file
}

func handleConnection(conn net.Conn, machine string) {
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
	out, log_file := executeGrep(query, machine)
	fmt.Printf("SERVER OUTPUT BEFORE SENDING BACK TO CLIENT: %s", out)
	_, err = conn.Write([]byte(out + " " + log_file + "\n"))
	if err != nil {
		fmt.Println("Error writing:", err)
		return
	}
}

func getMachineNumber() string {
	file, err := os.Open("mapping.txt")
	if err != nil {
		log.Fatalf("failed to open mapping file: %v", err)
	}
	defer file.Close()

	mapping := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue // skip invalid lines
		}
		host := strings.TrimSpace(parts[0])
		num := strings.TrimSpace(parts[1])
		mapping[host] = num
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("failed to scan mapping file: %v", err)
	}

	out, err := exec.Command("hostname").Output()
	if err != nil {
		log.Fatalf("failed to run hostname: %v", err)
	}
	hostname := strings.TrimSpace(string(out))

	num, ok := mapping[hostname]
	if !ok {
		log.Fatalf("hostname %s not found in mapping file", hostname)
	}

	return num
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server listening on :8080")

	machine := getMachineNumber()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err)
			continue
		}
		go handleConnection(conn, machine) // Handle connection in a goroutine
	}
}
