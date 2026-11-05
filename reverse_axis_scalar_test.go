package xpath

import "testing"

func TestReverseAxisNodeSetScalarOrder(t *testing.T) {
	doc := createNode("", RootNode)
	root := doc.createChildNode("root", ElementNode)
	root.NamespaceURL = "urn:outer"
	root.createChildNode("10", TextNode)
	parent := root.createChildNode("parent", ElementNode)
	parent.NamespaceURL = "urn:inner"
	parent.createChildNode("200", TextNode)
	parent.createChildNode("leaf", ElementNode)
	root.createChildNode("value", ElementNode).createChildNode("1", TextNode)
	root.createChildNode("value", ElementNode).createChildNode("2", TextNode)
	for _, tc := range []struct {
		expr string
		want interface{}
	}{
		{`name(//leaf/ancestor::*)`, "root"},
		{`local-name(//leaf/ancestor::*)`, "root"},
		{`string(//leaf/ancestor::*)`, "10"},
		{`number(//leaf/ancestor::*)`, float64(10)},
		{`concat('value:', //leaf/ancestor::*)`, "value:10"},
		{`name(//leaf/ancestor::*[1])`, "parent"},
		{`namespace-uri(//leaf/ancestor::*)`, "urn:outer"},
		{`string-length(//leaf/ancestor::*)`, float64(2)},
		{`name((//leaf/ancestor::*))`, "root"},
		{`name((((//leaf/ancestor::*))))`, "root"},
		{`name(//leaf/ancestor::*[position()>0])`, "root"},
		{`string(reverse(//value))`, "2"},
		{`normalize-space(//leaf/ancestor::*)`, "10"},
		{`substring(//leaf/ancestor::*,1)`, "10"},
		{`substring-before(//leaf/ancestor::*, '0')`, "1"},
		{`substring-after(//leaf/ancestor::*, '1')`, "0"},
		{`matches(//leaf/ancestor::*, '^10$')`, true},
		{`ends-with(//leaf/ancestor::*, '10')`, true},
		{`contains(//leaf/ancestor::*, '10')`, true},
		{`string-join(//leaf/ancestor::*, //leaf/ancestor::*)`, "2001010"},
	} {
		t.Run(tc.expr, func(t *testing.T) { test_xpath_eval(t, doc, tc.expr, tc.want) })
	}
}

func TestScalarReverseAxisResetsPerContext(t *testing.T) {
	doc := createNode("", RootNode)
	root := doc.createChildNode("root", ElementNode)
	root.createChildNode("10", TextNode)
	for _, value := range []string{"20", "30"} {
		parent := root.createChildNode("parent", ElementNode)
		parent.createChildNode(value, TextNode)
		parent.createChildNode("leaf", ElementNode)
	}
	expr := MustCompile(`//leaf[string(ancestor::*)='10' and number(ancestor::*)=10 and name(ancestor::*[1])='parent']`)
	for i := 0; i < 3; i++ {
		it := expr.Select(createNavigator(doc))
		count := 0
		for it.MoveNext() {
			count++
		}
		if count != 2 {
			t.Errorf("evaluation%d matched%d contexts, want2", i, count)
		}
	}
	test_xpath_eval(t, doc, `string(//missing/ancestor::*)`, "")
}
