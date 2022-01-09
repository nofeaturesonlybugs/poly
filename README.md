`poly` turns any function into `http.Handler`

## Why Poly
Poly is my attempt at addressing some of the common grievances when working with standard library's `net/http`, which is all the fluff and boilerplate required when unmarshaling data from `*http.Request` and writing to `http.ResponseWriter`.

There's certainly no shortage of existing packages that are supposed to make these tasks easier so what's so special about Poly?

With Poly you can write HTTP handlers that represent their concerns.  Your handler arguments and their return value(s) determine how Poly behaves.  The following is easier to write, easier to digest, and easier to test:
```go
type CreateUserRequest struct {
    Username string `form:"username"`
    Password string `form:"password"`
}
func CreateUser(r CreateUserRequest) (User, error) {
}
```
As far as I'm aware **none** of the existing Go packages and libraries allow you to write your handlers so succinctly.  Your existing choices are:  
1.  use `net/http` and deal with the boilerplate of unmarshaling `*http.Request` or
2.  learning a whole new<sup> and stdlib-incompatible</sup> API usually with new `Request` or `RequestContext` types.

Poly **is** `net/http` compatible.  When you apply Poly to one of your handlers it returns `http.Handler`.  So even though your handlers themselves may not be standard `http.Handler` signatures you can still use them anywhere else `http.Handler` is expected.  You can use all of the existing muxes, routers, and middlewares that expect `http.Handler` *while* writing handlers that only represent their concerns as explained above.

Poly retains your access to `http.ResponseWriter` and `*http.Request` if you still need them.  Any of the following handler signatures will be given the aforementioned arguments:
```go
func NeedsWriter(w http.ResponseWriter, ...) {} // Handler called with ResponseWriter
func NeedsRequest(req *http.Request, ...) {} // Handler called with Request
func NeedsBoth(w http.ResponseWriter, req *http.Request, ...) {} // Handler called with both
```

## When Poly  
I'm a bit biased so I think Poly is pretty awesome -- but it's not suitable for all purposes.

I. Consider Poly for low volume sites or low volume endpoints in a larger site.  
  > Poly needs `reflect` to work its magic and `reflect` isn't free.  There's overhead involved in creating handler arguments, populating them, and invoking your handler via `reflect`.  I would not recommend using Poly for high volume or high traffic sites or endpoints.  

II. Consider Poly for simple or typical handler behavior.  
  > Poly is intended to handle simple or typical requests.  Poly does not aim to replace the need for `http.Handler` altogether.  If you can represent and implement your handler ergonomically with Poly then by all means do so.  But if you need complicated behavior out of either `http.ResponseWriter` or `*http.Request` then consider using a standard `http.Handler`.  If you become bogged down with the unmarshaling behavior for a specific request with Poly then consider making it a standard `http.Handler.`

III. Prototyping or More Rapid Production  
  > Since Poly allows your handlers to take on the most succinct signatures possible you may be able to prototype a project or application more quickly than with `net/http` or other Go libraries.  As your project or site volume grows you can continue to use Poly or change high traffic or high volume endpoints to `http.Handler` while continuing to use Poly for low (maybe even medium) volume endpoints.


## How Poly

### Code Glossary  
Within the code examples:
+ `p` is taken to be an instance of `poly.Poly`
+ `mux` is taken to be an instance of `http.ServeMux`
+ `T` is taken to be a `type T struct{...}` of sufficient complexity.

### Create A Poly Instance  
```go
p := poly.Poly{}
```

### `Handler()` Your Functions  
If you pass a function to `Poly.Handler()` then you get an `http.Handler`.
```go
func IndexHandler() {
    fmt.Println("IndexHandler was called!")
}

mux.Handle("/", p.Handler(IndexHandler))
```

You can only pass functions to `Poly.Handler()` -- any other type results in a panic.
```go
mux.Handle("/", p.Handler("Hello, World!")) // panics
```

## Getting Data  
Handlers created by Poly can automatically unmarshal `*http.Request` data into your handler arguments.
```go
type CreateUserRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}
func CreateUser(req CreateUserRequest) (User, error) {
    return UserStore.CreateUser(req)
}
mux.Handle("/create-user", p.Handler(CreateUser))
```
In the above example the call to `p.Handler` inspects the arguments to `CreateUser` and determines which arguments are the targets for any of the following:
1. JSON unmarshaling  
  i. Currently requires request to have `application/json` content type.
2. Form unmarshaling  
  i. Currently requires request to have `application/x-www-form-urlencoded` content type.
3. Path-Parameter unmarshaling  
  i. Path params are those in the URI path before the `?` in `/a/b/c?query=string`  
  ii. Path params are usually provided by your chosen mux or router library.
4. Query string unmarshaling  
  i. The query string occurs after the `?` in `/a/b/c?this=is&the=query&string`

A *zero-value* Poly can only unmarshal requests containing JSON.

You must set the `FormMapper` and `QueryMapper` fields in Poly to enable form and query string unmarshaling respectively:
```go
p := poly.Poly{
    FormMapper:  poly.DefaultFormMapper,
    QueryMapper: poly.DefaultQueryMapper,
}
// DefaultFormMapper targets struct-tag: form
type OrderPizzaRequest struct {
    Size string `form:"size"`
    Toppings []string `form:"toppings"`
}
// DefaultQueryMapper targets struct-tag: query
type CancelOrderRequest struct {
    OrderId int `query:"order"`
}
```

To enable unmarshaling of path parameters you must set both the `PathMapper` and `PathParamer` fields in Poly:
```go
// A stub for some fancy router or mux library supporting path parameters.
rtr := routerlib.Router{}
p := poly.Poly{
    // PathMapper determines which handler arguments are targets for path parameters...
    PathMapper:  poly.DefaultPathMapper,
    // ...while PathParamer is the provider of argument values.
    PathParamer: rtr.GetParam,
}
type ViewStudentRequest struct {
    StudentId int `path:"student"`
}
func ViewStudent(req ViewStudentRequest) Student { /*...*/ }
rtr.Handle( "/view-student/:student", p.Handler(ViewStudent))
```

## Returning Data  
Poly handlers make it easy to write responses to the client by simply returning it from your function.

### `text/plain`  
If your handler returns `string` then it is written to the client as `text/plain; charset=utf-8`.
```go
func LoremHandler() string {
    return "Lorem ipsum..."
}
mux.Handle("/", p.Handler(LoremHandler))
```

### `application/json`  
Any of the following return types are encoded as `JSON`:
+ bool
+ float32 and float64
+ int and uint including all bit variants
+ pointers, arrays, slices, maps, and structs
```go
func CreateThingHandler(...) T {
    return T{...}
}
mux.Handle("/create-thing", p.Handler(CreateThingHandler))
```

### `error`  
If your handler returns one or more values and one of them is type `error` then it is checked before writing other output to the client.
1. `err != nil` means `http.Error(w, err.Error(), http.StatusInternalServerError)`
2. If your handler returns multiple error values then the last error is the one that is checked.
    + Write *sane* handlers that return zero or one error value.
```go
func CreateThingHandler(...) (*T, error) {
    if ... {
        return nil, fmt.Errorf("oops") // http.Error(...)
    }
    return &T{...}, nil // JSON encoded
}
mux.Handle("/create-thing", p.Handler(CreateThingHandler))
```

### Need More Control?  
If you need more control over the response then just include `http.ResponseWriter` in the argument list and use it like you normally would.
```go
func CreateThingHandler(w http.ResponseWriter, ...) {
    w.Header().Set("X-FOO", "value")
    w.Write([]byte("data")) 
}
mux.Handle("create-thing", p.Handler(CreateThingHandler))
```

# Performance Tips  
Try to limit the number of arguments to your handlers.  Poly has to create your handler arguments before invoking your handler.  Less arguments means less allocations by Poly.

Try to limit the number of unmarshal sources.  In general JSON is faster than paths and query strings which are faster than forms.
```go
type ViewStudentsRequest struct {
    // Fastest....
    // Single source: JSON request body
    // Requires on unmarshal pass in Poly.
    SchoolId int `json:"school"`
    Sort []string `json:"sort"`
    SelectFields []string `json:"fields"`
}
type ViewStudentsRequest struct {
    // Single source: query string
    // Requires one unmarshal pass in Poly.
    SchoolId int `query:"school"`
    Sort []string `query:"sort"`
    SelectFields []string `query:"fields"`
}
type ViewStudentsRequest struct {
    // Slowest...
    // Two sources: path, form
    // Requires two unmarshal passes in Poly.
    SchoolId int `path:"school"`
    Sort []string `form:"sort"`
    SelectFields []string `form:"fields"`
}
```

# Convenience Notes  

## `http.Handler`, `http.HandlerFunc`, & `func(w, req)`  
`Poly.Handler()` returns its argument if the argument is already `http.Handler` or `http.HandlerFunc`.
```go
mux.Handle("/404", p.Handler(http.NotFoundHandler()))
mux.Handle("/redir", p.Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
    http.Redirect(w, req, "/", http.StatusPermanentRedirect)
})))
```

If the argument is `func(http.ResponseWriter, *http.Request)` then it is cast to `http.HandlerFunc` and returned.  Note that this means you can replace `http.HandlerFunc(fn)` with `p.Handler(fn)`.
```go
mux.Handle("/redir2", p.Handler(func(w http.ResponseWriter, req *http.Request) {
    http.Redirect(w, req, "/", http.StatusPermanentRedirect)
}))
```

In other words `Poly.Handler()` only adds overhead if the argument is incompatible with `http.Handler` and is conveniently shorter than `http.HandlerFunc()`.

# Why Does `Poly.Handler()` Panic?  
`Poly.Handler()` panics instead of returning errors for two reasons.

First and foremost a panic from `Poly.Handler()` means you've passed it a type that is not a function.  This is a programmer mistake and likely to be caught during the development cycle as you're wiring your routes.  Such a panic is unlikely to make it into production and -- in my opinion -- equivalent or on par with a panic caused by a `nil receiver`.

Secondly `Poly.Handler()` returns `http.Handler` *and only* `http.Handler` so that it can be passed seemlessly into other middleware.
```go
// Other middlewares
logging := func(next http.Handler) http.Handler {...}
sessions := func(next http.Handler)  http.Handler {...}

mux.Handle("/one", logging(sessions(p.Handler(OneHandler))))
mux.Handle("/two", logging(sessions(p.Handler(TwoHandler))))
```

If `Poly.Handler()` returned `(http.Handler, error)` then the above needs to be rewritten to check errors.
```go
// Other middlewares
logging := func(next http.Handler) http.Handler {...}
sessions := func(next http.Handler)  http.Handler {...}

if h, err = p.Handler(OneHandler); err != nil {
    return err
}
mux.Handle("/one", logging(sessions(h)))
if h, err = p.Handler(TwoHandler); err != nil {
    return err
}
mux.Handle("/two", logging(sessions(h)))
```

Nobody would want to use this library if you had to check an error for every wrapped route.
