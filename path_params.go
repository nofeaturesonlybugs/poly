package poly

import "net/http"

// pathParams describes a handler argument that can be unmarshaled via the URL path.
type pathParams struct {
	// N is the argument index.
	N int

	// Keys is the list of param names.
	Keys []string
}

// PathParamer returns path-param values.
type PathParamer interface {
	// PathParam returns the value for path-param key in the request.
	PathParam(req *http.Request, key string) string
}

// PathParamFunc is an adapter to allow ordinary functions to work as PathParamers.
type PathParamFunc func(*http.Request, string) string

// PathParam returns the value for path-param key in the request.
func (f PathParamFunc) PathParam(req *http.Request, key string) string {
	return f(req, key)
}
