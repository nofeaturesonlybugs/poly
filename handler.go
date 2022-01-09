package poly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/nofeaturesonlybugs/call"
	"github.com/nofeaturesonlybugs/set"
)

var (
	// We need these reflect.Type values in the package; rather than repeatedly
	// creating these values we create once here and use as needed.
	argTypeRequest        = reflect.TypeOf((*http.Request)(nil))
	argTypeResponseWriter = reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()
)

// handler is the adapter that turns any function into an http.handler.
type handler struct {
	// The Poly instance used to create this handler.
	Poly Poly

	// Fn is the function this Handler wraps around.
	//
	// FnHasReturn indicates the first return value of calling Fn should
	// be written to the http.ResponseWriter.
	Fn          *call.Func
	FnHasReturn bool

	// PassThru is a slice of arguments we do not instantiate or unmarshal before calling
	// Fn but instead pass straight through.
	//
	// NB:  If Fn has http.ResponseWriter or *http.Request in its arguments they will be
	//      in PassThru.
	// TODO Want to add other per-request types here; for example Session; although Session
	//      might require separate field called Factory?
	PassThru []call.Arg

	// Each of the following slices contains indexes into Fn's argument list
	// for arguments that can be the target of the specified data.
	//
	// For example Form=[]int{1, 3} means Fn arguments with indexes 1, 3 can populated
	// from incoming form data.
	Form  []int
	JSON  []int
	Path  []pathParams
	Query []int
}

// newHandler creates a new Handler.
func newHandler(poly Poly, fn *call.Func) handler {
	rv := handler{
		Poly:     poly,
		Fn:       fn,
		PassThru: fn.PruneIn(argTypeRequest, argTypeResponseWriter),
		Form:     nil,
		JSON:     nil,
		Path:     nil,
		Query:    nil,
	}
	var mapped *set.Mapping
	//
	for k, T := range fn.InTypes {
		switch true {
		case T == argTypeRequest:
			continue
		case T.Kind() == reflect.Interface: // Also covers T == ArgTypeResponseWriter
			continue
		}
		//
		if poly.FormMapper != nil {
			mapped = poly.FormMapper.Map(T)
			if len(mapped.Keys) > 0 {
				rv.Form = append(rv.Form, k)
			}
		}
		//
		if poly.PathMapper != nil {
			mapped = poly.PathMapper.Map(T)
			if len(mapped.Keys) > 0 {
				rv.Path = append(rv.Path, pathParams{N: k, Keys: mapped.Keys})
			}
			// TODO If len(rv.Path)>0 and rv.Poly.PathParam is nil
			// then params can't be retrieved -- issue warning or error maybe.
		}
		//
		if poly.QueryMapper != nil {
			mapped = poly.QueryMapper.Map(T)
			if len(mapped.Keys) > 0 {
				rv.Query = append(rv.Query, k)
			}
		}
		//
		// If the argument is a struct or ptr-to-struct add the index to the JSONIndeces.
		// NB: Split onto multiple lines for code coverage reasons.
		// TODO Maybe check that it has public fields?
		if T.Kind() == reflect.Struct {
			rv.JSON = append(rv.JSON, k)
		} else if T.Kind() == reflect.Ptr && T.Elem().Kind() == reflect.Struct {
			rv.JSON = append(rv.JSON, k)
		}
	}
	//
	if fn.NumOut >= 1 {
		switch fn.OutTypes[0].Kind() {
		case reflect.Bool,
			reflect.Float32, reflect.Float64,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.String,
			reflect.Ptr,
			reflect.Array, reflect.Slice,
			reflect.Map,
			reflect.Struct:
			rv.FnHasReturn = true

		default:
			// TODO Warning that argument is not a type that can send to http.ResponseWriter
			// reflect.UnsafePointer
			// reflect.Uintptr
			// reflect.Invalid
			// reflect.Complex64
			// reflect.Complex128
			// reflect.Chan
			// reflect.Func
			// reflect.Interface
		}
	}
	//
	return rv
}

// ServeHTTP implements http.Handler.
func (h handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var err error
	args := h.Fn.Args()
	//
	// PassThru args are set directly.
	for _, passThru := range h.PassThru {
		switch passThru.T {
		case argTypeRequest:
			args.Values[passThru.N] = reflect.ValueOf(req)
		case argTypeResponseWriter:
			args.Values[passThru.N] = reflect.ValueOf(w)
		}
	}
	//
	// Unmarshal path parameters.
	if h.Poly.PathParamer != nil {
		for _, param := range h.Path {
			b := h.Poly.PathMapper.Bind(args.Pointers[param.N])
			for _, param := range param.Keys {
				b.Set(param, h.Poly.PathParam(req, param))
			}
			// TODO Error reporting?
		}
	}
	//
	// Unmarshal query parameters.
	if h.Query != nil {
		params := req.URL.Query()
		for _, n := range h.Query {
			b := h.Poly.QueryMapper.Bind(args.Pointers[n])
			for param := range params {
				b.Set(param, params[param])
			}
			// TODO Error reporting?
		}
	}
	//
	// Unmarshal body
	contentType := req.Header.Get("Content-Type")
	tryForm := h.Form != nil && contentType == "application/x-www-form-urlencoded"
	tryJSON := h.JSON != nil && contentType == "application/json"
	if tryJSON {
		buf := &bytes.Buffer{}
		if _, err = buf.ReadFrom(req.Body); err != nil {
			http.Error(w, "reading body", http.StatusBadRequest) // TODO Better
			return
		}
		for _, n := range h.JSON {
			if err = json.Unmarshal(buf.Bytes(), args.Pointers[n]); err != nil {
				http.Error(w, "decoding json", http.StatusBadRequest) // TODO Better
				return
			}
		}
	} else if tryForm {
		if err = req.ParseForm(); err != nil {
			http.Error(w, "parse form", http.StatusBadRequest) // TODO
			return
		}
		for _, n := range h.Form {
			b := h.Poly.FormMapper.Bind(args.Pointers[n])
			for name, value := range req.PostForm {
				b.Set(name, value)
			}
			// TODO Error reporting?
		}
	}

	//
	result := h.Fn.Call(args) // TODO Error, Results?
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
	} else if h.FnHasReturn {
		switch value := result.Values[0].(type) {
		case string:
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			fmt.Fprint(w, value)
			// if _, err = fmt.Fprint(w, value); err != nil {
			// 	// TODO Potential logging.
			// 	// TODO Potential custom handler provided by h.Poly.
			// }

		default:
			var blob []byte
			if blob, err = json.Marshal(value); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, string(blob))
			// if _, err = fmt.Fprint(w, string(blob)); err != nil {
			// 	// TODO Potential logging.
			// 	// TODO Potential custom handler provided by h.Poly.
			// }
			// TODO Old below
			// for n, remaining := 0, len(blob); remaining > 0; blob, remaining = blob[n:], remaining-n {
			// 	if n, err = w.Write(blob); err != nil {
			// 		// TODO http.Error()?
			// 		return
			// 	}
			// }
		}
	}
}
