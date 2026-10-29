package xpath

import "testing"

func TestEndsWithNodeSetSuffix(t *testing.T) {
	tests := []struct {
		expression string
		want       bool
	}{
		{`ends-with("Everyday Italian", //book[1]/title)`, true},
		{`ends-with("prefix-2005", //book[1]/year)`, true},
		{`ends-with("prefix-2004", //book[1]/year)`, false},
		{`ends-with(//book[1]/title, //book[1]/title)`, true},
		{`ends-with("anything", //missing)`, true},
		{`ends-with(//missing, //missing)`, true},
	}
	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Fatalf("evaluation panicked: %v", err)
				}
			}()
			test_xpath_eval(t, book_example, tt.expression, tt.want)
		})
	}
}
