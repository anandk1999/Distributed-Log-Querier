package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"mp1-g02/common"
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
	operands := parts[1:]
	// if !slices.Contains(parts, "-c") {
	// 	operands = append([]string{"-c"}, operands...)
	// }
	return parts[0], operands
}

// streamGrep runs the grep command and streams each matching line back as a
// newline-delimited JSON object over the provided connection.
func streamGrep(conn net.Conn, query string, logFile string) {
	name, args := parseCommand(query)
	args = append(args, logFile)

	cmd := exec.Command(name, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("failed to get stdout pipe: %v", err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("failed to get stderr pipe: %v", err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("failed to start grep: %v", err)
		return
	}

	// Stream matches line-by-line
	outScanner := bufio.NewScanner(stdout)
	for outScanner.Scan() {
		line := outScanner.Text()
		response := common.ServerResponse{Output: line, LogFile: logFile}
		if b, err := json.Marshal(response); err == nil {
			// newline-delimited JSON for the client scanner
			if _, werr := conn.Write(append(b, '\n')); werr != nil {
				log.Printf("failed to write to client: %v", werr)
				break
			}
		} else {
			log.Printf("failed to marshal response: %v", err)
			break
		}
	}
	if err := outScanner.Err(); err != nil {
		log.Printf("stdout scan error: %v", err)
	}

	// Drain/inspect stderr (optional): read and log any errors after command finishes
	// We purposefully start a goroutine to avoid blocking if grep is still running
	go func() {
		s := bufio.NewScanner(stderr)
		for s.Scan() {
			log.Printf("grep stderr: %s", s.Text())
		}
	}()

	// Wait for the command to finish. grep returns exit code 1 when no matches are found.
	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				// No matches; not an actual failure for our streaming use-case.
				return
			}
			log.Printf("grep exited with error code %d", exitErr.ExitCode())
			return
		}
		log.Printf("error waiting for grep to finish: %v", err)
	}
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

	var req common.ServerRequest
	err = json.Unmarshal(buffer[:n], &req)
	if err != nil {
		fmt.Println("Error unmarshaling request:", err)
		return
	}

	fmt.Printf("Received from client: %s\n", req.Input)

	var log_file string
	switch req.FileType {
	case "demo":
		log_file = fmt.Sprintf("./vm%s.log", machine)
	case "unit":
		log_file = fmt.Sprintf("./machine.%s.log", machine)
	default:
		log_file = fmt.Sprintf("./vm%s.log", machine)
	}

	// Stream responses back to the client, one JSON per matching line
	streamGrep(conn, req.Input, log_file)
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
