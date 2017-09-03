package xquerytest

import (
	"github.com/antchfx/xpath"
	"testing"
)

// Public XML snippets constants so the caller can replace with adapted markup
const XML1 = `
<?xml version="1.0"?>
<root>
	<foo>
		<bar a="b"/>
	</foo>
	<li>
		<a class="tag" href=""/>
	</li>
	<p class='meta'>
		<span class='date-time'>
			<span class='date'></span>
			<span class='time'></span>
		</span>
		<span class='tags'>
			<span class='tag'>
				<a class="tag" href="../../../../../../../tags/en/index.html"></a>
			</span>
		</span>
	</p>
</root>
`

type Create func(markup string) (xpath.NodeNavigator, error)
type Enable func(markup string) bool

func EnableAll(markup string) bool {
	return true
}

func compilePathFunc(t *testing.T) func(path string) *xpath.Expr {
	return func(path string) *xpath.Expr {
		res := xpath.MustCompile(path)
		t.Logf("Compile xpath %v", res.DebugString())
		return res
	}
}

func checkIteratorFunc(t *testing.T) func(in interface{}) ([]xpath.NodeNavigator, bool) {
	return func(in interface{}) ([]xpath.NodeNavigator, bool) {
		switch in.(type) {
		case *xpath.NodeIterator:
			nn := in.(*xpath.NodeIterator)
			var res []xpath.NodeNavigator = []xpath.NodeNavigator{} //nn.Current().Copy()}
			for nn.MoveNext() {
				n := nn.Current()
				t.Logf("Found node %v", n)
				res = append(res, n.Copy())
			}
			return res, true
		default:
			t.Errorf("expected NodeIterator, got %#v", in)
			return nil, false
		}
	}
}

func createFunc(t *testing.T, givenCreateFunc Create) func(markup string) xpath.NodeNavigator {
	return func(markup string) xpath.NodeNavigator {
		nn, err := givenCreateFunc(markup)
		if err != nil {
			t.Fatal(err)
		}
		return nn
	}
}

func TestAll(t *testing.T, givenCreateFunc Create, enable Enable) {
	create := createFunc(t, givenCreateFunc)
	checkIterator := checkIteratorFunc(t)
	compilePath := compilePathFunc(t)

	if enable(XML1) {
		nodes, ok := checkIterator(compilePath("//foo").Evaluate(create(XML1)))
		if ok {
			if len(nodes) != 1 {
				t.Errorf("Find incorrect number of foo nodes %v", len(nodes))
			}
			if len(nodes) >= 1 {
				node := nodes[0]
				t.Logf("Find node %v", node)
				if node.LocalName() != "foo" {
					t.Errorf("Did not find correct element localName=%v", node.LocalName())
				}
			}
		}

		nodes, ok = checkIterator(compilePath("//foo/bar[@a='b']").Evaluate(create(XML1)))
		if ok {
			if len(nodes) != 1 {
				t.Errorf("Find incorrect number of foo nodes %v", len(nodes))
			}
			if len(nodes) >= 1 {
				node := nodes[0]
				t.Logf("Find node %v", node)
				if node.LocalName() != "bar" {
					t.Errorf("Did not find correct attribute %v", node.LocalName())
				}
			}
		}

		nodes, ok = checkIterator(compilePath("//foo/bar/@a").Evaluate(create(XML1)))
		if ok {
			if len(nodes) != 1 {
				t.Errorf("Find incorrect number of foo nodes %v", len(nodes))
			}
			if len(nodes) >= 1 {
				node := nodes[0]
				t.Logf("Find node %v", node)
				if node.LocalName() != "a" || node.Value() != "b" {
					t.Errorf("Did not find correct attribute %v=%v", node.LocalName(), node.Value())
				}
			}
		}

		nodes, ok = checkIterator(compilePath("//li").Evaluate(create(XML1)))
		if ok {
			if len(nodes) != 1 {
				t.Errorf("Find incorrect number of li nodes %v", len(nodes))
			}
			subnodes, _ := checkIterator(compilePath("./a").Evaluate(nodes[0]))
			if len(subnodes) != 1 {
				t.Errorf("Find incorrect number of ./a nodes inside //li %v", len(subnodes))
			}
		}

		nodes, ok = checkIterator(compilePath("//root/p[@class='meta']").Evaluate(create(XML1)))
		if ok {
			if len(nodes) != 1 {
				t.Errorf("Find incorrect number of foo nodes %v", len(nodes))
			}
		}

		nodes, ok = checkIterator(compilePath("./root").Evaluate(create(XML1)))
		if ok {
			if len(nodes) != 1 {
				t.Errorf("Find incorrect number of foo nodes %v", len(nodes))
			}
		}

	}
}
