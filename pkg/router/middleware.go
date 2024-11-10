package router

import (
	"log"
	"net/http"
	"time"
)

// MiddlewareFunc defines the signature of a middleware function
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// Chain applies a list of middleware to a handler function
func Chain(h HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}

// Logger is a middleware that logs the request method, URI, and duration
func Logger() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) {
			start := time.Now()

			next(c)

			log.Printf("%s %s %v", c.Request.Method, c.Request.RequestURI, time.Since(start))
		}
	}
}

// Recover is a middleware that recovers from panics and logs the error
func Recover() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("Panic: %v", err)
					c.AbortWithStatus(http.StatusInternalServerError)
				}
			}()
			next(c)
		}
	}
}

// CORS is a middleware that adds Cross-Origin Resource Sharing headers
func CORS() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) {
			c.SetHeader("Access-Control-Allow-Origin", "*")
			c.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.SetHeader("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(http.StatusOK)
				return
			}

			next(c)
		}
	}
}

// Auth is a simple authentication middleware
func Auth(authFunc func(*Context) bool) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) {
			if !authFunc(c) {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			next(c)
		}
	}
}

// RateLimiter is a simple rate limiting middleware
func RateLimiter(limit int, per time.Duration) MiddlewareFunc {
	limiter := make(map[string]int)
	ticker := time.NewTicker(per)

	go func() {
		for range ticker.C {
			limiter = make(map[string]int)
		}
	}()

	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) {
			ip := c.Request.RemoteAddr
			if limiter[ip] >= limit {
				c.AbortWithStatus(http.StatusTooManyRequests)
				return
			}
			limiter[ip]++
			next(c)
		}
	}
}

// RequestID is a middleware that adds a unique request ID to each request
func RequestID() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) {
			requestID := c.Request.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}
			c.SetHeader("X-Request-ID", requestID)
			c.Set("RequestID", requestID)
			next(c)
		}
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

// randomString generates a random string of given length
func randomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[time.Now().UnixNano()%int64(len(letterBytes))]
	}
	return string(b)
}
