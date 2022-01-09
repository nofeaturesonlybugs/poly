package examples

import "fmt"

// Greet tests that methods wrapped through poly.Handler work as expected.
type Greet struct {
	MyName string
	MyAge  int
}

// Hello is the handler for saying hello to Greet.
func (g Greet) Hello(post struct {
	Name string `json:"name"`
}) string {
	return fmt.Sprintf("Hello %v!  I am %v and I am %v years old.", post.Name, g.MyName, g.MyAge)
}
