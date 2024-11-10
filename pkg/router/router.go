package router

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type HandlerFunc func(*Context)

func NewHandlerCache() *HandlerCache {
	return &HandlerCache{
		cache: make(map[string]CachedHandler),
	}
}

type node struct {
	children   map[string]*node
	handler    map[string]HandlerFunc
	paramName  string
	isParam    bool
	isWildcard bool
}

type Router struct {
	tree             *node
	middlewares      []MiddlewareFunc
	notFound         HandlerFunc
	methodNotAllowed HandlerFunc
	cache            *HandlerCache
}

type CachedHandler struct {
	Handler HandlerFunc
	Params  map[string]string
}

type HandlerCache struct {
	cache map[string]CachedHandler
	mu    sync.RWMutex
}

func (hc *HandlerCache) Get(path string) (HandlerFunc, map[string]string, bool) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	cached, ok := hc.cache[path]
	return cached.Handler, cached.Params, ok
}

func (hc *HandlerCache) Set(path string, handler HandlerFunc, params map[string]string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.cache[path] = CachedHandler{Handler: handler, Params: params}
}

func NewRouter() *Router {
	return &Router{
		tree: &node{
			children: make(map[string]*node),
			handler:  make(map[string]HandlerFunc),
		},
		notFound: func(c *Context) {
			c.String(http.StatusNotFound, "404 page not found")
		},
		methodNotAllowed: func(c *Context) {
			c.String(http.StatusMethodNotAllowed, "405 method not allowed")
		},
		cache: NewHandlerCache(),
	}
}

func (r *Router) addRoute(method, path string, handlers ...HandlerFunc) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	current := r.tree
	for i, part := range parts {
		if len(part) > 0 && part[0] == ':' {
			if current.isWildcard {
				panic("router: parameter after wildcard not allowed")
			}
			current.isParam = true
			current.paramName = part[1:]
			if _, ok := current.children["*param"]; !ok {
				current.children["*param"] = &node{
					children: make(map[string]*node),
					handler:  make(map[string]HandlerFunc),
				}
			}
			current = current.children["*param"]
		} else if part == "*" {
			if i != len(parts)-1 {
				panic("router: wildcard must be the last part of the path")
			}
			current.isWildcard = true
			break
		} else {
			if _, ok := current.children[part]; !ok {
				current.children[part] = &node{
					children: make(map[string]*node),
					handler:  make(map[string]HandlerFunc),
				}
			}
			current = current.children[part]
		}
	}

	if current.handler[method] != nil {
		panic("router: duplicate route")
	}

	// Combine all handlers into a single HandlerFunc
	current.handler[method] = func(c *Context) {
		for _, h := range handlers {
			h(c)
			if c.IsAborted() {
				break
			}
		}
	}
}

func (r *Router) GET(path string, handler HandlerFunc) {
	r.addRoute(http.MethodGet, path, handler)
}

func (r *Router) POST(path string, handler HandlerFunc) {
	r.addRoute(http.MethodPost, path, handler)
}

func (r *Router) PUT(path string, handler HandlerFunc) {
	r.addRoute(http.MethodPut, path, handler)
}

func (r *Router) DELETE(path string, handler HandlerFunc) {
	r.addRoute(http.MethodDelete, path, handler)
}

func (r *Router) PATCH(path string, handler HandlerFunc) {
	r.addRoute(http.MethodPatch, path, handler)
}

func (r *Router) Group(prefix string) *RouterGroup {
	return &RouterGroup{
		prefix: prefix,
		router: r,
	}
}

func (r *Router) Use(middleware ...MiddlewareFunc) {
	r.middlewares = append(r.middlewares, middleware...)
}

func (r *Router) find(method, path string) (HandlerFunc, map[string]string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	current := r.tree
	params := make(map[string]string)

	for _, part := range parts {
		if child, ok := current.children[part]; ok {
			current = child
		} else if current.isParam {
			params[current.paramName] = part
			current = current.children["*param"]
		} else if current.isWildcard {
			if handler, ok := current.handler[method]; ok {
				return handler, params
			}
			return nil, params
		} else {
			fmt.Println("No match found")
			return nil, nil
		}
	}

	if handler, ok := current.handler[method]; ok {
		return handler, params
	}
	return nil, params
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var handler HandlerFunc
	var params map[string]string

	c := newContext(w, req)

	// If not in cache, find the handler and params
	handler, params = r.find(req.Method, req.URL.Path)
	// Cache the handler and params for future use
	//r.cache.Set(req.URL.Path, handler, params)

	c.Params = params

	if handler == nil {
		r.methodNotAllowed(c)
		return
	}

	if handler = r.applyMiddleware(handler); handler != nil {
		handler(c)
	}
}

func (r *Router) applyMiddleware(handler HandlerFunc) HandlerFunc {
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		handler = r.middlewares[i](handler)
	}
	return handler
}

func (r *Router) PrintRoutes() {
	var printNode func(*node, string)
	printNode = func(n *node, prefix string) {
		for _, handler := range n.handler {
			if handler != nil {
			}
		}
		for part, child := range n.children {
			printNode(child, prefix+"/"+part)
		}
		if n.isParam {
			printNode(n.children["*param"], prefix+"/:"+n.paramName)
		}
		if n.isWildcard {
		}
	}
	printNode(r.tree, "")
}

func authMiddleware(c *Context) {
	// Implement authentication logic here
	// For example, check for a valid token in the request header
	token := c.GetHeader("Authorization")
	if token == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Validate the token
	// If valid, call c.Next() to continue to the next handler
	// If invalid, abort the request with an appropriate status code

	// For this example, we'll just check if the token is "valid_token"
	if token != "valid_token" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Next()
}

func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}
