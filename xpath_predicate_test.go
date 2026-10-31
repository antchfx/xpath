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
	test_xpath_elements(t, book_example, `//book[4 div 2]`, 9)
	test_xpath_count(t, book_example, `//book[5 div 2]`, 0) // 2.5 is not a position
	test_xpath_count(t, book_example, `//book[3 div 2]`, 0)
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

func TestFractionalPositions(t *testing.T) {
	// A number predicate is true only when it equals the context position,
	// so a number that is not an integer selects nothing.
	test_xpath_count(t, employee_example, `//employee[1.5]`, 0)
	test_xpath_count(t, employee_example, `//employee[0.5]`, 0)
	test_xpath_count(t, employee_example, `/empinfo/employee[2.5]`, 0)
	test_xpath_count(t, employee_example, `(//employee)[1.5]`, 0)
	test_xpath_count(t, employee_example, `//employee[@id][1.5]`, 0)
	test_xpath_count(t, employee_example, `//employee[last() div 2]`, 0)
	test_xpath_count(t, employee_example, `//employee[number("a")]`, 0)
	test_xpath_elements(t, employee_example, `//employee[2.0]`, 8)
	test_xpath_elements(t, book_example, `//book[last() div 2]`, 9)
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

// https://github.com/antchfx/xpath/pull/162
func TestPositionAfterEarlierPredicate(t *testing.T) {
	// position() counts the nodes the earlier predicate kept, as a numeric
	// predicate does.
	test_xpath_elements(t, employee_example, `/empinfo/employee[@id != "2"][1]`, 3)
	test_xpath_elements(t, employee_example, `/empinfo/employee[@id != "2"][2]`, 13)
	test_xpath_elements(t, employee_example, `/empinfo/employee[@id != "2"][position()=1]`, 3)
	test_xpath_elements(t, employee_example, `/empinfo/employee[@id != "2"][position()=2]`, 13)
	test_xpath_elements(t, employee_example, `/empinfo/employee[@id != "1"][position()=1]`, 8)
	test_xpath_elements(t, employee_example, `/empinfo/employee[@id != "2"][position()<2]`, 3)
	test_xpath_elements(t, employee_example, `//employee[@id != "2"][position()=2]`, 13)
	// A parenthesized node-set is numbered the same way.
	test_xpath_elements(t, employee_example, `(//employee[@id != "2"])[position()=2]`, 13)
	// A nested step keeps its own position.
	test_xpath_elements(t, employee_example, `/empinfo/employee[@id != "2"][name[position()=1]]`, 3, 13)
}

func TestPositionFunctionFollowsAxisOrder(t *testing.T) {
	// position() is the position along the axis, so [position()=n] selects
	// what [n] does.
	doc := createElement(0, "", createElement(1, "r",
		createElement(2, "a"), createElement(3, "b"), createElement(4, "a"),
		createElement(5, "c"), createElement(6, "a", createElement(7, "b"))))
	test_xpath_elements(t, doc, `/r/b/following-sibling::a[position()=1]`, 4)
	test_xpath_elements(t, doc, `/r/b/following-sibling::a[position()=2]`, 6)
	test_xpath_elements(t, doc, `/r/c/preceding-sibling::a[position()=1]`, 4)
	test_xpath_elements(t, doc, `/r/c/preceding-sibling::a[position()=2]`, 2)
	test_xpath_elements(t, doc, `/descendant::*[position()=3]`, 3)
	test_xpath_elements(t, doc, `/r/a/b/ancestor::*[position()=1]`, 6)
	test_xpath_elements(t, doc, `/r/a/b/ancestor::*[position()=2]`, 1)
}

func TestPositionAfterAnotherPredicate(t *testing.T) {
	// In a[@x][1], the position counts the a elements with an x attribute
	// for each context node, so every p contributes its first one.
	doc := createElement(0, "", createElement(1, "r",
		createElement(2, "p",
			createElement(3, "a"),
			createElementAttr(4, "a", map[string]string{"x": "1"}),
			createElementAttr(5, "a", map[string]string{"x": "2"})),
		createElement(6, "p",
			createElementAttr(7, "a", map[string]string{"x": "3"}),
			createElement(8, "a"),
			createElementAttr(9, "a", map[string]string{"x": "4"}))))
	test_xpath_elements(t, doc, `//a[@x][1]`, 4, 7)
	test_xpath_elements(t, doc, `//a[@x][2]`, 5, 9)
	test_xpath_elements(t, doc, `//a[@x][position()=1]`, 4, 7)
	test_xpath_elements(t, doc, `/r/p/a[@x][1]`, 4, 7)
	test_xpath_elements(t, doc, `//a[not(@x)][1]`, 3, 8)
	test_xpath_elements(t, doc, `//a[@x][@x][1]`, 4, 7)
	test_xpath_elements(t, doc, `//a[@x][1][@x]`, 4, 7)
	test_xpath_elements(t, doc, `//*[@x][1]`, 4, 7)
	test_xpath_count(t, doc, `//a[@x][1]/@x`, 2)
	test_xpath_eval(t, doc, `count(//a[@x][1])`, float64(2))
	// With a single context node, the position is over the whole node-set.
	test_xpath_elements(t, doc, `/descendant::a[@x][1]`, 4)
	test_xpath_elements(t, doc, `(//a[@x])[2]`, 5)
}
