package examples

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// EchoRequest is the request type for STDEcho and Echo.
type EchoRequest struct {
	Message string `json:"message" form:"message" path:"message" query:"message"`
}

// STDEcho is an echo server based on stdlib net/http.
func STDEcho(w http.ResponseWriter, req *http.Request) {
	var dst EchoRequest
	dcd := json.NewDecoder(req.Body)
	if err := dcd.Decode(&dst); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, dst.Message)
}

// Echo is an echo handler.
func Echo(post EchoRequest) string {
	return post.Message
}
