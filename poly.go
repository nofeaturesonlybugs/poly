package poly

import (
	"net/http"

	"github.com/nofeaturesonlybugs/call"
	"github.com/nofeaturesonlybugs/set"
)

// Poly is the polymorphic wrapper.
type Poly struct {
	// The following mappers determine if forms, path, or query params will
	// be unmarshaled into your handler arguments.
	// TODO Going to need more documentation here.
	FormMapper  *set.Mapper
	PathMapper  *set.Mapper
	QueryMapper *set.Mapper

	// PathParamer is the provider for path parameters.
	//
	// PathMapper!=nil means Poly will attempt to populate your
	// structs with data from the URL path.  When an argument is
	// the target for path parameter unmarshaling PathParamer(req,name)
	// is called for each parameter.
	PathParamer
}

// Handler wraps the passed function and returns an http.Handler.
func (p Poly) Handler(fn interface{}) http.Handler {
	switch h := fn.(type) {
	case http.HandlerFunc:
		return h
	case http.Handler:
		return h
	}
	//
	F := call.StatFunc(fn)
	//
	if F.NumIn == 2 && F.InTypes[0] == argTypeResponseWriter && F.InTypes[1] == argTypeRequest {
		return http.HandlerFunc(fn.(func(http.ResponseWriter, *http.Request)))
	}
	//
	return newHandler(p, F)
}
