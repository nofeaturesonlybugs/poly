package poly_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/nofeaturesonlybugs/poly"
	"github.com/nofeaturesonlybugs/poly/examples"
)

func ExamplePoly_forms() {
	var w *httptest.ResponseRecorder
	var req *http.Request
	var form url.Values
	b := &bytes.Buffer{}

	p := poly.Poly{
		// In order to unmarshal forms a FormMapper must be provided; this allows
		// Poly to map incoming form fields to struct fields on your handler arguments.
		FormMapper: poly.DefaultFormMapper,
	}
	mux := http.NewServeMux()

	Do := func(path string, form url.Values) {
		b.Reset()
		w = httptest.NewRecorder()
		w.Body = b
		req = httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		mux.ServeHTTP(w, req)
		fmt.Println(w.Code, strings.TrimSpace(b.String()))
	}

	//
	type OrderPizzaRequest struct {
		Size     string   `form:"size"`
		Toppings []string `form:"toppings"`
	}
	OrderPizza := func(in OrderPizzaRequest) string {
		return fmt.Sprint(in)
	}
	mux.Handle("/order-pizza", p.Handler(OrderPizza))

	//
	type CreateUserRequest struct {
		Id       int
		Username string `form:"username"`
		Password string `form:"password"`
	}
	CreateUser := func(in CreateUserRequest) string {
		in.Id = 42
		return fmt.Sprint(in)
	}
	mux.Handle("/create-user", p.Handler(CreateUser))

	form = url.Values{
		"size":     []string{"Large"},
		"toppings": []string{"Pepperoni", "Olives"},
	}
	Do("/order-pizza", form)

	form = url.Values{
		"username": []string{"nofeaturesonlybugs"},
		"password": []string{"hunter2"},
	}
	Do("/create-user", form)

	// Output: 200 {Large [Pepperoni Olives]}
	// 200 {42 nofeaturesonlybugs hunter2}
}

func ExamplePoly_jSON() {
	var w *httptest.ResponseRecorder
	var req *http.Request
	b := &bytes.Buffer{}

	p := poly.Poly{}
	mux := http.NewServeMux()

	Do := func(path string, data interface{}) {
		b.Reset()
		w = httptest.NewRecorder()
		w.Body = b
		//
		in, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}
		//
		req = httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer(in))
		req.Header.Set("Content-Type", "application/json")
		mux.ServeHTTP(w, req)
		fmt.Println(w.Code, strings.TrimSpace(b.String()))
	}

	//
	type OrderPizzaRequest struct {
		Size     string   `json:"size"`
		Toppings []string `json:"toppings"`
	}
	OrderPizza := func(in OrderPizzaRequest) string {
		return fmt.Sprint(in)
	}
	mux.Handle("/order-pizza", p.Handler(OrderPizza))

	//
	type CreateUserRequest struct {
		Id       int
		Username string `json:"username"`
		Password string `json:"password"`
	}
	CreateUser := func(in CreateUserRequest) string {
		in.Id = 42
		return fmt.Sprint(in)
	}
	mux.Handle("/create-user", p.Handler(CreateUser))

	opr := OrderPizzaRequest{
		Size:     "Large",
		Toppings: []string{"Pepperoni", "Olives"},
	}
	Do("/order-pizza", opr)

	cur := CreateUserRequest{
		Username: "nofeaturesonlybugs",
		Password: "hunter2",
	}
	Do("/create-user", cur)

	// Output: 200 {Large [Pepperoni Olives]}
	// 200 {42 nofeaturesonlybugs hunter2}
}

func ExamplePoly_pathParams() {
	// NB:  This example uses poly.KeyValueParams to exract path parameters from URLs.
	//      poly.KeyValueParams is really just a basic implementation of a path-param provider
	//      and as such adds some overhead when registering routes.
	//
	//      In practice your chosen mux or router will act as the path-param provider and
	//      you should only need to set the PathParam field when creating your Poly instance.

	var w *httptest.ResponseRecorder
	var req *http.Request
	b := &bytes.Buffer{}

	// KeyValueParams is both a middleware and path-param provider for Poly.
	kv := poly.KeyValueParams{}

	p := poly.Poly{
		// In order to unmarshal path params a PathMapper must be provided; this allows
		// Poly to map incoming path params to struct fields on your handler arguments.
		PathMapper: poly.DefaultPathMapper,

		// PathParam is also required to extract path parameters; it is a provider function
		// that accepts the request and param name and returns the value.
		PathParamer: kv,
	}
	mux := http.NewServeMux()

	Do := func(path string) {
		b.Reset()
		w = httptest.NewRecorder()
		w.Body = b
		req = httptest.NewRequest(http.MethodGet, path, nil)
		mux.ServeHTTP(w, req)
		fmt.Println(w.Code, strings.TrimSpace(b.String()))
	}

	//
	type OrderPizzaRequest struct {
		Size     string `path:"size"`
		Toppings string `path:"topping"`
	}
	OrderPizza := func(in OrderPizzaRequest) string {
		return fmt.Sprint(in)
	}
	mux.Handle("/order-pizza/", http.StripPrefix("/order-pizza", kv.Handler(p.Handler(OrderPizza))))

	//
	type CreateUserRequest struct {
		Id       int
		Username string `path:"username"`
		Password string `path:"password"`
	}
	CreateUser := func(in CreateUserRequest) string {
		in.Id = 42
		return fmt.Sprint(in)
	}
	mux.Handle("/create-user/", http.StripPrefix("/create-user", kv.Handler(p.Handler(CreateUser))))

	Do("/order-pizza/size/Large/topping/Pepperoni")

	Do("/create-user/username/nofeaturesonlybugs/password/hunter2")

	// Output: 200 {Large Pepperoni}
	// 200 {42 nofeaturesonlybugs hunter2}
}

func ExamplePoly_queryParams() {
	var w *httptest.ResponseRecorder
	var req *http.Request
	var form url.Values
	b := &bytes.Buffer{}

	p := poly.Poly{
		// In order to unmarshal a query string (GET) a QueryMapper must be provided; this allows
		// Poly to map incoming query fields to struct fields on your handler arguments.
		QueryMapper: poly.DefaultQueryMapper,
	}
	mux := http.NewServeMux()

	Do := func(path string, form url.Values) {
		b.Reset()
		w = httptest.NewRecorder()
		w.Body = b
		req = httptest.NewRequest(http.MethodGet, path+"?"+form.Encode(), nil)
		mux.ServeHTTP(w, req)
		fmt.Println(w.Code, strings.TrimSpace(b.String()))
	}

	//
	type OrderPizzaRequest struct {
		Size     string   `query:"size"`
		Toppings []string `query:"toppings"`
	}
	OrderPizza := func(in OrderPizzaRequest) string {
		return fmt.Sprint(in)
	}
	mux.Handle("/order-pizza", p.Handler(OrderPizza))

	//
	type CreateUserRequest struct {
		Id       int
		Username string `query:"username"`
		Password string `query:"password"`
	}
	CreateUser := func(in CreateUserRequest) string {
		in.Id = 42
		return fmt.Sprint(in)
	}
	mux.Handle("/create-user", p.Handler(CreateUser))

	form = url.Values{
		"size":     []string{"Large"},
		"toppings": []string{"Pepperoni", "Olives"},
	}
	Do("/order-pizza", form)

	form = url.Values{
		"username": []string{"nofeaturesonlybugs"},
		"password": []string{"hunter2"},
	}
	Do("/create-user", form)

	// Output: 200 {Large [Pepperoni Olives]}
	// 200 {42 nofeaturesonlybugs hunter2}
}

func ExamplePoly_returningPlainText() {
	var w *httptest.ResponseRecorder
	var req *http.Request
	b := &bytes.Buffer{}

	p := poly.Poly{
		FormMapper:  poly.DefaultFormMapper,
		PathMapper:  poly.DefaultPathMapper,
		QueryMapper: poly.DefaultQueryMapper,
	}
	mux := http.NewServeMux()

	Do := func(path string) {
		b.Reset()
		w = httptest.NewRecorder()
		w.Body = b
		req = httptest.NewRequest(http.MethodGet, path, nil)
		mux.ServeHTTP(w, req)
		fmt.Println(w.Code, "["+w.Header().Get("Content-Type")+"]", strings.TrimSpace(b.String()))
	}

	//
	PlainText := func() string {
		return "plain text handler"
	}
	mux.Handle("/plaintext", p.Handler(PlainText))

	// Multiple return values.
	multiple := 0
	Multiple := func() (string, error) {
		if multiple == 0 {
			multiple++
			return "No error!", nil
		}
		return "", fmt.Errorf("second call is error")
	}
	mux.Handle("/multiple", p.Handler(Multiple))

	Do("/plaintext")
	Do("/multiple")
	Do("/multiple")

	// Output: 200 [text/plain; charset=utf-8] plain text handler
	// 200 [text/plain; charset=utf-8] No error!
	// 500 [text/plain; charset=utf-8] second call is error
}

func ExamplePoly_returningJSON() {
	// If a handler returns a type T that is not a string and not an error
	// then T will be marshaled as JSON and output with content type application/json.

	var w *httptest.ResponseRecorder
	var req *http.Request
	b := &bytes.Buffer{}

	p := poly.Poly{
		FormMapper:  poly.DefaultFormMapper,
		PathMapper:  poly.DefaultPathMapper,
		QueryMapper: poly.DefaultQueryMapper,
	}
	mux := http.NewServeMux()

	Do := func(path string, dst interface{}) {
		b.Reset()
		w = httptest.NewRecorder()
		w.Body = b
		req = httptest.NewRequest(http.MethodGet, path, nil)
		mux.ServeHTTP(w, req)
		fmt.Println(w.Code, "["+w.Header().Get("Content-Type")+"]")
		if err := json.Unmarshal(b.Bytes(), dst); err != nil {
			panic(err)
		}
	}

	JSONInt := func() int {
		return 42
	}
	mux.Handle("/jsonInt", p.Handler(JSONInt))
	JSONMap := func() map[string]interface{} {
		return map[string]interface{}{
			"Message": "Hello, World!",
			"Number":  42,
		}
	}
	mux.Handle("/jsonMap", p.Handler(JSONMap))

	i := 0
	Do("/jsonInt", &i)
	fmt.Printf("  %v\n", i)

	type T struct {
		Message string
		Number  int
	}
	t := T{}
	Do("/jsonMap", &t)
	fmt.Printf("  %v\n", t)

	// Output: 200 [application/json]
	//   42
	// 200 [application/json]
	//   {Hello, World! 42}
}

func ExamplePoly_returningErrors() {
	// If a handler returns an error and the error != nil then
	// the response will be http.Error(w, err.Error(), http.StatusInternalServerError)

	var w *httptest.ResponseRecorder
	var req *http.Request
	b := &bytes.Buffer{}

	p := poly.Poly{
		FormMapper:  poly.DefaultFormMapper,
		PathMapper:  poly.DefaultPathMapper,
		QueryMapper: poly.DefaultQueryMapper,
	}
	mux := http.NewServeMux()

	Do := func(path string) {
		b.Reset()
		w = httptest.NewRecorder()
		w.Body = b
		req = httptest.NewRequest(http.MethodGet, path, nil)
		mux.ServeHTTP(w, req)
		fmt.Println(w.Code, "["+w.Header().Get("Content-Type")+"]", strings.TrimSpace(b.String()))
	}

	Error := func() error {
		return fmt.Errorf("internal error")
	}
	mux.Handle("/error", p.Handler(Error))

	// Multiple return values.
	multiple := 0
	Multiple := func() (string, error) {
		if multiple == 0 {
			multiple++
			return "No error!", nil
		}
		return "", fmt.Errorf("second call is error")
	}
	mux.Handle("/multiple", p.Handler(Multiple))

	Do("/error")
	Do("/multiple")
	Do("/multiple")

	// Output: 500 [text/plain; charset=utf-8] internal error
	// 200 [text/plain; charset=utf-8] No error!
	// 500 [text/plain; charset=utf-8] second call is error
}

func ExamplePoly_echo() {
	p := poly.Poly{}
	mux := http.NewServeMux()

	mux.Handle("/stdlib", http.HandlerFunc(examples.STDEcho))
	mux.Handle("/poly", p.Handler(examples.Echo))

	message := examples.EchoRequest{Message: "stdlib echo"}
	w := httptest.NewRecorder()
	w.Body = &bytes.Buffer{}
	body, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/stdlib", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(w, req)
	fmt.Println(w.Body.String())

	message = examples.EchoRequest{Message: "poly echo"}
	w = httptest.NewRecorder()
	w.Body = &bytes.Buffer{}
	body, err = json.Marshal(message)
	if err != nil {
		panic(err)
	}
	req = httptest.NewRequest(http.MethodPost, "/poly", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(w, req)
	fmt.Println(w.Body.String())

	// Output: stdlib echo
	// poly echo
}

func ExamplePoly_login() {
	p := poly.Poly{
		FormMapper: poly.DefaultFormMapper,
	}
	mux := http.NewServeMux()

	mux.Handle("/stdlib", http.HandlerFunc(examples.STDLogin))
	mux.Handle("/poly", p.Handler(examples.Login))

	form := url.Values{
		"username": []string{"nofeaturesonlybugs"},
		"password": []string{"hunter2"},
	}
	w := httptest.NewRecorder()
	w.Body = &bytes.Buffer{}
	req := httptest.NewRequest(http.MethodPost, "/stdlib", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(w, req)
	fmt.Println(w.Code)

	form = url.Values{
		"username": []string{"nofeaturesonlybugs"},
		"password": []string{"hunter2"},
	}
	w = httptest.NewRecorder()
	w.Body = &bytes.Buffer{}
	req = httptest.NewRequest(http.MethodPost, "/poly", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(w, req)
	fmt.Println(w.Code)

	// Output: 200
	// 200
}

func ExamplePoly_methods() {
	p := poly.Poly{}
	mux := http.NewServeMux()

	bill := examples.Greet{MyName: "Bill", MyAge: 16}
	ted := examples.Greet{MyName: "Ted", MyAge: 17}

	mux.Handle("/bill", p.Handler(bill.Hello))
	mux.Handle("/ted", p.Handler(ted.Hello))

	js := map[string]interface{}{
		"name": "Rufus",
	}
	blob, err := json.Marshal(js)
	if err != nil {
		panic(err)
	}

	w := httptest.NewRecorder()
	w.Body = &bytes.Buffer{}
	req := httptest.NewRequest(http.MethodPost, "/bill", bytes.NewBuffer(blob))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(w, req)
	fmt.Println(w.Code, w.Body.String())

	w = httptest.NewRecorder()
	w.Body = &bytes.Buffer{}
	req = httptest.NewRequest(http.MethodPost, "/ted", bytes.NewBuffer(blob))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(w, req)
	fmt.Println(w.Code, w.Body.String())

	// Output: 200 Hello Rufus!  I am Bill and I am 16 years old.
	// 200 Hello Rufus!  I am Ted and I am 17 years old.
}
