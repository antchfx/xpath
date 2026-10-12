package xpath

import "testing"

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
