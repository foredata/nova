package pretty_test

import (
	"fmt"
	"testing"

	"github.com/foredata/nova/pkg/pretty"
)

func TestJson(t *testing.T) {
	d1 := `{"name":{"first":"Tom","last":"Anderson"},"age":37,"children":["Sara","Alex","Jack"],"fav.movie":"Deer Hunter","friends":[{"first":"Janet","last":"Murphy","age":44}]}`
	fmt.Print(pretty.Indent(d1))
	fmt.Print(pretty.Flat(d1), "\n")

	d2 := `{"name":  {"first":"Tom","last":"Anderson"},  "age":37,
	"children": ["Sara","Alex","Jack"],
	"fav.movie": "Deer Hunter", "friends": [
		{"first": "Janet", "last": "Murphy", "age": 44}
	  ]}`
	fmt.Print(pretty.Indent(d2))
	fmt.Print(pretty.Flat(d2), "\n")

	type foo struct {
		Name struct {
			First string
			Last  string
		}
		Age      int
		Children []string
		Items    []int
	}

	f := &foo{}
	f.Name.First = "Tom"
	f.Name.Last = "Anderson"
	f.Age = 37
	f.Children = []string{"Sara", "Alex", "Jack"}
	f.Items = []int{1, 2, 3}
	fmt.Print(pretty.Indent(f))
	fmt.Print(pretty.Flat(f), "\n")
}
