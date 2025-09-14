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

var (
	countMap = make(map[string]int)
	countMu  sync.Mutex
)

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
				countMu.Lock()
				countMap[a]++
				countMu.Unlock()
				// do NOT return; keep streaming until server closes or ctx cancels
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
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Give me your grep command: ")
	message, _ := reader.ReadString('\n')
	message = strings.TrimSpace(message)

	parts := strings.Fields(message)
	if len(parts) == 0 {
		fmt.Println("Incorrect grep command:")
	}
	if !slices.Contains(parts, "-c") {
		countFlag = true
	}

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

	// Read hosts.txt and build address list
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

	// Create a context that can be cancelled by Ctrl+C.
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
			// fmt.Printf("[%s] error: %v\n", r.addr, r.err)
			continue
		}
		fmt.Printf("[%s from %s] response:\n%s\n", r.addr, r.file_name, r.resp)

		if !countFlag {
			// count, err := strconv.Atoi(strings.TrimSpace(r.resp))
			count, err := strconv.Atoi(strings.Split(strings.TrimSpace(r.resp), ":")[1])
			if err != nil {
				fmt.Println("Cannot convert response to integer")
			}
			total_count += count
		}

	}
	end := time.Since(start)

	if countFlag {
		countMu.Lock()
		for key, value := range countMap {
			fmt.Println(key, value)
			total_count += value
		}
		countMu.Unlock()

	}

	fmt.Println("Total count is: ", total_count)
	fmt.Println("Total time taken for last byte in last vm: ", end)

	os.Exit(0)
}
