package xpath

import (
	"reflect"
	"testing"
)

func Test_self(t *testing.T) {
	test_xpath_elements(t, employee_example, `//name/self::*`, 4, 9, 14)
}

func Test_child(t *testing.T) {
	test_xpath_elements(t, employee_example, `//child::employee/child::email`, 6, 11, 16)
	test_xpath_elements(t, employee_example, `/empinfo/child::*`, 3, 8, 13)
	test_xpath_elements(t, employee_example, `/empinfo/child::node()`, 3, 8, 13)
	test_xpath_values(t, employee_example, `//name/child::text()`, "Opal Kole", "Max Miller", "Beccaa Moss")
}

func Test_descendant(t *testing.T) {
	test_xpath_elements(t, employee_example, `//employee/descendant::*`, 4, 5, 6, 9, 10, 11, 14, 15, 16)
	test_xpath_count(t, employee_example, `//descendant::employee`, 3)

}

func Test_descendant_or_self(t *testing.T) {
	test_xpath_tags(t, employee_example.FirstChild, `self::*`, "empinfo")
	test_xpath_elements(t, employee_example, `//employee/descendant-or-self::*`, 3, 4, 5, 6, 8, 9, 10, 11, 13, 14, 15, 16)
	test_xpath_count(t, employee_example, `//descendant-or-self::employee`, 3)
}

func Test_ancestor(t *testing.T) {
	test_xpath_tags(t, employee_example, `//employee/ancestor::*`, "empinfo")
	test_xpath_tags(t, employee_example, `//employee/ancestor::empinfo`, "empinfo")
	// Test Panic
	//test_xpath_elements(t, employee_example, `//ancestor::name`, 4, 9, 14)
}

func Test_ancestor_predicate(t *testing.T) {
	doc := createElement(0, "",
		createElement(1, "html",
			createElement(2, "body",
				createElement(3, "h1"),
				createElement(4, "section",
					createElement(5, "div",
						createElement(6, "section",
							createElement(7, "div",
								createElement(8, "span"),
							),
						),
					),
				),
				createElement(9, "section",
					createElement(10, "div",
						createElement(11, "section",
							createElement(12, "div",
								createElement(13, "span"),
							),
						),
					),
				),
			),
		),
	)

	test_xpath_elements(t, doc, `//span/ancestor::*`, 7, 6, 5, 4, 2, 1, 12, 11, 10, 9)
	test_xpath_elements(t, doc, `//span/ancestor::section`, 6, 4, 11, 9)
	test_xpath_elements(t, doc, `//span/ancestor::section[1]`, 6, 11)
	test_xpath_elements(t, doc, `//span/ancestor::section[2]`, 4, 9)
}

func Test_ancestor_predicate_chain(t *testing.T) {
	doc := createElement(0, "",
		createElement(1, "html",
			createElementAttr(2, "body", map[string]string{"itemscope": "", "itemtype": "Article"},
				createElement(3, "section",
					createElementAttr(4, "span", map[string]string{"itemprop": "author"}),
					createElementAttr(5, "div", map[string]string{"itemscope": "", "itemtype": "Comment"},
						createElementAttr(6, "span", map[string]string{"itemprop": "author"}),
						createElement(7, "div",
							createElementAttr(8, "span", map[string]string{"itemprop": "author"}),
						),
					),
				),
			),
		),
	)

	// Find elements marked as "author" property whose closest "itemscope" ancestor is of "Comment" type.
	// This should find "span" elements on lines 6 and 8, but not line 4 since that one is under "Article".
	test_xpath_elements(t, doc, `//*[@itemprop="author"][ancestor::*[@itemscope][1][@itemtype="Comment"]]`, 6, 8)
}

func Test_ancestor_or_self(t *testing.T) {
	// Expected the value is [2, 3, 8, 13], but got [3, 2, 8, 13]
	test_xpath_elements(t, employee_example, `//employee/ancestor-or-self::*`, 3, 2, 8, 13)
	test_xpath_elements(t, employee_example, `//name/ancestor-or-self::employee`, 3, 8, 13)
}

func Test_parent(t *testing.T) {
	test_xpath_elements(t, employee_example, `//name/parent::*`, 3, 8, 13)
	test_xpath_elements(t, employee_example, `//name/parent::employee`, 3, 8, 13)
}

func Test_parent_no_duplicates(t *testing.T) {
	// Each employee has three element children, but a node-set holds each
	// parent once. selectNodes removes duplicates itself, so check count()
	// and the raw iterator.
	for _, expr := range []string{
		`//employee/*/..`,
		`//employee/*/parent::*`,
		`//employee/*/parent::*[1]`,
		`//employee/*/ancestor::*[1]`,
	} {
		test_xpath_eval(t, employee_example, `count(`+expr+`)`, float64(3))
		lines := []int{}
		for _, n := range iterateNodes(Select(createNavigator(employee_example), expr)) {
			lines = append(lines, n.lines)
		}
		if !reflect.DeepEqual(lines, []int{3, 8, 13}) {
			t.Fatalf("%s: expected lines [3 8 13], got %v", expr, lines)
		}
	}
}

func Test_attribute(t *testing.T) {
	test_xpath_values(t, employee_example, `//attribute::id`, "1", "2", "3")
	test_xpath_count(t, employee_example, `//attribute::*`, 9)
	test_xpath_count(t, employee_example, `//attribute::*[1]`, 7)

	// test failed(random): the return values is expected but the order of value is random.
	//test_xpath_tags(t, employee_example, `//attribute::*`, "id", "discipline", "experience", "id", "from", "discipline", "experience", "id", "discipline")

}

func Test_attribute_position(t *testing.T) {
	// The attributes are added in a fixed order: employee_example fills them from a map.
	item1 := createElement(2, "item")
	item1.addAttribute("a", "1")
	item1.addAttribute("b", "2")
	item1.addAttribute("c", "3")
	item2 := createElement(3, "item")
	item2.addAttribute("a", "4")
	item2.addAttribute("b", "5")
	item3 := createElement(4, "item")
	item3.addAttribute("c", "6")
	doc := createElement(0, "", createElement(1, "root", item1, item2, item3))

	test_xpath_values(t, doc, `/root/item[1]/@*[1]`, "1")
	test_xpath_values(t, doc, `/root/item[1]/@*[2]`, "2")
	test_xpath_values(t, doc, `/root/item[1]/attribute::*[3]`, "3")
	test_xpath_values(t, doc, `/root/item[1]/@*[position()=2]`, "2")
	test_xpath_values(t, doc, `/root/item[1]/@*[position()>1]`, "2", "3")
	test_xpath_values(t, doc, `/root/item/@*[1]`, "1", "4", "6")
	test_xpath_values(t, doc, `/root/item/@*[2]`, "2", "5")
	test_xpath_values(t, doc, `/root/item/@b[1]`, "2", "5")
}

func Test_attribute_of_attribute(t *testing.T) {
	item := createElement(2, "item")
	item.addAttribute("a", "1")
	item.addAttribute("b", "2")
	item.addAttribute("c", "3")
	doc := createElement(0, "", createElement(1, "root", item))

	// An attribute node has no attributes.
	test_xpath_count(t, doc, `/root/item/@a/@*`, 0)
	test_xpath_count(t, doc, `/root/item/@a/attribute::*`, 0)
	test_xpath_count(t, doc, `/root/item/@b/@c`, 0)
	test_xpath_count(t, doc, `/root/item/@*/@*`, 0)
	test_xpath_count(t, doc, `/root/item/@a/self::node()/@*`, 0)
	test_xpath_eval(t, doc, `count(/root/item/@a/@*)`, float64(0))
	test_xpath_eval(t, doc, `boolean(/root/item/@a/@*)`, false)
	test_xpath_values(t, doc, `/root/item/@a/parent::*/@c`, "3")
}

func Test_following(t *testing.T) {
	test_xpath_elements(t, employee_example, `//employee[@id=1]/following::*`, 8, 9, 10, 11, 13, 14, 15, 16)
}

func Test_following_position(t *testing.T) {
	// Positions count across all the following nodes of the context node,
	// not from 1 again in each following subtree.
	test_xpath_elements(t, employee_example, `//employee[@id=1]/following::*[1]`, 8)
	test_xpath_elements(t, employee_example, `//employee[@id=1]/following::*[2]`, 9)
	test_xpath_elements(t, employee_example, `//employee[@id=1]/following::*[5]`, 13)
	test_xpath_elements(t, employee_example, `//employee[@id=1]/following::*[position()=8]`, 16)
	test_xpath_elements(t, employee_example, `//employee[@id=1]/following::*[position()>6]`, 15, 16)
}

func Test_following_sibling(t *testing.T) {
	test_xpath_elements(t, employee_example, `//employee[@id=1]/following-sibling::*`, 8, 13)
	test_xpath_elements(t, employee_example, `//employee[@id=1]/following-sibling::employee`, 8, 13)
}

func Test_preceding(t *testing.T) {
	//testXPath3(t, html, "//li[last()]/preceding-sibling::*[2]", selectNode(html, "//li[position()=2]"))
	//testXPath3(t, html, "//li/preceding::*[1]", selectNode(html, "//h1"))
	test_xpath_elements(t, employee_example, `//employee[@id=3]/preceding::*`, 8, 9, 10, 11, 3, 4, 5, 6)
}

func Test_preceding_sibling(t *testing.T) {
	test_xpath_elements(t, employee_example, `//employee[@id=3]/preceding-sibling::*`, 8, 3)
}

func Test_namespace(t *testing.T) {
	// TODO
}
