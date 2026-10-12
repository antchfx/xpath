package xpath

import (
	"fmt"
	"testing"
)

func TestBooleanEqualityConversion(t *testing.T) {
	for _, tc := range []struct {
		expression string
		truth      bool
	}{
		{"2", true}, {"-1", true}, {"0", false},
		{"number('invalid')", false}, {"'text'", true}, {"'0'", true},
		{"''", false}, {"/", true}, {"missing", false},
	} {
		for _, operator := range []string{"=", "!="} {
			want := tc.truth
			if operator == "!=" {
				want = !want
			}
			for _, expression := range []string{
				fmt.Sprintf("%s %s true()", tc.expression, operator),
				fmt.Sprintf("true() %s %s", operator, tc.expression),
			} {
				t.Run(expression, func(t *testing.T) {
					compiled, err := Compile(expression)
					if err != nil {
						t.Fatal(err)
					}
					if got := compiled.Evaluate(createNavigator(empty_example)); got != want {
						t.Errorf("Evaluate = %v, want %v", got, want)
					}
				})
			}
		}
	}
}
