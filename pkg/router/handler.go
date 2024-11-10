package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Context encapsulates the HTTP request and response
type Context struct {
	Writer     http.ResponseWriter
	Request    *http.Request
	Params     map[string]string
	StatusCode int
	handlers   []HandlerFunc
	index      int
	Keys       map[string]interface{}
}

var (
	contextPool = sync.Pool{
		New: func() interface{} {
			return &Context{}
		},
	}
)

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	c := contextPool.Get().(*Context)
	c.Writer = w
	c.Request = req
	c.Params = make(map[string]string)
	c.StatusCode = http.StatusOK
	c.handlers = nil
	c.index = -1
	return c
}

func (c *Context) reset() {
	c.Writer = nil
	c.Request = nil
	c.Params = nil
	c.StatusCode = http.StatusOK
	c.handlers = nil
	c.index = -1
}

// Next is used to pass control to the next middleware
func (c *Context) Next() {
	c.index++
	for c.index < len(c.handlers) {
		c.handlers[c.index](c)
		c.index++
	}
}

// Param returns the value of the URL param
func (c *Context) Param(key string) string {
	return c.Params[key]
}

// Query returns the query param for the provided key
func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

// SetHeader sets a response header
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

// Add this to your Context struct in handler.go
func (c *Context) Set(key string, value interface{}) {
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys[key] = value
}

func (c *Context) IsAborted() bool {
	return c.index >= len(c.handlers)
}

// In pkg/myrouter/handler.go or wherever your Context struct is defined

func (c *Context) BindJSON(obj interface{}) error {
	return json.NewDecoder(c.Request.Body).Decode(obj)
}

// Status sets the HTTP response status code
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

// String sends a string response
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	fmt.Fprintf(c.Writer, format, values...)
}

// JSON sends a JSON response
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

// Data sends a byte slice as the response
func (c *Context) Data(code int, contentType string, data []byte) {
	c.SetHeader("Content-Type", contentType)
	c.Status(code)
	c.Writer.Write(data)
}

// HTML sends an HTML response
func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}

// Redirect sends an HTTP redirect
func (c *Context) Redirect(code int, location string) {
	http.Redirect(c.Writer, c.Request, location, code)
}

// Abort prevents pending handlers from being called
func (c *Context) Abort() {
	c.index = len(c.handlers)
}

// AbortWithStatus calls Abort and writes the headers with the specified status code
func (c *Context) AbortWithStatus(code int) {
	c.Status(code)
	c.Abort()
}

// AbortWithJSON calls Abort and then JSON
func (c *Context) AbortWithJSON(code int, obj interface{}) {
	c.Abort()
	c.JSON(code, obj)
}

// Error is a shortcut for AbortWithStatus(500)
func (c *Context) Error(err error) {
	c.AbortWithStatus(http.StatusInternalServerError)
	c.Writer.Write([]byte(err.Error()))
}

// HandlersChain defines a HandlerFunc array.
type HandlersChain []HandlerFunc

// Last returns the last handler in the chain. ie. the last handler is the main one.
func (c HandlersChain) Last() HandlerFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}
