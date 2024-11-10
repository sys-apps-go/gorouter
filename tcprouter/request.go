package simplehttp

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// Request represents a simplified HTTP request.
type Request struct {
	Method  string
	URI     string
	Headers map[string]string
	Conn    net.Conn
	Reader  *bufio.Reader
}

// parseRequest parses an HTTP request from the reader.
func parseRequest(reader *bufio.Reader, conn net.Conn) (*Request, error) {
	// Read the request line.
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)
	parts := strings.Split(line, " ")
	if len(parts) < 3 {
		return nil, fmt.Errorf("malformed request line")
	}
	method, uri := parts[0], parts[1]

	// Read headers.
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" { // End of headers
			break
		}
		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) != 2 {
			continue // Skip malformed headers
		}
		headers[strings.TrimSpace(headerParts[0])] = strings.TrimSpace(headerParts[1])
	}

	return &Request{
		Method:  method,
		URI:     uri,
		Headers: headers,
		Conn:    conn,
		Reader:  reader,
	}, nil
}

