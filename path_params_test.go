package poly_test

import (
	"net/http"
	"testing"

	"github.com/nofeaturesonlybugs/poly"
	"github.com/stretchr/testify/assert"
)

func TestPathParamFunc(t *testing.T) {
	chk := assert.New(t)
	pp := func(req *http.Request, key string) string {
		return "Hello"
	}
	pper := poly.PathParamFunc(pp)
	chk.Equal("Hello", pper.PathParam(nil, ""))
}
