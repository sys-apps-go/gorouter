package router

import (
	"net/http"
	"path"
	"strings"
)

// RouterGroup is used for grouping routes with common prefix and middleware
type RouterGroup struct {
	prefix      string
	parent      *RouterGroup
	router      *Router
	middlewares []HandlerFunc
}

// Group creates a new router group

func (group *RouterGroup) Group(prefix string) *RouterGroup {
	return &RouterGroup{
		prefix:      group.prefix + prefix,
		parent:      group,
		router:      group.router,
		middlewares: []HandlerFunc{},
	}
}

func (group *RouterGroup) Use(middleware ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middleware...)
}

// GET registers a new GET route for a path with handler
func (group *RouterGroup) GET(relativePath string, handlers ...HandlerFunc) {
	group.handle(http.MethodGet, relativePath, handlers)
}

// POST registers a new POST route for a path with handler
func (group *RouterGroup) POST(relativePath string, handlers ...HandlerFunc) {
	group.handle(http.MethodPost, relativePath, handlers)
}

// PUT registers a new PUT route for a path with handler
func (group *RouterGroup) PUT(relativePath string, handlers ...HandlerFunc) {
	group.handle(http.MethodPut, relativePath, handlers)
}

// DELETE registers a new DELETE route for a path with handler
func (group *RouterGroup) DELETE(relativePath string, handlers ...HandlerFunc) {
	group.handle(http.MethodDelete, relativePath, handlers)
}

// PATCH registers a new PATCH route for a path with handler
func (group *RouterGroup) PATCH(relativePath string, handlers ...HandlerFunc) {
	group.handle(http.MethodPatch, relativePath, handlers)
}

// HEAD registers a new HEAD route for a path with handler
func (group *RouterGroup) HEAD(relativePath string, handlers ...HandlerFunc) {
	group.handle(http.MethodHead, relativePath, handlers)
}

// OPTIONS registers a new OPTIONS route for a path with handler
func (group *RouterGroup) OPTIONS(relativePath string, handlers ...HandlerFunc) {
	group.handle(http.MethodOptions, relativePath, handlers)
}

// handle registers a new route for a path with matching method and handlers
func (group *RouterGroup) handle(httpMethod, relativePath string, handlers []HandlerFunc) {
	absolutePath := group.calculateAbsolutePath(relativePath)
	handlers = group.combineHandlers(handlers)
	group.router.addRoute(httpMethod, absolutePath, handlers...)
}

// calculateAbsolutePath returns absolute path of current group combined with given relative path
func (group *RouterGroup) calculateAbsolutePath(relativePath string) string {
	if relativePath == "" {
		return group.prefix
	}
	absolutePath := path.Join(group.prefix, relativePath)
	// Append a trailing slash if the last component had one
	if strings.HasSuffix(relativePath, "/") && !strings.HasSuffix(absolutePath, "/") {
		return absolutePath + "/"
	}
	return absolutePath
}

// combineHandlers returns handlers combined with group middleware
func (group *RouterGroup) combineHandlers(handlers []HandlerFunc) []HandlerFunc {
	finalSize := len(group.middlewares) + len(handlers)
	mergedHandlers := make([]HandlerFunc, finalSize)
	copy(mergedHandlers, group.middlewares)
	copy(mergedHandlers[len(group.middlewares):], handlers)
	return mergedHandlers
}

// Static serves files from the given file system root
func (group *RouterGroup) Static(relativePath, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")

	// Register GET and HEAD handlers
	group.GET(urlPattern, handler)
	group.HEAD(urlPattern, handler)
}

// createStaticHandler creates a handler to serve static files
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := group.calculateAbsolutePath(relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))

	return func(c *Context) {
		file := c.Param("filepath")
		// Check if file exists and/or if we have permission to access it
		f, err := fs.Open(file)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		f.Close()

		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}
