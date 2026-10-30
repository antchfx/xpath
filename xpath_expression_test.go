package xpath

import (
	"fmt"
	"math"
	"testing"
)

func TestPositionKeepsStepWhenPredicateHasAPath(t *testing.T) {
	root := createElement(1, "",
		createElement(2, "body",
			createElement(3, "section",
				createElementAttr(4, "div", map[string]string{"class": "test"}),
				createElementAttr(5, "div", map[string]string{"class": "test"}),
			),
			createElement(6, "section",
				createElementAttr(7, "div", map[string]string{"class": "other"}),
				createElementAttr(8, "div", map[string]string{"class": "test"}),
			),
		),
	)
	// position() is the position among the step's nodes, not among
	// attribute nodes compiled earlier in the same predicate.
	test_xpath_elements(t, root, `//div[@class='test' and position()=1]`, 4)
	test_xpath_elements(t, root, `//div[position()=1 and @class='test']`, 4)
	test_xpath_elements(t, root, `//div[@class='test' and position()=2]`, 5, 8)
	test_xpath_elements(t, root, `//div[@class='other' and position()=1]`, 7)
	test_xpath_elements(t, root, `//div[@class='test' and position()=last()]`, 5, 8)
	test_xpath_elements(t, root, `//div[position()=last() and @class='test']`, 5, 8)

	sections := createElement(1, "",
		createElement(2, "body",
			createElement(3, "section",
				createElement(4, "div"),
			),
			createElement(5, "section",
				createElement(6, "div"),
			),
		),
	)
	test_xpath_elements(t, sections, `//section[div and position()=1]`, 3)
	test_xpath_elements(t, sections, `//section[position()=1 and div]`, 3)
}

// https://github.com/antchfx/xpath/pull/145
func TestBooleanComparison(t *testing.T) {
	/*
		<a>
		  <b>1</b>
		  <b>0</b>
		  <b>foo</b>
		  <b></b>
		  <b>123</b>
		  <b>-1</b>
		  <empty/>
		</a>
	*/
	doc := createNode("", RootNode)
	a := doc.createChildNode("a", ElementNode)

	for _, v := range []string{"1", "0", "foo", "", "123", "-1"} {
		b := a.createChildNode("b", ElementNode)
		if v != "" {
			b.createChildNode(v, TextNode)
		}
	}
	a.createChildNode("empty", ElementNode)
	tests := []struct {
		expr     string
		expected int
	}{
		// boolean vs boolean
		{"//b[true()=1]", 6},
		{"//b[1=true()]", 6},
		{"//b[false()=0]", 6},
		{"//b[true()!='foo']", 0},
		{"//b[not(false()) = 1]", 6},
		{"//b[true()=0]", 0},
		{"//b[false()='foo']", 0},
		// number vs boolean
		{"//b[1 = true()]", 6},
		{"//b[0 = false()]", 6},
		{"//b[2 = true()]", 6},
		{"//b[0 = true()]", 0},
		{"//b[1 != false()]", 6},
		{"//b[0 != true()]", 6},
		{"//b[-1 = false()]", 0},
		{"//b[0.0 = false()]", 6},
		{"//b[1.5 = true()]", 6},
		// string vs boolean
		{"//b['1' = true()]", 6},
		{"//b['0' = false()]", 0},
		{"//b['123' = true()]", 6},
		{"//b['0' = true()]", 6},
		{"//b['foo' != true()]", 0},
		{"//b['' = false()]", 6},
		{"//b['' != false()]", 0},
		{"//b[' 1 ' = true()]", 6},
		{"//b['-0' = false()]", 0},
		// NodeSet vs boolean
		{"//b[. = true()]", 6},
		{"//b[. = false()]", 0},
		{"//b[. != true()]", 0},
		{"//b[//empty = false()]", 0},
		{"//b[//empty != true()]", 0},
		// relational comparisons keep the operand order
		{"//b[true() > 0]", 6},
		{"//b[true() >= 1]", 6},
		{"//b[false() < 1]", 6},
		{"//b[false() <= 0]", 6},
		{"//b[true() < 1]", 0},
		{"//b[false() > 0]", 0},
		{"//b[2 > 1 > 0]", 6},
		{"//b[false() < //b]", 6},
		{"//b[true() > //b]", 0},
		// relational comparisons between booleans compare them as numbers
		{"//b[true() > false()]", 6},
		{"//b[true() >= true()]", 6},
		{"//b[false() < true()]", 6},
		{"//b[false() <= false()]", 6},
		{"//b[true() < false()]", 0},
		{"//b[false() >= true()]", 0},
		{"//b[(1 < 2) > (2 < 1)]", 6},
		// a node-set is converted to a boolean, not to a number
		{"//b[. > false()]", 6},
		{"//b[. >= true()]", 6},
		{"//b[. < true()]", 0},
		{"//b[. <= false()]", 0},
		{"//b[false() < .]", 6},
		{"//b[true() > .]", 0},
		{"//b[//empty > false()]", 6},
		{"//b[//missing < true()]", 6},
		{"//b[true() > //missing]", 6},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			expr, err := Compile(tt.expr)
			if err != nil {
				t.Fatalf("failed to compile expr %q: %v", tt.expr, err)
			}
			iter := expr.Select(createNavigator(doc))
			count := 0
			for iter.MoveNext() {
				count++
			}

			if count != tt.expected {
				t.Errorf("expr %q: expected %d nodes, got %d", tt.expr, tt.expected, count)
			}
		})
	}
}

// https://github.com/antchfx/xpath/pull/142/
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

func TestUnaryMinus(t *testing.T) {
	test_xpath_eval(t, book_example, `-'2'`, float64(-2))
	test_xpath_eval(t, book_example, `--'2'`, float64(2))
	test_xpath_eval(t, book_example, `---'2'`, float64(-2))
	test_xpath_eval(t, book_example, `--true()`, float64(1))
	test_xpath_eval(t, book_example, `--//book[1]/price`, float64(30))
	test_xpath_eval(t, book_example, `string(--'02')`, "2")
	test_xpath_eval(t, book_example, `boolean(--'0')`, false)
	v, ok := MustCompile(`--'abc'`).Evaluate(createNavigator(book_example)).(float64)
	assertTrue(t, ok && math.IsNaN(v))
}

func Test_descendant_issue(t *testing.T) {
	// Issue #93 https://github.com/antchfx/xpath/issues/93
	/*
	   <div id="wrapper">
	     <span>span one</span>
	     <div>
	       <span>span two</span>
	     </div>
	   </div>
	*/
	doc := createNode("", RootNode)
	div := doc.createChildNode("div", ElementNode)
	div.lines = 1
	div.addAttribute("id", "wrapper")
	span := div.createChildNode("span", ElementNode)
	span.lines = 2
	span.createChildNode("span one", TextNode)
	div = div.createChildNode("div", ElementNode)
	div.lines = 3
	span = div.createChildNode("span", ElementNode)
	span.lines = 4
	span.createChildNode("span two", TextNode)

	test_xpath_elements(t, doc, `//div[@id='wrapper']/descendant::span[1]`, 2)
	test_xpath_elements(t, doc, `//div[@id='wrapper']//descendant::span[1]`, 2, 4)
}

// https://github.com/antchfx/htmlquery/issues/52

func TestRelativePaths(t *testing.T) {
	test_xpath_elements(t, book_example, `//bookstore`, 2)
	test_xpath_elements(t, book_example, `//book`, 3, 9, 15, 25)
	test_xpath_elements(t, book_example, `//bookstore/book`, 3, 9, 15, 25)
	test_xpath_tags(t, book_example, `//book/..`, "bookstore")
	test_xpath_elements(t, book_example, `//book[@category="cooking"]/..`, 2)
	test_xpath_elements(t, book_example, `//book/year[text() = 2005]/../..`, 2) // bookstore
	test_xpath_elements(t, book_example, `//book/year/../following-sibling::*`, 9, 15, 25)
	test_xpath_count(t, book_example, `//bookstore/book/*`, 20)
	test_xpath_tags(t, html_example, "//title/../..", "html")
	test_xpath_elements(t, html_example, "//ul/../p", 19)
}

func TestAbsolutePaths(t *testing.T) {
	test_xpath_elements(t, book_example, `bookstore`, 2)
	test_xpath_elements(t, book_example, `bookstore/book`, 3, 9, 15, 25)
	test_xpath_elements(t, book_example, `(bookstore/book)`, 3, 9, 15, 25)
	test_xpath_elements(t, book_example, `bookstore/book[2]`, 9)
	test_xpath_elements(t, book_example, `bookstore/book[last()]`, 25)
	test_xpath_elements(t, book_example, `bookstore/book[last()]/title`, 26)
	test_xpath_values(t, book_example, `/bookstore/book[last()]/title/text()`, "Learning XML")
	test_xpath_values(t, book_example, `/bookstore/book[@category = "children"]/year`, "2005")
	test_xpath_elements(t, book_example, `bookstore/book/..`, 2)
	test_xpath_elements(t, book_example, `/bookstore/*`, 3, 9, 15, 25)
	test_xpath_elements(t, book_example, `/bookstore/*/title`, 4, 10, 16, 26)
}

func TestAttributes(t *testing.T) {
	test_xpath_tags(t, html_example.FirstChild, "@*", "lang")
	test_xpath_count(t, employee_example, `//@*`, 9)
	test_xpath_values(t, employee_example, `//@discipline`, "web", "DBA", "appdev")
	test_xpath_count(t, employee_example, `//employee/@id`, 3)
}

func TestExpressions(t *testing.T) {
	test_xpath_elements(t, book_example, `//book[@category = "cooking"] | //book[@category = "children"]`, 3, 9)
	test_xpath_eval(t, book_example, `//book[@category = "web"] and //book[price = "39.95"]`, true)
	test_xpath_count(t, html_example, `//ul/*`, 3)
	test_xpath_count(t, html_example, `//ul/*/a`, 3)
	// Sequence
	//
	// table/tbody/tr/td/(para, .[not(para)], ..)
}

func TestSequence(t *testing.T) {
	// `//table/tbody/tr/td/(para, .[not(para)],..)`
	test_xpath_count(t, html_example, `//body/(h1, h2, p)`, 2)
	test_xpath_count(t, html_example, `//body/(h1, h2, p, ..)`, 3)
}

func TestLatinAttributesInXPath(t *testing.T) {
	doc := createNode("", RootNode)
	div := doc.createChildNode("div", ElementNode)
	div.addAttribute("language", "english")
	div.lines = 1
	test_xpath_elements(t, doc, `//div[@language='english']`, 1)
}

func TestCyrillicAttributesInXPath(t *testing.T) {
	doc := createNode("", RootNode)
	div := doc.createChildNode("div", ElementNode)
	div.addAttribute("язык", "русский")
	div.lines = 1
	test_xpath_elements(t, doc, `//div[@язык='русский']`, 1)
}

func TestGreekAttributesInXPath(t *testing.T) {
	doc := createNode("", RootNode)
	div := doc.createChildNode("div", ElementNode)
	div.addAttribute("γλώσσα", "ελληνικά")
	div.lines = 1
	test_xpath_elements(t, doc, `//div[@γλώσσα='ελληνικά']`, 1)
}

func TestCyrillicAndGreekAttributesMixedInXPath(t *testing.T) {
	doc := createNode("", RootNode)
	div := doc.createChildNode("div", ElementNode)
	div.addAttribute("язык", "русский")
	div.addAttribute("γλώσσα", "ελληνικά")
	div.lines = 1
	test_xpath_elements(t, doc, `//div[@язык='русский' and @γλώσσα='ελληνικά']`, 1)
}

func TestCyrillicAttributesInXPath_NoMatch(t *testing.T) {
	doc := createNode("", RootNode)
	div := doc.createChildNode("div", ElementNode)
	div.addAttribute("язык", "русский")
	div.lines = 1
	test_xpath_elements(t, doc, `//div[@язык='английский']`)
}

func TestGreekAttributesInXPath_NoMatch(t *testing.T) {
	doc := createNode("", RootNode)
	div := doc.createChildNode("div", ElementNode)
	div.addAttribute("γλώσσα", "ελληνικά")
	div.lines = 1
	test_xpath_elements(t, doc, `//div[@γλώσσα='αγγλικά']`)
}

func TestNonEnglishExpression(t *testing.T) {
	doc := createNode("", RootNode)
	n_1 := doc.createChildNode("Σειρά", ElementNode)
	n_1.lines = 1
	n_2 := n_1.createChildNode("ελληνικά", ElementNode)
	n_2.lines = 2
	n_2.createChildNode("hello", TextNode)
	test_xpath_elements(t, doc, "//Σειρά", 1)
	test_xpath_values(t, doc, "//Σειρά/ελληνικά", "hello")
}

func TestChineseCharactersExpression(t *testing.T) {
	doc := createNode("", RootNode)
	n := doc.createChildNode("中文", ElementNode)
	n.createChildNode("你好世界", TextNode)
	test_xpath_values(t, doc, "//中文", "你好世界")
}

func TestBUG_104(t *testing.T) {
	// BUG https://github.com/antchfx/xpath/issues/104
	test_xpath_count(t, book_example, `//author[1]`, 4)
	test_xpath_values(t, book_example, `//author[1]/text()`, "Giada De Laurentiis", "J K. Rowling", "James McGovern", "Erik T. Ray")
}

func TestNonEnglishPredicateExpression(t *testing.T) {
	// https://github.com/antchfx/xpath/issues/106
	doc := createNode("", RootNode)
	n := doc.createChildNode("h1", ElementNode)
	n.addAttribute("id", "断点")
	n.createChildNode("Hello,World!", TextNode)
	test_xpath_count(t, doc, "//h1[@id='断点']", 1)
}

func TestUnionDocumentOrder(t *testing.T) {
	// A union is a node-set, so its nodes come back in document order no
	// matter how the operands are written.
	test_xpath_elements(t, employee_example, `//email | //name`, 4, 6, 9, 11, 14, 16)
	test_xpath_elements(t, employee_example, `//name | //email | //employee/name`, 4, 6, 9, 11, 14, 16)
	test_xpath_elements(t, employee_example, `//email | (//designation | //name)`, 4, 5, 6, 9, 10, 11, 14, 15, 16)
	test_xpath_elements(t, employee_example, `(//email | //designation) | //name`, 4, 5, 6, 9, 10, 11, 14, 15, 16)
	test_xpath_elements(t, employee_example, `//email | //employee | /empinfo`, 2, 3, 6, 8, 11, 13, 16)
	test_xpath_elements(t, employee_example, `(//email | //name)[1]`, 4)
	test_xpath_elements(t, employee_example, `(//email | //name)[last()]`, 16)
	test_xpath_elements(t, employee_example, `(//email | //name)[position() = 2]`, 6)
	test_xpath_eval(t, employee_example, `string((//email | //name)[1])`, "Opal Kole")
	test_xpath_eval(t, employee_example, `string(//email | //name)`, "Opal Kole")
	test_xpath_eval(t, employee_example, `count(//email | //name)`, float64(6))
	// unions evaluated relative to each context node
	test_xpath_elements(t, employee_example, `//employee[string((email | name)[1]) = "Max Miller"]`, 8)
	// an attribute follows its element and precedes the element's children
	test_xpath_values(t, employee_example, `//designation | //@from`, "Senior Engineer", "CA", "DBA Engineer", "Application Developer")
	test_xpath_values(t, employee_example, `//@from | //name`, "Opal Kole", "Max Miller", "CA", "Beccaa Moss")
	test_xpath_values(t, employee_example, `//employee[2]/* | //employee[2]/@id`, "2", "Max Miller", "DBA Engineer", "maxmiller@email.com")
}
