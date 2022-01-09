package poly

import (
	"context"
	"net/http"
)

// kvp is the Context key-type for KeyValueParams.
type kvp string

const kvpParams = kvp("params")

// KeyValueParams acts as both a middleware and a getter for URI path parmeters
// treated as key/value pairs.
//	/Size/Large/Color/Blue ->
//		Size: Large
//		Color: Blue
type KeyValueParams struct {
}

// Handler is the middleware that performs path parsing on incoming requests and
// adds them to the request's context.
func (kv KeyValueParams) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		values := kv.ParsePath(req.URL.Path)
		ctx := context.WithValue(req.Context(), kvpParams, values)
		next.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// Parse parses a path into a url.Values type.
func (kv KeyValueParams) ParsePath(p string) map[string]string {
	rv := map[string]string{}
	remaining := len(p)
	var keystr string
	for p != "" {
		var k, begin int
		for ; begin < remaining && p[begin] == '/'; begin++ {
		}
		if begin == remaining {
			break
		}

		for k = begin; k < remaining && p[k] != '/'; k++ {
		}
		keystr = p[begin:k]
		if k == remaining {
			rv[keystr] = ""
			break
		}

		for begin = k; begin < remaining && p[begin] == '/'; begin++ {
		}

		for k = begin; k < remaining && p[k] != '/'; k++ {
		}
		rv[keystr] = p[begin:k]
		if k == remaining {
			break
		}

		remaining = remaining - k
		p = p[k:]
	}
	return rv
}

// PathParam returns the value associated with key in the given request.
func (kv KeyValueParams) PathParam(req *http.Request, name string) string {
	if m := req.Context().Value(kvpParams); m != nil {
		return m.(map[string]string)[name]
	}
	return ""
}
