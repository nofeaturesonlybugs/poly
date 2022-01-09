package poly_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/nofeaturesonlybugs/poly"
	"github.com/stretchr/testify/assert"
)

const (
	ContentTypeFormEncoded = "application/x-www-form-urlencoded"
	ContentTypeJSON        = "application/json"
)

// BufferedResponseWriter forces poly.Handler to make multiple Write calls
// when responding.
type BufferedResponseWriter struct {
	http.ResponseWriter
	B   *bufio.Writer
	Err error
}

func (b BufferedResponseWriter) Write(buf []byte) (int, error) {
	if b.Err != nil {
		return 0, b.Err
	}
	return b.B.Write(buf)
}

// ErrorReader returns errors from Read().
type ErrorReader struct {
	Err error
}

func (r ErrorReader) Read([]byte) (int, error) {
	return 0, r.Err
}

func TestHandler_CodeCoverage(t *testing.T) {
	p := poly.Poly{
		FormMapper:  poly.DefaultFormMapper,
		PathMapper:  poly.DefaultPathMapper,
		QueryMapper: poly.DefaultQueryMapper,
	}

	t.Run("no args", func(t *testing.T) {
		chk := assert.New(t)
		// Handler with no arguments.
		str := ""
		fn := func() {
			str = "no args"
		}
		h := p.Handler(fn)
		//
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		//
		chk.Equal("", str)
		h.ServeHTTP(w, req)
		chk.Equal("no args", str)
	})

	t.Run("http.Handler", func(t *testing.T) {
		// Poly.Handler on http.Handler bypasses all of the wrapped logic and just
		// returns the handler.
		chk := assert.New(t)
		str := ""
		mux := http.NewServeMux()
		mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			str = "http.Handler"
		}))
		h := p.Handler(mux)
		chk.Equal(mux, h)
		//
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		//
		chk.Equal("", str)
		h.ServeHTTP(w, req)
		chk.Equal("http.Handler", str)
	})

	t.Run("http.HandlerFunc", func(t *testing.T) {
		// Poly.Handler on http.HandlerFunc bypasses all of the wrapped logic but
		// returns http.Handler
		chk := assert.New(t)
		str := ""
		fn := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			str = "http.HandlerFunc"
		})
		h := p.Handler(fn)
		//
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		//
		chk.Equal("", str)
		h.ServeHTTP(w, req)
		chk.Equal("http.HandlerFunc", str)
	})

	t.Run("func(ResponseWriter, Request)", func(t *testing.T) {
		// Poly.Handler on http.Handler-compatible bypasses all of the wrapped logic but
		// returns http.Handler
		chk := assert.New(t)
		str := ""
		fn := func(w http.ResponseWriter, req *http.Request) {
			str = "func(ResponseWriter, Request)"
		}
		h := p.Handler(fn)
		//
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		//
		chk.Equal("", str)
		h.ServeHTTP(w, req)
		chk.Equal("func(ResponseWriter, Request)", str)
	})

	t.Run("func(ResponseWriter, Request, ...)", func(t *testing.T) {
		// Handler includes standard arguments plus additional arguments.
		chk := assert.New(t)
		str := ""
		fn := func(w http.ResponseWriter, req *http.Request, i int, f float32) {
			str = fmt.Sprintf("func(ResponseWriter, Request, i=%v, f=%v)", i, f)
		}
		h := p.Handler(fn)
		//
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		//
		chk.Equal("", str)
		h.ServeHTTP(w, req)
		chk.Equal("func(ResponseWriter, Request, i=0, f=0)", str)
	})

	t.Run("returns error", func(t *testing.T) {
		chk := assert.New(t)
		// Handler that returns an error.
		fn := func() (string, error) {
			return "Hello", nil
		}
		h := p.Handler(fn)
		//
		buf := &bytes.Buffer{}
		w := httptest.NewRecorder()
		w.Body = buf
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		//
		h.ServeHTTP(w, req)
		chk.Equal("Hello", buf.String())
		//
		// Now repeat similar but return the error
		fn = func() (string, error) {
			return "", fmt.Errorf("returns error")
		}
		h = p.Handler(fn)
		//
		buf = &bytes.Buffer{}
		w = httptest.NewRecorder()
		w.Body = buf
		req = httptest.NewRequest(http.MethodGet, "/", nil)
		//
		h.ServeHTTP(w, req)
		chk.Equal(http.StatusInternalServerError, w.Code)
		chk.Equal("returns error", strings.TrimSpace(buf.String()))
	})

}

func TestHandler_ReturnsString(t *testing.T) {
	p := poly.Poly{
		FormMapper:  poly.DefaultFormMapper,
		PathMapper:  poly.DefaultPathMapper,
		QueryMapper: poly.DefaultQueryMapper,
	}

	type Test struct {
		Name   string
		Fn     interface{}
		ReqFn  func() *http.Request
		Expect interface{}
	}
	type T struct {
		Name string
		Age  int
	}

	tests := []Test{
		{
			Name: "simple",
			Fn: func() string {
				return "Hello, World!"
			},
			ReqFn: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/", nil)
			},
			Expect: "Hello, World!",
		},
		{
			Name: "json",
			Fn: func(t T) string {
				return fmt.Sprintf("%v %v", t.Name, t.Age)
			},
			ReqFn: func() *http.Request {
				blob, _ := json.Marshal(T{"Fred", 42})
				req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(blob))
				req.Header.Set("Content-Type", ContentTypeJSON)
				return req
			},
			Expect: "Fred 42",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			chk := assert.New(t)
			h := p.Handler(test.Fn)
			//
			buf := &bytes.Buffer{}
			w := httptest.NewRecorder()
			w.Body = buf
			req := test.ReqFn()
			//
			h.ServeHTTP(w, req)
			// TODO Check status?
			// TODO Check content-type
			chk.Equal(test.Expect, buf.String())
		})
	}

	t.Run("multiple writes", func(t *testing.T) {
		chk := assert.New(t)
		h := p.Handler(func() string {
			return "Hello, World!"
		})
		//
		buf := &bytes.Buffer{}
		rw := httptest.NewRecorder()
		rw.Body = buf
		w := BufferedResponseWriter{
			ResponseWriter: httptest.NewRecorder(),
			B:              bufio.NewWriterSize(rw, 2),
		}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		//
		h.ServeHTTP(w, req)
		// TODO Check status?
		// TODO Check content-type
		chk.Equal("Hello, World!", buf.String())
	})

	t.Run("multiple writes has error", func(t *testing.T) {
		chk := assert.New(t)
		h := p.Handler(func() string {
			return "Hello, World!"
		})
		//
		err := fmt.Errorf("buffered error")
		buf := &bytes.Buffer{}
		rw := httptest.NewRecorder()
		rw.Body = buf
		w := BufferedResponseWriter{
			ResponseWriter: httptest.NewRecorder(),
			B:              bufio.NewWriterSize(rw, 2),
			Err:            err,
		}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		//
		h.ServeHTTP(w, req)
		// TODO Check content-type
		chk.Equal("", buf.String())
	})

}

func TestHandler_ReturnsJSON(t *testing.T) {
	p := poly.Poly{
		FormMapper:  poly.DefaultFormMapper,
		PathMapper:  poly.DefaultPathMapper,
		QueryMapper: poly.DefaultQueryMapper,
	}

	type Test struct {
		Name   string
		Fn     interface{}
		Dest   interface{}
		Expect interface{}
	}
	type T struct {
		Name string
		Age  int
	}

	tests := []Test{
		{
			Name: "struct",
			Fn: func() T {
				return T{"Fred", 42}
			},
			Dest:   &T{},
			Expect: T{"Fred", 42},
		},
		{
			Name: "*struct",
			Fn: func() *T {
				return &T{"Barney", 78}
			},
			Dest:   &T{},
			Expect: T{"Barney", 78},
		},
		{
			Name: "map",
			Fn: func() map[string]interface{} {
				return map[string]interface{}{
					"Name": "Betty",
					"Age":  99,
				}
			},
			Dest:   &T{},
			Expect: T{"Betty", 99},
		},
		{
			Name: "array",
			Fn: func() [2]T {
				return [...]T{
					{"Flim", 10},
					{"Flam", 20},
				}
			},
			Dest: &[]T{},
			Expect: []T{
				{"Flim", 10},
				{"Flam", 20},
			},
		},
		{
			Name: "slice",
			Fn: func() []T {
				return []T{
					{"Fred", 42},
					{"Barney", 78},
				}
			},
			Dest: &[]T{},
			Expect: []T{
				{"Fred", 42},
				{"Barney", 78},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			chk := assert.New(t)
			h := p.Handler(test.Fn)
			//
			buf := &bytes.Buffer{}
			w := httptest.NewRecorder()
			w.Body = buf
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			//
			h.ServeHTTP(w, req)
			// TODO Check status?
			// TODO Check content-type (waiting to decide if application/json)
			err := json.Unmarshal(buf.Bytes(), test.Dest)
			chk.NoError(err)
			chk.Equal(test.Expect, reflect.Indirect(reflect.ValueOf(test.Dest)).Interface())
		})
	}
}

func TestHandler_ReturnsJSONErrors(t *testing.T) {
	p := poly.Poly{
		FormMapper:  poly.DefaultFormMapper,
		PathMapper:  poly.DefaultPathMapper,
		QueryMapper: poly.DefaultQueryMapper,
	}

	type Test struct {
		Name string
		Fn   interface{}
		Code int
	}
	type T struct {
		Name string
		Age  int
		C    complex128
	}

	tests := []Test{
		{
			Name: "struct",
			Fn: func() T {
				return T{"Fr\x00ed", 42, complex(10, 11)}
			},
			Code: http.StatusInternalServerError,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			chk := assert.New(t)
			h := p.Handler(test.Fn)
			//
			buf := &bytes.Buffer{}
			w := httptest.NewRecorder()
			w.Body = buf
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			//
			h.ServeHTTP(w, req)
			chk.Equal(test.Code, w.Code)
		})
	}
}

func TestHandler_Unmarshals(t *testing.T) {
	kv := poly.KeyValueParams{}
	p := poly.Poly{
		FormMapper:  poly.DefaultFormMapper,
		PathMapper:  poly.DefaultPathMapper,
		QueryMapper: poly.DefaultQueryMapper,

		PathParamer: kv,
	}

	type Test struct {
		Name   string
		Fn     interface{}
		ReqFn  func() *http.Request
		Dest   interface{}
		Expect interface{}
	}
	type T struct {
		Name string `json:"name" form:"name" query:"name" path:"Name"`
		Age  int    `json:"age" form:"age" query:"age" path:"Age"`
	}

	tests := []Test{
		{
			Name: "form",
			Fn: func(t T) T {
				return t
			},
			ReqFn: func() *http.Request {
				form := url.Values{
					"name": []string{"Fred"},
					"age":  []string{"42"},
				}
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", ContentTypeFormEncoded)
				return req
			},
			Dest:   &T{},
			Expect: T{"Fred", 42},
		},
		{
			Name: "json",
			Fn: func(t T) T {
				return t
			},
			ReqFn: func() *http.Request {
				blob, _ := json.Marshal(T{"Fred", 42})
				req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(blob))
				req.Header.Set("Content-Type", ContentTypeJSON)
				return req
			},
			Dest:   &T{},
			Expect: T{"Fred", 42},
		},
		{
			Name: "json ptr",
			Fn: func(t *T) *T {
				return t
			},
			ReqFn: func() *http.Request {
				blob, _ := json.Marshal(T{"Barney", 80})
				req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(blob))
				req.Header.Set("Content-Type", ContentTypeJSON)
				return req
			},
			Dest:   &T{},
			Expect: T{"Barney", 80},
		},
		{
			Name: "path",
			Fn: func(t T) T {
				return t
			},
			ReqFn: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/Name/Fred/Age/42", nil)
				return req
			},
			Dest:   &T{},
			Expect: T{"Fred", 42},
		},
		{
			Name: "query",
			Fn: func(t T) T {
				return t
			},
			ReqFn: func() *http.Request {
				form := url.Values{
					"name": []string{"Fred"},
					"age":  []string{"42"},
				}
				req := httptest.NewRequest(http.MethodGet, "/?"+form.Encode(), nil)
				return req
			},
			Dest:   &T{},
			Expect: T{"Fred", 42},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			chk := assert.New(t)
			h := kv.Handler(p.Handler(test.Fn))
			//
			buf := &bytes.Buffer{}
			w := httptest.NewRecorder()
			w.Body = buf
			req := test.ReqFn()
			//
			h.ServeHTTP(w, req)
			// TODO Check status?
			// TODO Check content-type (waiting to decide if application/json)
			err := json.Unmarshal(buf.Bytes(), test.Dest)
			chk.NoError(err)
			chk.Equal(test.Expect, reflect.Indirect(reflect.ValueOf(test.Dest)).Interface())
		})
	}
}

func TestHandler_UnmarshalErrors(t *testing.T) {
	p := poly.Poly{
		FormMapper:  poly.DefaultFormMapper,
		PathMapper:  poly.DefaultPathMapper,
		QueryMapper: poly.DefaultQueryMapper,
	}

	type Test struct {
		Name       string
		Fn         interface{}
		ReqFn      func() *http.Request
		Middleware func(http.Handler) http.Handler
		Code       int
	}
	type T struct {
		Name string `json:"name" form:"name" query:"name" path:"Name"`
		Age  int    `json:"age" form:"age" query:"age" path:"Age"`
	}

	tests := []Test{
		{
			Name: "form parseform error",
			Fn: func(t T) T {
				return t
			},
			ReqFn: func() *http.Request {
				form := url.Values{
					"name": []string{"Fred"},
					"age":  []string{"42"},
				}
				// By placing an invalid escape sequence in the body we can force
				// the req.ParseForm() to return err != nil.
				body := strings.Replace(form.Encode(), "=", "%2z", -1)
				//
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
				req.Header.Set("Content-Type", ContentTypeFormEncoded)
				return req
			},
			Code: http.StatusBadRequest,
		},
		{
			Name: "json unmarshal error",
			Fn: func(t T) T {
				return t
			},
			ReqFn: func() *http.Request {
				blob, _ := json.Marshal(T{"Fred", 42})
				// By replacing { with ! we invalidate the JSON and json.Unmarshal should
				// return err != nil
				body := strings.Replace(string(blob), "{", "!", -1)
				//
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
				req.Header.Set("Content-Type", ContentTypeJSON)
				return req
			},
			Code: http.StatusBadRequest,
		},
		{
			Name: "json readbody error",
			Fn: func(t T) T {
				return t
			},
			ReqFn: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Content-Type", ContentTypeJSON)
				return req
			},
			Middleware: func(next http.Handler) http.Handler {
				// We use a middleare to replace the req.Body with one that returns
				// error during Read.
				fn := func(w http.ResponseWriter, req *http.Request) {
					rdr := ErrorReader{
						Err: fmt.Errorf("error reader"),
					}
					req.Body = io.NopCloser(rdr)
					next.ServeHTTP(w, req)
				}
				return http.HandlerFunc(fn)
			},
			Code: http.StatusBadRequest,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			chk := assert.New(t)
			h := p.Handler(test.Fn)
			if test.Middleware != nil {
				h = test.Middleware(h)
			}
			//
			buf := &bytes.Buffer{}
			w := httptest.NewRecorder()
			w.Body = buf
			req := test.ReqFn()
			//
			h.ServeHTTP(w, req)
			// TODO Check status?
			// TODO Check content-type (waiting to decide if application/json)
			chk.Equal(test.Code, w.Code)
		})
	}
}
