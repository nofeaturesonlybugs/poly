package poly_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/nofeaturesonlybugs/poly"
	"github.com/nofeaturesonlybugs/poly/examples"
)

func BenchmarkPoly_Echo(b *testing.B) {
	p := poly.Poly{}
	mux := http.NewServeMux()

	mux.Handle("/stdlib", http.HandlerFunc(examples.STDEcho))
	mux.Handle("/poly", p.Handler(examples.Echo))

	b.Run("stdlib", func(b *testing.B) {
		var message examples.EchoRequest
		var w *httptest.ResponseRecorder
		var req *http.Request
		var body []byte
		var err error
		//
		message = examples.EchoRequest{Message: "stdlib echo"}
		body, err = json.Marshal(message)
		if err != nil {
			panic(err)
		}
		//
		for k := 0; k < b.N; k++ {
			w = httptest.NewRecorder()
			w.Body = &bytes.Buffer{}
			req = httptest.NewRequest(http.MethodPost, "/stdlib", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			mux.ServeHTTP(w, req)
			if http.StatusOK != w.Code {
				b.Fatalf("unexpected status %v", w.Code)
			}
			if "stdlib echo" != w.Body.String() {
				b.Fatalf("unexpected response %v", w.Body.String())
			}
		}
	})

	b.Run("poly", func(b *testing.B) {
		var message examples.EchoRequest
		var w *httptest.ResponseRecorder
		var req *http.Request
		var body []byte
		var err error
		//
		message = examples.EchoRequest{Message: "poly echo"}
		body, err = json.Marshal(message)
		if err != nil {
			panic(err)
		}
		//
		for k := 0; k < b.N; k++ {
			w = httptest.NewRecorder()
			w.Body = &bytes.Buffer{}
			req = httptest.NewRequest(http.MethodPost, "/poly", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			mux.ServeHTTP(w, req)
			if http.StatusOK != w.Code {
				b.Fatalf("unexpected status %v", w.Code)
			}
			if "poly echo" != w.Body.String() {
				b.Fatalf("unexpected response %v", w.Body.String())
			}
		}
	})
}

func BenchmarkPoly_Login(b *testing.B) {
	p := poly.Poly{
		FormMapper: poly.DefaultFormMapper,
	}
	mux := http.NewServeMux()

	mux.Handle("/stdlib", http.HandlerFunc(examples.STDLogin))
	mux.Handle("/poly", p.Handler(examples.Login))

	b.Run("stdlib", func(b *testing.B) {
		var w *httptest.ResponseRecorder
		var req *http.Request
		//
		form := url.Values{
			"username": []string{"nofeaturesonlybugs"},
			"password": []string{"hunter2"},
		}
		//
		for k := 0; k < b.N; k++ {
			w = httptest.NewRecorder()
			w.Body = &bytes.Buffer{}
			req = httptest.NewRequest(http.MethodPost, "/stdlib", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			mux.ServeHTTP(w, req)
			if http.StatusOK != w.Code {
				b.Fatalf("unexpected status %v", w.Code)
			}
		}
	})

	b.Run("poly", func(b *testing.B) {
		var w *httptest.ResponseRecorder
		var req *http.Request
		//
		form := url.Values{
			"username": []string{"nofeaturesonlybugs"},
			"password": []string{"hunter2"},
		}
		//
		for k := 0; k < b.N; k++ {
			w = httptest.NewRecorder()
			w.Body = &bytes.Buffer{}
			req = httptest.NewRequest(http.MethodPost, "/poly", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			mux.ServeHTTP(w, req)
			if http.StatusOK != w.Code {
				b.Fatalf("unexpected status %v", w.Code)
			}
		}
	})
}
