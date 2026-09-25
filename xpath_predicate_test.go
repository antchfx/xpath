package xpath

import (
	"testing"
)

func TestPositionalCondition(t *testing.T) {
	tests := []struct {
		name string
		node node
		want bool
	}{
		{"numeric literal", newOperandNode(float64(1)), true},
		{"string literal", newOperandNode("text"), false},
		{"position()", newFunctionNode("position", "", nil), true},
		{"last()", newFunctionNode("last", "", nil), true},
		{"count()", newFunctionNode("count", "", nil), true},
		{"sum()", newFunctionNode("sum", "", nil), true},
		{"string-length()", newFunctionNode("string-length", "", nil), true},
		{"number()", newFunctionNode("number", "", nil), true},
		{"floor()", newFunctionNode("floor", "", nil), true},
		{"ceiling()", newFunctionNode("ceiling", "", nil), true},
		{"round()", newFunctionNode("round", "", nil), true},
		{"name()", newFunctionNode("name", "", nil), false},
		{"normalize-space()", newFunctionNode("normalize-space", "", nil), false},
		{"string(count())", newFunctionNode("string", "", []node{newFunctionNode("count", "", nil)}), true},
		{"+", newOperatorNode("+", newOperandNode("a"), newOperandNode("b")), true},
		{"-", newOperatorNode("-", newOperandNode("a"), newOperandNode("b")), true},
		{"*", newOperatorNode("*", newOperandNode("a"), newOperandNode("b")), true},
		{"div", newOperatorNode("div", newOperandNode("a"), newOperandNode("b")), true},
		{"mod", newOperatorNode("mod", newOperandNode("a"), newOperandNode("b")), true},
		{"@href", newAxisNode("attribute", AttributeNode, "href", "", "", nil), false},
		{"= non-positional", newOperatorNode("=", newOperandNode("a"), newOperandNode("b")), false},
		{"!= non-positional", newOperatorNode("!=", newOperandNode("a"), newOperandNode("b")), false},
		{"position()=1", newOperatorNode("=", newFunctionNode("position", "", nil), newOperandNode(float64(1))), true},
		{"last()-1", newOperatorNode("-", newFunctionNode("last", "", nil), newOperandNode(float64(1))), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := positionalCondition(tt.node); got != tt.want {
				t.Errorf("positionalCondition(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestLogicals(t *testing.T) {
	test_xpath_elements(t, book_example, `//book[1 + 1]`, 9)
	test_xpath_elements(t, book_example, `//book[1 * 2]`, 9)
	test_xpath_elements(t, book_example, `//book[5 div 2]`, 9) // equal to `//book[2]`
	test_xpath_elements(t, book_example, `//book[3 div 2]`, 3)
	test_xpath_elements(t, book_example, `//book[3 - 2]`, 3)
	test_xpath_elements(t, book_example, `//book[price > 35]`, 15, 25)
	test_xpath_elements(t, book_example, `//book[price >= 30]`, 3, 15, 25)
	test_xpath_elements(t, book_example, `//book[price < 30]`, 9)
	test_xpath_elements(t, book_example, `//book[price <= 30]`, 3, 9)
	test_xpath_elements(t, book_example, `//book[count(author) > 1]`, 15)
	test_xpath_elements(t, book_example, `//book[position() mod 2 = 0]`, 9, 25)
}

func TestPositions(t *testing.T) {
	test_xpath_elements(t, employee_example, `/empinfo/employee[2]`, 8)
	test_xpath_elements(t, employee_example, `//employee[position() = 2]`, 8)
	test_xpath_elements(t, employee_example, `/empinfo/employee[2]/name`, 9)
	test_xpath_elements(t, employee_example, `//employee[position() > 1]`, 8, 13)
	test_xpath_elements(t, employee_example, `//employee[position() <= 2]`, 3, 8)
	test_xpath_elements(t, employee_example, `//employee[last()]`, 13)
	test_xpath_elements(t, employee_example, `//employee[position() = last()]`, 13)
	test_xpath_elements(t, book_example, `//book[@category = "web"][2]`, 25)
	test_xpath_elements(t, book_example, `//book[@category = "web"][last()]`, 25)
	test_xpath_elements(t, book_example, `(//book[@category = "web"])[2]`, 25)
	test_xpath_elements(t, book_example, `(//book[@category = "web"])[last()]`, 25)
	test_xpath_elements(t, employee_example, `(//employee)[last()]`, 13)
	test_xpath_elements(t, book_example, `(//author)[last()]`, 27)
}

func TestPredicates(t *testing.T) {
	test_xpath_elements(t, employee_example, `//employee[name]`, 3, 8, 13)
	test_xpath_elements(t, employee_example, `/empinfo/employee[@id]`, 3, 8, 13)
	test_xpath_elements(t, book_example, `//book[@category = "web"]`, 15, 25)
	test_xpath_elements(t, book_example, `//book[author = "J K. Rowling"]`, 9)
	test_xpath_elements(t, book_example, `//book[./author/text() = "J K. Rowling"]`, 9)
	test_xpath_elements(t, book_example, `//book[year = 2005]`, 3, 9)
	test_xpath_elements(t, book_example, `//year[text() = 2005]`, 6, 12)
	test_xpath_elements(t, employee_example, `/empinfo/employee[1][@id=1]`, 3)
	test_xpath_elements(t, employee_example, `/empinfo/employee[@id][2]`, 8)
}

func TestOperators(t *testing.T) {
	test_xpath_elements(t, employee_example, `//designation[@discipline and @experience]`, 5, 10)
	test_xpath_elements(t, employee_example, `//designation[@discipline or @experience]`, 5, 10, 15)
	test_xpath_elements(t, employee_example, `//designation[@discipline | @experience]`, 5, 10, 15)
	test_xpath_elements(t, employee_example, `/empinfo/employee[@id != "2" ]`, 3, 13)
	test_xpath_elements(t, employee_example, `/empinfo/employee[@id and @id = "2"]`, 8)
	test_xpath_elements(t, employee_example, `/empinfo/employee[@id = "1" or @id = "2"]`, 3, 8)
}

func TestNestedPredicates(t *testing.T) {
	test_xpath_elements(t, employee_example, `//employee[./name[@from]]`, 8)
	test_xpath_elements(t, employee_example, `//employee[.//name[@from = "CA"]]`, 8)
}

func TestChainedPredicateOnGroupedExpression(t *testing.T) {
	test_xpath_count(t, employee_example, `(//employee)[1][false()]`, 0)
	test_xpath_count(t, employee_example, `(//employee)[2][false()]`, 0)
	test_xpath_count(t, employee_example, `(//employee)[1][@id = "2"]`, 0)
	test_xpath_elements(t, employee_example, `(//employee)[1][true()]`, 3)
	test_xpath_elements(t, employee_example, `(//employee)[1][@id = "1"]`, 3)
	test_xpath_elements(t, employee_example, `(//employee)[2][@id = "2"]`, 8)
}
