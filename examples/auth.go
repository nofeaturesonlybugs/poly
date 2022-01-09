package examples

import "net/http"

const (
	AuthUser     = "nofeaturesonlybugs"
	AuthPassword = "hunter2"
)

// AuthLoginRequest is a request from user to login.
type AuthLoginRequest struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

// STDLogin is a standard login handler.
func STDLogin(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	username := req.PostForm.Get("username")
	password := req.PostForm.Get("password")
	if username == "nofeaturesonlybugs" && password == "hunter2" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusForbidden)
}

// Login is a login handler.
func Login(w http.ResponseWriter, post AuthLoginRequest) {
	if post.Username == "nofeaturesonlybugs" && post.Password == "hunter2" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusForbidden)
}
