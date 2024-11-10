package main

import (
	"github.com/sys-apps-go/gorouter/simplehttp"
	"log"
	"fmt"
)

func main() {
	// Define server addresses
	httpAddr := ":50051"
	httpsAddr := ":50052"
	workerCount := 100

	// Initialize the HTTP server
	httpServer := simplehttp.NewServer(httpAddr, defaultHandler, workerCount)

	// Initialize the HTTPS server
	httpsServer := simplehttp.NewServer(httpsAddr, defaultHandler, workerCount)

	// Set up TLS configuration for HTTPS
	certFile := "localhost.crt"
	keyFile := "localhost.key"
	err := httpsServer.SetTLSConfig(certFile, keyFile)
	if err != nil {
		log.Fatalf("Error setting up TLS: %v", err)
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Printf("Starting HTTP server on %s", httpAddr)
		if err := httpServer.Start(); err != nil {
			log.Fatalf("HTTP Server failed: %v", err)
		}
	}()

	// Start HTTPS server in the main goroutine
	log.Printf("Starting HTTPS server on %s", httpsAddr)
	if err := httpsServer.Start(); err != nil {
		log.Fatalf("HTTPS Server failed: %v", err)
	}
}

func defaultHandler(w *simplehttp.ResponseWriter, r *simplehttp.Request) {
	body := "Hello!\n"

	headers := map[string]string{
		"Content-Type":   "text/plain",
		"Content-Length": fmt.Sprintf("%d", len(body)),
		"Connection":     "keep-alive",
	}

	response := fmt.Sprintf("HTTP/1.1 200 OK\r\n")
	for key, value := range headers {
		response += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	response += "\r\n" + body

	_, err := w.Write([]byte(response))
	if err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
