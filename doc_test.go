package xpath

import (
	"fmt"
)

// XPath package example.
func Example() {
	expr, err := Compile("count(//book)")
	if err != nil {
		panic(err)
	}
	var root NodeNavigator
	// using Evaluate() method
	val := expr.Evaluate(root) // it returns float64 type
	fmt.Println(val.(float64))

	// using Evaluate() method
	expr = MustCompile("//book")
	val = expr.Evaluate(root) // it returns NodeIterator type.
	iter := val.(*NodeIterator)
	for iter.MoveNext() {
		fmt.Println(iter.Current().Value())
	}

	// using Select() method
	iter = expr.Select(root) // it always returns NodeIterator object.
	for iter.MoveNext() {
		fmt.Println(iter.Current().Value())
	}
}
