package poly_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nofeaturesonlybugs/poly"
	"github.com/stretchr/testify/assert"
)

func TestKeyValueParams_Handler(t *testing.T) {
	kv := poly.KeyValueParams{}

	type Test struct {
		Name   string
		ReqFn  func() *http.Request
		Expect string
	}

	tests := []Test{
		{
			Name: "simple",
			ReqFn: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/Name/Fred/Age/42", nil)
				return req
			},
			Expect: "Fred 42",
		},
		{
			Name: "dangling",
			ReqFn: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/Name/Fred/Age", nil)
				return req
			},
			Expect: "Fred ",
		},
		{
			Name: "last wins",
			ReqFn: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/Name/Fred/Age/42/Name/Barney/Age/38", nil)
				return req
			},
			Expect: "Barney 38",
		},
		{
			Name: "trailing slash",
			ReqFn: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/Name/Fred/Age/42/", nil)
				return req
			},
			Expect: "Fred 42",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			chk := assert.New(t)
			//
			var h http.Handler
			var str string
			h = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				name := kv.PathParam(req, "Name")
				age := kv.PathParam(req, "Age")
				str = fmt.Sprintf("%v %v", name, age)
			})
			h = kv.Handler(h)
			//
			w := httptest.NewRecorder()
			req := test.ReqFn()
			//
			h.ServeHTTP(w, req)
			//
			chk.Equal(test.Expect, str)
		})
	}

	t.Run("without middleware", func(t *testing.T) {
		chk := assert.New(t)
		//
		var str string
		h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			name := kv.PathParam(req, "Name")
			age := kv.PathParam(req, "Age")
			str = fmt.Sprintf("%v %v", name, age)
		})
		//
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/Name/Fred/Age/42", nil)
		//
		h.ServeHTTP(w, req)
		//
		chk.Equal(" ", str)
	})
}

func TestKeyValueParams_ParsePath(t *testing.T) {
	kv := poly.KeyValueParams{}

	type Test struct {
		Name   string
		Path   string
		Expect map[string]string
	}

	tests := []Test{
		{
			Name:   "empty",
			Path:   "",
			Expect: map[string]string{},
		},
		{
			Name:   "slash",
			Path:   "/",
			Expect: map[string]string{},
		},
		{
			Name:   "slashes",
			Path:   "///",
			Expect: map[string]string{},
		},

		{
			Name:   "key",
			Path:   "Name",
			Expect: map[string]string{"Name": ""},
		},
		{
			Name:   "/key",
			Path:   "/Name",
			Expect: map[string]string{"Name": ""},
		},
		{
			Name:   "key/",
			Path:   "Name/",
			Expect: map[string]string{"Name": ""},
		},
		{
			Name:   "/key/",
			Path:   "/Name/",
			Expect: map[string]string{"Name": ""},
		},

		{
			Name:   "key/value",
			Path:   "Name/Larry",
			Expect: map[string]string{"Name": "Larry"},
		},
		{
			Name:   "/key/value",
			Path:   "/Name/Larry",
			Expect: map[string]string{"Name": "Larry"},
		},
		{
			Name:   "key/value/",
			Path:   "Name/Larry/",
			Expect: map[string]string{"Name": "Larry"},
		},
		{
			Name:   "/key/value/",
			Path:   "/Name/Larry/",
			Expect: map[string]string{"Name": "Larry"},
		},

		{
			Name:   "key/value/key/value",
			Path:   "Name/Larry/Age/42",
			Expect: map[string]string{"Name": "Larry", "Age": "42"},
		},
		{
			Name:   "/key/value/key/value",
			Path:   "/Name/Larry/Age/42",
			Expect: map[string]string{"Name": "Larry", "Age": "42"},
		},
		{
			Name:   "key/value/key/value/",
			Path:   "Name/Larry/Age/42/",
			Expect: map[string]string{"Name": "Larry", "Age": "42"},
		},
		{
			Name:   "/key/value/key/value/",
			Path:   "/Name/Larry/Age/42/",
			Expect: map[string]string{"Name": "Larry", "Age": "42"},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			chk := assert.New(t)
			m := kv.ParsePath(test.Path)
			chk.Equal(test.Expect, m)
		})
	}
}
