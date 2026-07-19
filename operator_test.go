package xpath

import "testing"

func createComparisonDoc() *TNode {
	doc := createNode("", RootNode)
	root := doc.createChildNode("Root", ElementNode)
	nodes := []struct {
		name  string
		value string
	}{
		{name: "Low", value: "2"},
		{name: "High", value: "10"},
		{name: "Padded", value: " 14 "},
		{name: "Invalid", value: "abc"},
	}
	for _, node := range nodes {
		child := root.createChildNode(node.name, ElementNode)
		child.createChildNode(node.value, TextNode)
	}
	return doc
}

func TestXPathComparisons(t *testing.T) {
	doc := createComparisonDoc()
	tests := []struct {
		name string
		expr string
		want bool
	}{
		{name: "string number greater", expr: `'14' > 0`, want: true},
		{name: "number string less", expr: `0 < '14'`, want: true},
		{name: "string number less false", expr: `'14' < 0`, want: false},
		{name: "number string greater false", expr: `0 > '14'`, want: false},
		{name: "string number equal", expr: `'14' = 14`, want: true},
		{name: "number string equal", expr: `14 = '14'`, want: true},
		{name: "string number not equal", expr: `'14' != 15`, want: true},
		{name: "number string not equal", expr: `15 != '14'`, want: true},

		{name: "invalid string equals number", expr: `'abc' = 1`, want: false},
		{name: "number equals invalid string", expr: `1 = 'abc'`, want: false},
		{name: "invalid string differs from number", expr: `'abc' != 1`, want: true},
		{name: "number differs from invalid string", expr: `1 != 'abc'`, want: true},
		{name: "invalid string relational", expr: `'abc' < 1`, want: false},
		{name: "number invalid string relational", expr: `1 < 'abc'`, want: false},
		{name: "number trims string whitespace", expr: `number(' 14 ') = 14`, want: true},
		{name: "number invalid string is NaN", expr: `number('abc') != 1`, want: true},
		{name: "number converts true", expr: `number(true()) = 1`, want: true},
		{name: "number converts false", expr: `number(false()) = 0`, want: true},

		{name: "string relational less", expr: `'2' < '10'`, want: true},
		{name: "string relational greater", expr: `'10' > '2'`, want: true},
		{name: "string relational less or equal", expr: `'2' <= '10'`, want: true},
		{name: "string relational greater or equal", expr: `'10' >= '2'`, want: true},
		{name: "string equality stays lexical", expr: `'2' = '02'`, want: false},
		{name: "string inequality stays lexical", expr: `'2' != '02'`, want: true},

		{name: "padded node equals number", expr: `//Padded = 14`, want: true},
		{name: "number equals padded node", expr: `14 = //Padded`, want: true},
		{name: "invalid node equals number", expr: `//Invalid = 1`, want: false},
		{name: "number equals invalid node", expr: `1 = //Invalid`, want: false},
		{name: "invalid node differs from number", expr: `//Invalid != 1`, want: true},
		{name: "number differs from invalid node", expr: `1 != //Invalid`, want: true},

		{name: "node set less than string", expr: `//Low < '10'`, want: true},
		{name: "string greater than node set", expr: `'10' > //Low`, want: true},
		{name: "node set greater than string", expr: `//High > '2'`, want: true},
		{name: "string less than node set", expr: `'2' < //High`, want: true},
		{name: "node set less or equal string", expr: `//Low <= '2'`, want: true},
		{name: "string greater or equal node set", expr: `'10' >= //High`, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MustCompile(tt.expr).Evaluate(createNavigator(doc))
			if got != tt.want {
				t.Fatalf("%s: got %v, want %v", tt.expr, got, tt.want)
			}
		})
	}
}
