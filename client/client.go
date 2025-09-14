// client: tiny CLI that sends a grep command to all servers and streams results.
// It reads a grep-style command and a file type (demo/unit), fans out requests
// to all hosts in hosts.txt, and aggregates the streaming responses.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"mp1-g02/common"
	"net"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type result struct {
	addr      string
	resp      string
	file_name string
	err       error
}

// fanIn merges many read-only result channels into one channel. When all
// inputs close, the output is closed too. This lets us range over one stream.
func fanIn(channels ...<-chan result) <-chan result {
	out := make(chan result)
	var wg sync.WaitGroup

	wg.Add(len(channels))
	for _, ch := range channels {
		go func(c <-chan result) {
			defer wg.Done()
			for r := range c {
				out <- r
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)

	}()

	return out
}

// connection dials one server, sends the JSON request, and streams newline-
// delimited JSON objects back. We keep pushing results until the server closes
// the connection or the context is canceled.
func connection(a string, message string, ctx context.Context) <-chan result {
	results := make(chan result)

	go func() {
		defer close(results)

		var d net.Dialer
		conn, err := d.DialContext(ctx, "tcp", a)
		if err != nil {
			results <- result{addr: a, err: err}
			return
		}
		defer conn.Close()

		_, err = conn.Write([]byte(message))
		if err != nil {
			results <- result{addr: a, err: err}
			return
		}

		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			// Each line is a JSON object: {output, log_file}
			line := scanner.Text()
			var resp common.ServerResponse
			if err := json.Unmarshal([]byte(line), &resp); err != nil {
				fmt.Println("Error unmarshaling:", err)
				continue
			}

			select {
			case <-ctx.Done():
				return
			case results <- result{addr: a, resp: resp.Output, file_name: resp.LogFile}:
			}
		}
		// Check for scanner errors
		if err := scanner.Err(); err != nil {
			results <- result{addr: a, err: err}
		}
	}()

	return results
}

func main() {
	var countFlag bool = false

	counts := map[string]int{}
	reader := bufio.NewReader(os.Stdin)

	// Ask for user grep command
	fmt.Print("Give me your grep command: ")
	message, _ := reader.ReadString('\n')
	message = strings.TrimSpace(message)

	// If grep command is empty, then don't accept
	parts := strings.Fields(message)
	if len(parts) == 0 {
		fmt.Println("Incorrect grep command:")
	}
	// If user didn't pass -c, we'll count lines locally by tallying messages.
	if !slices.Contains(parts, "-c") {
		countFlag = true
	}

	// Determine if log file is vm[0-9].log or machine.i.log
	fmt.Print("File type (demo/unit): ")
	fileType, _ := reader.ReadString('\n')
	fileType = strings.TrimSpace(fileType)

	// Build JSON request
	req := common.ServerRequest{
		Input:    message,
		FileType: fileType,
	}
	jsonReq, err := json.Marshal(req)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		os.Exit(1)
	}

	// Read hosts.txt and build address list (host:8080 on each line)
	file, err := os.Open("../hosts.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open hosts.txt: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	var addresses []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		host := strings.TrimSpace(scanner.Text())
		if host == "" {
			continue
		}
		addresses = append(addresses, host+":8080")
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to read hosts.txt: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var chans []<-chan result
	start := time.Now()
	for _, addr := range addresses {
		chans = append(chans, connection(addr, string(jsonReq), ctx))
	}
	c := fanIn(chans...)

	total_count := 0

	for r := range c {
		if r.err != nil {
			continue
		}
		fmt.Printf("[%s from %s] response:\n%s\n", r.addr, r.file_name, r.resp)

		if !countFlag {
			count, err := strconv.Atoi(strings.TrimSpace(r.resp))
			// count, err := strconv.Atoi(strings.Split(strings.TrimSpace(r.resp), ":")[1])
			if err != nil {
				fmt.Println("Cannot convert response to integer")
			}
			total_count += count
		} else {
			counts[r.addr]++
		}

	}
	end := time.Since(start)

	if countFlag {
		for key, value := range counts {
			fmt.Println(key, value)
			total_count += value
		}
	}

	fmt.Println("Total count is: ", total_count)
	fmt.Println("Total time taken for last byte in last vm: ", end)

	os.Exit(0)
}
