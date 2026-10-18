package xpath

import "testing"

func TestStringPredicatesWithEmptyNodeSet(t *testing.T) {
	for _, function := range []string{"contains", "starts-with", "ends-with"} {
		t.Run(function, func(t *testing.T) {
			test_xpath_eval(t, empty_example, function+`(//missing, "")`, true)
			test_xpath_eval(t, empty_example, function+`(//missing, "x")`, false)
			test_xpath_eval(t, empty_example, function+`("", "")`, true)
		})
	}
}
