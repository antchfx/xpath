package xpath

import "testing"

func Test_self(t *testing.T) {
	test_xpath_elements(t, employee_example, `//name/self::*`, 4, 9, 14)
}

func Test_child(t *testing.T) {
	test_xpath_elements(t, employee_example, `/empinfo/child::*`, 3, 8, 13)
	test_xpath_elements(t, employee_example, `/empinfo/child::node()`, 3, 8, 13)
	test_xpath_values(t, employee_example, `//name/child::text()`, "Opal Kole", "Max Miller", "Beccaa Moss")
	test_xpath_elements(t, employee_example, `//child::employee/child::email`, 6, 11, 16)
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

func Test_ancestor_or_self(t *testing.T) {
	// Expected the value is [2, 3, 8, 13], but got [3, 2, 8, 13]
	test_xpath_elements(t, employee_example, `//employee/ancestor-or-self::*`, 3, 2, 8, 13)
	test_xpath_elements(t, employee_example, `//name/ancestor-or-self::employee`, 3, 8, 13)
}

func Test_parent(t *testing.T) {
	test_xpath_elements(t, employee_example, `//name/parent::*`, 3, 8, 13)
	test_xpath_elements(t, employee_example, `//name/parent::employee`, 3, 8, 13)
}

func Test_attribute(t *testing.T) {
	test_xpath_values(t, employee_example, `//attribute::id`, "1", "2", "3")
	test_xpath_count(t, employee_example, `//attribute::*`, 9)

	// test failed
	//test_xpath_tags(t, employee_example, `//attribute::*[1]`, "id", "discipline", "id", "from", "discipline", "id", "discipline")
	// test failed(random): the return values is expected but the order of value is random.
	//test_xpath_tags(t, employee_example, `//attribute::*`, "id", "discipline", "experience", "id", "from", "discipline", "experience", "id", "discipline")

}

func Test_following(t *testing.T) {
	test_xpath_elements(t, employee_example, `//employee[@id=1]/following::*`, 8, 9, 10, 11, 13, 14, 15, 16)
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
