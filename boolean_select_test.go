package xpath

import "testing"

func TestBooleanExpressionsDoNotSelectNodes(t *testing.T) {
	doc := createComparisonDoc()
	nav := createNavigator(doc)
	for _, tc := range []struct {
		expr string
		want bool
	}{
		{`//Low or //High`, true},
		{`//Low and //High`, true},
		{`1 = 1 or //Low`, true},
		{`//Low and 1 = 1`, true},
		{`//Missing or //Absent`, false},
		{`//Low and //Absent`, false},
	} {
		t.Run(tc.expr, func(t *testing.T) {
			compiled := MustCompile(tc.expr)
			iter := compiled.Select(nav)
			for i := 0; i < 2; i++ {
				if iter.MoveNext() {
					t.Fatalf("boolean expression %q selected a node on attempt %d", tc.expr, i+1)
				}
			}
			if got := compiled.Evaluate(nav); got != tc.want {
				t.Fatalf("boolean expression %q evaluated to %v, want %v", tc.expr, got, tc.want)
			}
		})
	}
}
