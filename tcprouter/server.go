package simplehttp

import (
	"bufio"
	"crypto/tls"
	"io"
	"log"
	"net"
	"sync"
)

// Server holds the server configuration.
type Server struct {
	Addr        string
	Handler     HandlerFunc
	workerCount int
	jobQueue    chan *Request
	TLSConfig   *tls.Config
	wg          sync.WaitGroup
}

// NewServer initializes a new Server.
func NewServer(addr string, handler HandlerFunc, workerCount int) *Server {
	return &Server{
		Addr:        addr,
		Handler:     handler,
		workerCount: workerCount,
		jobQueue:    make(chan *Request, 100), // Buffered channel to hold incoming requests
	}
}

// handleConnection reads requests from the connection and enqueues them for processing.
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		// Parse the HTTP request.
		req, err := parseRequest(reader, conn)
		if err != nil {
			if err != io.EOF {
			}
			return
		}

		// Enqueue the request for processing.
		s.jobQueue <- req
	}
}

// worker processes incoming requests from the job queue.
func (s *Server) worker(id int) {
	defer s.wg.Done()
	for req := range s.jobQueue {
		// Create a ResponseWriter.
		w := &ResponseWriter{Conn: req.Conn}

		// Handle the request.
		s.Handler(w, req)
	}
}

func (s *Server) SetTLSConfig(certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	s.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	return nil
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	if s.TLSConfig != nil {
		listener = tls.NewListener(listener, s.TLSConfig)
	}

	// Start worker goroutines
	for i := 0; i < s.workerCount; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}

	// Accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		// Set TCP_NODELAY to disable Nagle's algorithm.
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			err := tcpConn.SetNoDelay(true)
			if err != nil {
				log.Printf("Failed to set TCP_NODELAY: %v", err)
			}
		}

		go s.handleConnection(conn)
	}
}
