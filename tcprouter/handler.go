package simplehttp

// HandlerFunc defines the handler function type.
type HandlerFunc func(w *ResponseWriter, r *Request)

