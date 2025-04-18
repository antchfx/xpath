package xpath

import (
	"testing"
)

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
	test_xpath_elements(t, book_example, `//book[@category = "web"] and //book[price = "39.95"]`, 25)
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

// TestLibraryCrashMinimal isolates the crash observed with "//div[string()]"
// using only the htmlquery and xpath libraries directly.
// Adapted to use the test framework's TNode structure.
func TestLibraryCrashMinimal(t *testing.T) {
	// Create the equivalent of <div>hi</div> using TNode
	doc := createNode("", RootNode)
	div := doc.createChildNode("div", ElementNode)
	div.lines = 1 // Assign a line number for the test helper
	div.createChildNode("hi", TextNode)

	xpathStr := `//div[string()]`

	// Use the existing test helper to evaluate the XPath
	// The expression selects the div if its string value ("hi") is true in a boolean context.
	test_xpath_elements(t, doc, xpathStr, 1)
}
