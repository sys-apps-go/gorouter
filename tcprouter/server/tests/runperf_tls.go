package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

type connectionBlock struct {
	totalLatency   time.Duration
	averageLatency time.Duration
	totalBytes     int64
	totalRequests  int
	mu             sync.Mutex
	wg             sync.WaitGroup
}

func main() {
	c := &connectionBlock{}

	// Define command-line flags
	batchSize := flag.Int("b", 1, "Number messages in a batch")
	connections := flag.Int("c", 1, "Number of TCP connections")
	duration := flag.String("d", "30s", "Test duration (e.g., 30s)")
	url := flag.String("url", "http://localhost:50051/api/users", "URL to send requests to")
	insecure := flag.Bool("k", false, "Allow insecure server connections when using SSL")

	// Parse the flags
	flag.Parse()

	// Extract scheme, host and path from the URL
	urlParts := strings.SplitN(*url, "/", 4)
	if len(urlParts) < 4 {
		fmt.Println("Invalid URL format")
		return
	}
	scheme := urlParts[0][:len(urlParts[0])-1] // Extract the scheme (http or https)
	host := urlParts[2]                        // Extract the host (e.g., localhost:8080)
	path := "/" + urlParts[3]                  // Extract the path (e.g., /api/users)

	// Parse the duration
	testDuration, err := time.ParseDuration(*duration)
	if err != nil {
		fmt.Printf("Invalid duration format: %v\n", err)
		return
	}

	batchRequest := createBatchRequests(host, path, *batchSize)
	for i := 0; i < *connections; i++ {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			// Open a TCP or TLS connection to the host
			var conn net.Conn
			var err error
			if scheme == "https" {
				tlsConfig := &tls.Config{
					InsecureSkipVerify: *insecure,
				}
				conn, err = tls.Dial("tcp", host, tlsConfig)
			} else {
				conn, err = net.Dial("tcp", host)
			}
			if err != nil {
				fmt.Printf("Error opening connection: %v\n", err)
				return
			}
			defer conn.Close()

			// Send requests and receive responses
			startTime := time.Now()
			requestsSent := 0
			totalLatency := time.Duration(0)
			totalBytes := int64(0)

			for time.Since(startTime) < testDuration {
				// Send the HTTP request
				requestStartTime := time.Now()
				_, err = conn.Write([]byte(batchRequest))
				if err != nil {
					fmt.Printf("Error sending request: %v\n", err)
					return
				}
				requestsSent++

				// Read and process the response
				bytesRead, err := readAndProcessResponse(conn, *batchSize)
				requestEndTime := time.Now()

				if err != nil {
					if err != io.EOF {
						fmt.Printf("Error reading response: %v\n", err)
					}
					break
				}

				totalLatency += requestEndTime.Sub(requestStartTime)
				totalBytes += bytesRead
			}
			c.mu.Lock()
			c.totalLatency += totalLatency
			c.totalBytes += totalBytes
			c.totalRequests += (requestsSent * (*batchSize))
			c.mu.Unlock()
		}()
	}

	c.wg.Wait()

	averageLatency := c.totalLatency / time.Duration(c.totalRequests)

	fmt.Printf("Test completed. Sent %d requests.\n", c.totalRequests)
	fmt.Printf("Average Latency: %v\n", averageLatency)
	fmt.Printf("Total Bytes Transferred: %d Mb\n", c.totalBytes/(1024*1024))
	seconds := int(testDuration.Seconds())
	fmt.Printf("Requests per second: %v\n", c.totalRequests/seconds)
	fmt.Printf("Total Bytes Transferred per second: %d Mb\n", c.totalBytes/(1024*1024*int64(seconds)))
}

func readAndProcessResponse(conn net.Conn, n int) (int64, error) {
	reader := bufio.NewReader(conn)
	var totalBytesRead int64

	// Read headers and body
	for i := 0; i < n; i++ {
		// Read headers
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return totalBytesRead, err
			}
			totalBytesRead += int64(len(line))
			if strings.TrimSpace(line) == "" {
				break
			}
		}
	}

	return totalBytesRead, nil
}

func createBatchRequests(host string, path string, numRequests int) string {
	var requests []string

	request := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nConnection: keep-alive\r\n\r\n", path, host)
	for i := 0; i < numRequests; i++ {
		requests = append(requests, request)
	}

	return strings.Join(requests, "")
}
