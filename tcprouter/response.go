package simplehttp

import "net"

// ResponseWriter is a simplified response writer.
type ResponseWriter struct {
	Conn net.Conn
}

// Write sends data to the client.
func (w *ResponseWriter) Write(data []byte) (int, error) {
	return w.Conn.Write(data)
}

