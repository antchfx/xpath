package xpath

import "testing"

// XPath 1.0 section 3.2 requires an error when the number of arguments is wrong.
func TestFunctionArity(t *testing.T) {
	invalid := []string{
		`true(1)`, `false(1)`, `last(1)`, `position(1)`,
		`not(true(), false())`, `count(//book, //book)`, `sum(//book, 1)`,
		`ceiling(1, 2)`, `floor(1, 2)`, `round(1, 2)`,
		`normalize-space(" a ", "b")`, `substring("abc", 1, 1, 1)`,
		`starts-with("abc", "a", "ignored")`,
		`contains("abc", "b", "ignored")`,
		`ends-with("abc", "c", "ignored")`,
		`lower-case("ABC", "ignored")`, `reverse(//book, //book)`,
	}
	for _, expression := range invalid {
		t.Run(expression, func(t *testing.T) {
			if _, err := Compile(expression); err == nil {
				t.Fatalf("Compile(%q) accepted an invalid argument count", expression)
			}
		})
	}
	valid := []string{
		`true()`, `false()`, `last()`, `position()`, `not(true())`,
		`count(//book)`, `sum(//book/price)`, `ceiling(1)`, `floor(1)`, `round(1)`,
		`normalize-space()`, `normalize-space(" a ")`,
		`substring("abc", 1)`, `substring("abc", 1, 1)`,
		`starts-with("abc", "a")`, `contains("abc", "b")`, `ends-with("abc", "c")`,
		`lower-case("ABC")`, `reverse(//book)`, `concat("a", "b", "c")`,
	}
	for _, expression := range valid {
		t.Run(expression, func(t *testing.T) {
			if _, err := Compile(expression); err != nil {
				t.Fatalf("Compile(%q): %v", expression, err)
			}
		})
	}
}
