package xpath

import (
	"math"
	"testing"
)

// Some test examples from http://zvon.org/comp/r/ref-XPath_2.html

func Test_func_boolean(t *testing.T) {
	test_xpath_eval(t, empty_example, `true()`, true)
	test_xpath_eval(t, empty_example, `false()`, false)
	test_xpath_eval(t, empty_example, `boolean(0)`, false)
	test_xpath_eval(t, empty_example, `boolean(null)`, false)
	test_xpath_eval(t, empty_example, `boolean(1)`, true)
	test_xpath_eval(t, empty_example, `boolean(2)`, true)
	test_xpath_eval(t, empty_example, `boolean(true)`, false)
	test_xpath_eval(t, empty_example, `boolean(1 > 2)`, false)
	test_xpath_eval(t, book_example, `boolean(//*[@lang])`, true)
	test_xpath_eval(t, book_example, `boolean(//*[@x])`, false)
}

func Test_func_name(t *testing.T) {
	test_xpath_eval(t, html_example, `name(//html/@lang)`, "lang")
	test_xpath_eval(t, html_example, `name(html/head/title)`, "title")
	test_xpath_count(t, html_example, `//*[name() = "li"]`, 3)
}

func Test_func_not(t *testing.T) {
	//test_xpath_eval(t, empty_example, `not(0)`, true)
	//test_xpath_eval(t, empty_example, `not(1)`, false)
	test_xpath_elements(t, employee_example, `//employee[not(@id = "1")]`, 8, 13)
	test_xpath_elements(t, book_example, `//book[not(year = 2005)]`, 15, 25)
	test_xpath_count(t, book_example, `//book[not(title)]`, 0)
}

func Test_func_ceiling_floor(t *testing.T) {
	test_xpath_eval(t, empty_example, "ceiling(5.2)", float64(6))
	test_xpath_eval(t, empty_example, "floor(5.2)", float64(5))
}

func Test_func_concat(t *testing.T) {
	test_xpath_eval(t, empty_example, `concat("1", "2", "3")`, "123")
	//test_xpath_eval(t, empty_example, `concat("Ciao!", ())`, "Ciao!")
	test_xpath_eval(t, book_example, `concat(//book[1]/title, ", ", //book[1]/year)`, "Everyday Italian, 2005")
	result := concatFunc(testQuery("a"), testQuery("b"))(nil, nil).(string)
	assertEqual(t, result, "ab")
}

func Test_func_contains(t *testing.T) {
	test_xpath_eval(t, empty_example, `contains("tattoo", "t")`, true)
	test_xpath_eval(t, empty_example, `contains("tattoo", "T")`, false)
	test_xpath_eval(t, empty_example, `contains("tattoo", "ttt")`, false)
	//test_xpath_eval(t, empty_example, `contains("", ())`, true)
	test_xpath_elements(t, book_example, `//book[contains(title, "Potter")]`, 9)
	test_xpath_elements(t, book_example, `//book[contains(year, "2005")]`, 3, 9)
	assertPanic(t, func() { selectNode(html_example, "//*[contains(0, 0)]") })
}

func Test_func_count(t *testing.T) {
	test_xpath_eval(t, book_example, `count(//book)`, float64(4))
	test_xpath_eval(t, book_example, `count(//book[3]/author)`, float64(5))
}

func Test_func_ends_with(t *testing.T) {
	test_xpath_eval(t, empty_example, `ends-with("tattoo", "tattoo")`, true)
	test_xpath_eval(t, empty_example, `ends-with("tattoo", "atto")`, false)
	test_xpath_elements(t, book_example, `//book[ends-with(@category,'ing')]`, 3)
	test_xpath_elements(t, book_example, `//book[ends-with(./price,'.99')]`, 9, 15)
	assertPanic(t, func() { selectNode(html_example, `//*[ends-with(0, 0)]`) }) // arg must be start with string
	assertPanic(t, func() { selectNode(html_example, `//*[ends-with(name(), 0)]`) })
}

func Test_func_last(t *testing.T) {
	test_xpath_elements(t, book_example, `//bookstore[last()]`, 2)
	test_xpath_elements(t, book_example, `//bookstore/book[last()]`, 25)
	test_xpath_elements(t, book_example, `(//bookstore/book)[last()]`, 25)
	//https: //github.com/antchfx/xpath/issues/76
	test_xpath_elements(t, book_example, `(//bookstore/book[year = 2005])[last()]`, 9)
	test_xpath_elements(t, book_example, `//bookstore/book[year = 2005][last()]`, 9)
	test_xpath_elements(t, html_example, `//ul/li[last()]`, 15)
	test_xpath_elements(t, html_example, `(//ul/li)[last()]`, 15)
}

func Test_func_local_name(t *testing.T) {
	test_xpath_eval(t, book_example, `local-name(bookstore)`, "bookstore")
	test_xpath_eval(t, mybook_example, `local-name(//mybook:book)`, "book")
}

func Test_func_starts_with(t *testing.T) {
	test_xpath_eval(t, employee_example, `starts-with("tattoo", "tat")`, true)
	test_xpath_eval(t, employee_example, `starts-with("tattoo", "att")`, false)
	test_xpath_elements(t, book_example, `//book[starts-with(title,'Everyday')]`, 3)
	assertPanic(t, func() { selectNode(html_example, `//*[starts-with(0, 0)]`) })
	assertPanic(t, func() { selectNode(html_example, `//*[starts-with(name(), 0)]`) })
}

func Test_func_string(t *testing.T) {
	test_xpath_eval(t, empty_example, `string(1.23)`, "1.23")
	test_xpath_eval(t, empty_example, `string(3)`, "3")
	test_xpath_eval(t, book_example, `string(//book/@category)`, "cooking")
}

func Test_func_string_join(t *testing.T) {
	//(t, empty_example, `string-join(('Now', 'is', 'the', 'time', '...'), '')`, "Now is the time ...")
	test_xpath_eval(t, empty_example, `string-join("some text", ";")`, "some text")
	test_xpath_eval(t, book_example, `string-join(//book/@category, ";")`, "cooking;children;web;web")
}

func Test_func_string_length(t *testing.T) {
	test_xpath_eval(t, empty_example, `string-length("Harp not on that string, madam; that is past.")`, float64(45))
	test_xpath_eval(t, empty_example, `string-length(normalize-space(' abc '))`, float64(3))
	test_xpath_eval(t, html_example, `string-length(//title/text())`, float64(len("My page")))
	test_xpath_eval(t, html_example, `string-length(//html/@lang)`, float64(len("en")))
	test_xpath_count(t, employee_example, `//employee[string-length(@id) > 0]`, 3) // = //employee[@id]
}

func Test_func_substring(t *testing.T) {
	test_xpath_eval(t, empty_example, `substring("motor car", 6)`, " car")
	test_xpath_eval(t, empty_example, `substring("metadata", 4, 3)`, "ada")
	test_xpath_eval(t, empty_example, `substring("12345", 5, -3)`, "")
	test_xpath_eval(t, empty_example, `substring("12345", 1.5, 2.6)`, "234")
	test_xpath_eval(t, empty_example, `substring("12345", 0, 3)`, "12")
	test_xpath_eval(t, empty_example, `substring("12345", 5, -3)`, "")
	test_xpath_eval(t, empty_example, `substring("12345", 0, 5)`, "1234")
	test_xpath_eval(t, empty_example, `substring("12345", 1, 5)`, "12345")
	test_xpath_eval(t, empty_example, `substring("12345", 1, 6)`, "12345")
	test_xpath_eval(t, html_example, `substring(//title/child::node(), 1)`, "My page")
	//assertPanic(t, func() { selectNode(empty_example, `substring("12345", 5, "")`) })
}

func Test_func_substring_after(t *testing.T) {
	test_xpath_eval(t, empty_example, `substring-after("tattoo", "tat")`, "too")
	test_xpath_eval(t, empty_example, `substring-after("tattoo", "tattoo")`, "")
}

func Test_func_substring_before(t *testing.T) {
	test_xpath_eval(t, empty_example, `substring-before("tattoo", "attoo")`, "t")
	test_xpath_eval(t, empty_example, `substring-before("tattoo", "tatto")`, "")
}

func Test_func_sum(t *testing.T) {
	test_xpath_eval(t, empty_example, `sum(1 + 2)`, float64(3))
	test_xpath_eval(t, empty_example, `sum(1.1 + 2)`, float64(3.1))
	test_xpath_eval(t, book_example, `sum(//book/price)`, float64(149.93))
	test_xpath_elements(t, book_example, `//book[sum(./price) > 40]`, 15)
	assertPanic(t, func() { selectNode(html_example, `//title[sum('Hello') = 0]`) })
}

func Test_func_translate(t *testing.T) {
	test_xpath_eval(t, empty_example, `translate("bar","abc","ABC")`, "BAr")
	test_xpath_eval(t, empty_example, `translate("--aaa--","abc-","ABC")`, "AAA")
	test_xpath_eval(t, empty_example, `translate("abcdabc", "abc", "AB")`, "ABdAB")
	test_xpath_eval(t, empty_example, `translate('The quick brown fox', 'brown', 'red')`, "The quick red fdx")
}

func Test_func_matches(t *testing.T) {
	test_xpath_eval(t, empty_example, `matches("abracadabra", "bra")`, true)
	test_xpath_eval(t, empty_example, `matches("abracadabra", "(?i)^A.*A$")`, true)
	test_xpath_eval(t, empty_example, `matches("abracadabra", "^a.*a$")`, true)
	test_xpath_eval(t, empty_example, `matches("abracadabra", "^bra")`, false)
	assertPanic(t, func() { selectNode(html_example, `//*[matches()]`) })                   // arg len check failure
	assertPanic(t, func() { selectNode(html_example, "//*[matches(substring(), 0)]") })     // first arg processing failure
	assertPanic(t, func() { selectNode(html_example, "//*[matches(@href, substring())]") }) // second arg processing failure
	assertPanic(t, func() { selectNode(html_example, "//*[matches(@href, 0)]") })           // second arg not string
	assertPanic(t, func() { selectNode(html_example, "//*[matches(@href, '[invalid')]") })  // second arg invalid regexp
	// testing unexpected the regular expression.
	_, err := Compile(`//*[matches(., '^[\u0621-\u064AA-Za-z\-]+')]`)
	assertErr(t, err)
	_, err = Compile(`//*[matches(., '//*[matches(., '\w+`)
	assertErr(t, err)
}

func Test_func_number(t *testing.T) {
	test_xpath_eval(t, empty_example, `number(10)`, float64(10))
	test_xpath_eval(t, empty_example, `number(1.11)`, float64(1.11))
	test_xpath_eval(t, empty_example, `number("10") > 10`, false)
	test_xpath_eval(t, empty_example, `number("10") = 10`, true)
	test_xpath_eval(t, empty_example, `number("123") < 1000`, true)
	test_xpath_eval(t, empty_example, `number(//non-existent-node) = 0`, false)
	assertTrue(t, math.IsNaN(MustCompile(`number(//non-existent-node)`).Evaluate(createNavigator(empty_example)).(float64)))
	assertTrue(t, math.IsNaN(MustCompile(`number("123a")`).Evaluate(createNavigator(empty_example)).(float64)))
}

func Test_func_position(t *testing.T) {
	test_xpath_elements(t, book_example, `//book[position() = 1]`, 3)
	test_xpath_elements(t, book_example, `//book[(position() mod 2) = 0]`, 9, 25)
	test_xpath_elements(t, book_example, `//book[position() = last()]`, 25)
	test_xpath_elements(t, book_example, `//book/*[position() = 1]`, 4, 10, 16, 26)
	// Test Failed
	//test_xpath_elements(t, book_example, `(//book/title)[position() = 1]`, 3)
}

func Test_func_replace(t *testing.T) {
	test_xpath_eval(t, empty_example, `replace('aa-bb-cc','bb','ee')`, "aa-ee-cc")
	test_xpath_eval(t, empty_example, `replace("abracadabra", "bra", "*")`, "a*cada*")
	test_xpath_eval(t, empty_example, `replace("abracadabra", "a", "")`, "brcdbr")
	// The below xpath expressions is not supported yet
	//
	test_xpath_eval(t, empty_example, `replace("abracadabra", "a.*a", "*")`, "*")
	test_xpath_eval(t, empty_example, `replace("abracadabra", "a.*?a", "*")`, "*c*bra")
	// test_xpath_eval(t, empty_example, `replace("abracadabra", ".*?", "$1")`, "*c*bra") // error, because the pattern matches the zero-length string
	test_xpath_eval(t, empty_example, `replace("AAAA", "A+", "b")`, "b")
	test_xpath_eval(t, empty_example, `replace("AAAA", "A+?", "b")`, "bbbb")
	test_xpath_eval(t, empty_example, `replace("darted", "^(.*?)d(.*)$", "$1c$2")`, "carted")
	test_xpath_eval(t, empty_example, `replace("abracadabra", "a(.)", "a$1$1")`, "abbraccaddabbra")
	test_xpath_eval(t, empty_example, `replace("abcd", "(ab)|(a)", "[1=$1][2=$2]")`, "[1=ab][2=]cd")
	test_xpath_eval(t, empty_example, `replace("1/1/c11/1", "(.*)/[^/]+$", "$1")`, "1/1/c11")
	test_xpath_eval(t, empty_example, `replace("A/B/C/D/E/F/G/H/I/J/K/L", "([^/]*)/([^/]*)/([^/]*)/([^/]*)/([^/]*)/([^/]*)/([^/]*)/([^/]*)/([^/]*)/(.*)", "$1-$2-$3-$4-$5-$6-$7-$8-$9-$10")`, "A-B-C-D-E-F-G-H-I-J/K/L")
}

func Test_func_reverse(t *testing.T) {
	//test_xpath_eval(t, employee_example, `reverse(("hello"))`, "hello") // Not passed
	test_xpath_elements(t, employee_example, `reverse(//employee)`, 13, 8, 3)
	test_xpath_elements(t, employee_example, `//employee[reverse(.) = reverse(.)]`, 3, 8, 13)
	assertPanic(t, func() { selectNode(html_example, "reverse(concat())") }) // invalid node-sets argument.
	assertPanic(t, func() { selectNode(html_example, "reverse()") })         //  missing node-sets argument.
}

func Test_func_round(t *testing.T) {
	test_xpath_eval(t, employee_example, `round(2.5)`, 3) // int
	test_xpath_eval(t, employee_example, `round(2.5)`, 3)
	test_xpath_eval(t, employee_example, `round(2.4999)`, 2)
}

func Test_func_namespace_uri(t *testing.T) {
	test_xpath_eval(t, mybook_example, `namespace-uri(//mybook:book)`, "http://www.contoso.com/books")
	test_xpath_elements(t, mybook_example, `//*[namespace-uri()='http://www.contoso.com/books']`, 3, 9)
}

func Test_func_normalize_space(t *testing.T) {
	const testStr = "\t    \rloooooooonnnnnnngggggggg  \r \n tes  \u00a0 t strin \n\n \r g "
	const expectedStr = `loooooooonnnnnnngggggggg tes t strin g`
	test_xpath_eval(t, empty_example, `normalize-space("`+testStr+`")`, expectedStr)
	test_xpath_eval(t, empty_example, `normalize-space(' abc ')`, "abc")
	n := selectNode(employee_example, `//employee[@id="1"]/name`)
	test_xpath_eval(t, n, `normalize-space()`, "Opal Kole")
	test_xpath_eval(t, n, `normalize-space(.)`, "Opal Kole")
	test_xpath_eval(t, book_example, `normalize-space(//book/title)`, "Everyday Italian")
	test_xpath_eval(t, book_example, `normalize-space(//book[1]/title)`, "Everyday Italian")

}

func Test_func_lower_case(t *testing.T) {
	test_xpath_eval(t, empty_example, `lower-case("ABc!D")`, "abc!d")
	test_xpath_count(t, employee_example, `//name[@from="ca"]`, 0)
	test_xpath_elements(t, employee_example, `//name[lower-case(@from) = "ca"]`, 9)
	//test_xpath_eval(t, employee_example, `//employee/name/lower-case(text())`, "opal kole", "max miller", "beccaa moss")
}

func Benchmark_NormalizeSpaceFunc(b *testing.B) {
	b.ReportAllocs()
	const strForNormalization = "\t    \rloooooooonnnnnnngggggggg  \r \n tes  \u00a0 t strin \n\n \r g "
	for i := 0; i < b.N; i++ {
		_ = normalizespaceFunc(testQuery(strForNormalization))(nil, nil)
	}
}

func Benchmark_ConcatFunc(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = concatFunc(testQuery("a"), testQuery("b"))(nil, nil)
	}
}

func Benchmark_GetHashCode(b *testing.B) {
	// Create a more complex test node that will actually go through the full getHashCode paths
	doc := createNavigator(book_example)
	doc.MoveToRoot()
	doc.MoveToChild() // Move to the first element
	doc.MoveToChild() // Deeper
	// Find a node with attributes
	var node NodeNavigator
	for {
		if doc.NodeType() == AttributeNode || doc.NodeType() == TextNode || doc.NodeType() == CommentNode {
			node = doc.Copy()
			break
		}
		if !doc.MoveToNext() {
			if !doc.MoveToChild() {
				// If we can't find a suitable node, default to using the first element
				doc.MoveToRoot()
				doc.MoveToChild()
				node = doc.Copy()
				break
			}
		}
	}

	b.Run("getHashCode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			n := node.Copy()
			_ = getHashCode(n)
		}
	})
}
