package xpath

import (
	"testing"
)

// `*/employee` [Not supported]

func TestRelativePaths(t *testing.T) {
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
	test_xpath_tags(t, employee_example, `//@*`, "id", "discipline", "experience", "id", "from", "discipline", "experience", "id", "discipline")
	test_xpath_values(t, employee_example, `//@discipline`, "web", "DBA", "appdev")
	test_xpath_count(t, employee_example, `//employee/@id`, 3)
}

func TestExpressions(t *testing.T) {
	test_xpath_elements(t, book_example, `//book[@category = "cooking"] | //book[@category = "children"]`, 3, 9)
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
