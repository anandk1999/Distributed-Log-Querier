package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

type result struct {
	addr      string
	resp      string
	file_name string
	err       error
}

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
			// Send each line as a separate result
			out, file, _ := strings.Cut(scanner.Text(), " ")
			select {
			case <-ctx.Done(): // Check if the operation was cancelled.
				return
			case results <- result{addr: a, resp: out, file_name: file}:
				return
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
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Give me your grep command: ")
	message, _ := reader.ReadString('\n')
	message = strings.TrimSpace(message)

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
	for _, addr := range addresses {
		chans = append(chans, connection(addr, message, ctx))
	}
	c := fanIn(chans...)

	for r := range c {
		if r.err != nil {
			fmt.Printf("[%s] error: %v\n", r.addr, r.err)
			continue
		}
		fmt.Printf("[%s from %s] response:\n%s\n", r.addr, r.file_name, r.resp)
	}

	os.Exit(0)
}
