package xpath

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

var (
	html  = example()
	html2 = example2()
)

func TestCompile(t *testing.T) {
	var err error
	_, err = Compile("//title[@元素名称='初步诊断']")
	if err != nil {
		t.Fatalf("//title[@元素名称='初步诊断'] shoud be correct but got error %s", err)
	}
	_, err = Compile("//a")
	if err != nil {
		t.Fatalf("//a should be correct but got error %s", err)
	}
	_, err = Compile("//a[id=']/span")
	if err == nil {
		t.Fatal("//a[id=] should be got correct but is nil")
	}
	_, err = Compile("//ul/li/@class")
	if err != nil {
		t.Fatalf("//ul/li/@class should be correct but got error %s", err)
	}
	_, err = Compile("/a/b/(c, .[not(c)])")
	if err != nil {
		t.Fatalf("/a/b/(c, .[not(c)]) should be correct but got error %s", err)
	}
	_, err = Compile("/pre:foo")
	if err != nil {
		t.Fatalf("/pre:foo should be correct but got error %s", err)
	}
}

func TestCompileWithNS(t *testing.T) {
	_, err := CompileWithNS("/foo", nil)
	if err != nil {
		t.Fatalf("/foo {nil} should be correct but got error %s", err)
	}
	_, err = CompileWithNS("/foo", map[string]string{})
	if err != nil {
		t.Fatalf("/foo {} should be correct but got error %s", err)
	}
	_, err = CompileWithNS("/foo", map[string]string{"a": "b"})
	if err != nil {
		t.Fatalf("/foo {a:b} should be correct but got error %s", err)
	}
	_, err = CompileWithNS("/a:foo", map[string]string{"a": "b"})
	if err != nil {
		t.Fatalf("/a:foo should be correct but got error %s", err)
	}
	_, err = CompileWithNS("/u:foo", map[string]string{"a": "b"})
	msg := fmt.Sprintf("%v", err)
	if msg != "prefix u not defined." {
		t.Fatalf("expected 'prefix u not defined' but got: %s", msg)
	}
}

func TestNamespace(t *testing.T) {
	doc := createNode("", RootNode)
	books := doc.createChildNode("books", ElementNode)
	book1 := books.createChildNode("book", ElementNode)
	book1.createChildNode("book1", TextNode)
	book2 := books.createChildNode("b:book", ElementNode)
	book2.addAttribute("xmlns:b", "ns")
	book2.createChildNode("book2", TextNode)
	book3 := books.createChildNode("c:book", ElementNode)
	book3.addAttribute("xmlns:c", "ns")
	book3.createChildNode("book3", TextNode)

	// Existing behaviour:
	v := joinValues(selectNodes(doc, "//b:book"))
	if v != "book2" {
		t.Fatalf("expected book2 but got %s", v)
	}

	// With namespace bindings:
	exp, _ := CompileWithNS("//x:book", map[string]string{"x": "ns"})
	v = joinValues(iterateNodes(exp.Select(createNavigator(doc))))
	if v != "book2,book3" {
		t.Fatalf("expected 'book2,book3' but got %s", v)
	}
}

func TestMustCompile(t *testing.T) {
	expr := MustCompile("//")
	if expr == nil {
		t.Fatal("// should be compiled but got nil object")
	}

	if wanted := (nopQuery{}); expr.q != wanted {
		t.Fatalf("wanted nopQuery object but got %s", expr)
	}
	iter := expr.Select(createNavigator(html))
	if iter.MoveNext() {
		t.Fatal("should be an empty node list but got one")
	}
}

func TestSelf(t *testing.T) {
	testXPath(t, html, ".", "html")
	testXPath(t, html.FirstChild, ".", "head")
	testXPath(t, html, "self::*", "html")
	testXPath(t, html.LastChild, "self::body", "body")
	testXPath2(t, html, "//body/./ul/li/a", 3)
}

func TestLastFunc(t *testing.T) {
	testXPath3(t, html, "/head[last()]", html.FirstChild)
	ul := selectNode(html, "//ul")
	testXPath3(t, html, "//li[last()]", ul.LastChild)
	list := selectNodes(html, "//li/a[last()]")
	if got := len(list); got != 3 {
		t.Fatalf("expected %d, but got %d", 3, got)
	}
	testXPath3(t, html, "(//ul/li)[last()]", ul.LastChild)

	n := selectNode(html, "//meta[@name][last()]")
	if n == nil {
		t.Fatal("should found one, but got nil")
	}
	if expected, value := "description", n.getAttribute("name"); expected != value {
		t.Fatalf("expected, %s but got %s", expected, value)
	}

}

func TestParent(t *testing.T) {
	testXPath(t, html.LastChild, "..", "html")
	testXPath(t, html.LastChild, "parent::*", "html")
	a := selectNode(html, "//li/a")
	testXPath(t, a, "parent::*", "li")
	testXPath(t, html, "//title/parent::head", "head")
}

func TestAttribute(t *testing.T) {
	testXPath(t, html, "@lang='en'", "html")
	testXPath2(t, html, "@lang='zh'", 0)
	testXPath2(t, html, "//@href", 3)
	testXPath2(t, html, "//a[@*]", 3)
}

func TestSequence(t *testing.T) {
	testXPath2(t, html2, "//table/tbody/tr/td/(para, .[not(para)])", 9)
	testXPath2(t, html2, "//table/tbody/tr/td/(para, .[not(para)], ..)", 12)
}

func TestRelativePath(t *testing.T) {
	testXPath(t, html, "head", "head")
	testXPath(t, html, "/head", "head")
	testXPath(t, html, "body//li", "li")
	testXPath(t, html, "/head/title", "title")

	testXPath2(t, html, "/body/ul/li/a", 3)
	testXPath(t, html, "//title", "title")
	testXPath(t, html, "//title/..", "head")
	testXPath(t, html, "//title/../..", "html")
	testXPath2(t, html, "//a[@href]", 3)
	testXPath(t, html, "//ul/../footer", "footer")
}

func TestChild(t *testing.T) {
	testXPath(t, html, "/child::head", "head")
	testXPath(t, html, "/child::head/child::title", "title")
	testXPath(t, html, "//title/../child::title", "title")
	testXPath(t, html.Parent, "//child::*", "html")
}

func TestDescendant(t *testing.T) {
	testXPath2(t, html, "descendant::*", 16)
	testXPath2(t, html, "/head/descendant::*", 3)
	testXPath2(t, html, "//ul/descendant::*", 7)  // <li> + <a>
	testXPath2(t, html, "//ul/descendant::li", 4) // <li>
}

func TestAncestor(t *testing.T) {
	testXPath2(t, html, "/body/footer/ancestor::*", 2) // body>html
	testXPath2(t, html, "/body/ul/li/a/ancestor::li", 3)
	testXPath2(t, html, "/body/ul/li/a/ancestor-or-self::li", 3)
}

func TestFollowingSibling(t *testing.T) {
	var list []*TNode
	list = selectNodes(html2, "//h1/span/following-sibling::text()")
	for _, n := range list {
		if n.Type != TextNode {
			t.Errorf("expected node is text but got:%s nodes %d", n.Data, len(list))
		}
	}
	list = selectNodes(html, "//li/following-sibling::*")
	for _, n := range list {
		if n.Data != "li" {
			t.Fatalf("expected node is li,but got:%s", n.Data)
		}
	}

	list = selectNodes(html, "//ul/following-sibling::*") // p,footer
	for _, n := range list {
		if n.Data != "p" && n.Data != "footer" {
			t.Fatal("expected node is not one of the following nodes: [p,footer]")
		}
	}
	testXPath(t, html, "//ul/following-sibling::footer", "footer")
	list = selectNodes(html, "//h1/following::*") // ul>li>a,p,footer
	if list[0].Data != "ul" {
		t.Fatal("expected node is not ul")
	}
	if list[1].Data != "li" {
		t.Fatal("expected node is not li")
	}
	if list[len(list)-1].Data != "footer" {
		t.Fatal("expected node is not footer")
	}
}

func TestPrecedingSibling(t *testing.T) {
	testXPath(t, html, "/body/footer/preceding-sibling::*", "p")
	testXPath2(t, html, "/body/footer/preceding-sibling::*", 3) // p,ul,h1
	list := selectNodes(html, "//h1/preceding::*")              // head>title>meta
	if list[0].Data != "head" {
		t.Fatal("expected is not head")
	}
	if list[1].Data != "title" {
		t.Fatal("expected is not title")
	}
	if list[2].Data != "meta" {
		t.Fatal("expected is not meta")
	}
}

func TestStarWide(t *testing.T) {
	testXPath(t, html, "/head/*", "title")
	testXPath2(t, html, "//ul/*", 4)
	testXPath(t, html, "@*", "html")
	testXPath2(t, html, "/body/h1/*", 0)
	testXPath2(t, html, `//ul/*/a`, 3)
}

func TestNodeTestType(t *testing.T) {
	testXPath(t, html, "//title/text()", "Hello")
	testXPath(t, html, "//a[@href='/']/text()", "Home")
	testXPath2(t, html, "//head/node()", 3)
	testXPath2(t, html, "//ul/node()", 4)
}

func TestPosition(t *testing.T) {
	testXPath3(t, html, "/head[1]", html.FirstChild) // compare to 'head' element
	ul := selectNode(html, "//ul")

	testXPath3(t, html, "//li[1]", ul.FirstChild)
	testXPath3(t, html, "//li[4]", ul.LastChild)
	testXPath3(t, html, "//li[last()]", ul.LastChild)
	testXPath2(t, html2, "//td[2]", 3)
}

func TestPredicate(t *testing.T) {
	testXPath(t, html.Parent, "html[@lang='en']", "html")
	testXPath(t, html, "//a[@href='/']", "a")
	testXPath(t, html, "//meta[@name]", "meta")
	ul := selectNode(html, "//ul")
	testXPath3(t, html, "//li[position()=4]", ul.LastChild)
	testXPath3(t, html, "//li[position()=1]", ul.FirstChild)
	testXPath2(t, html, "//li[position()>0]", 4)
	testXPath3(t, html, "//a[text()='Home']", selectNode(html, "//a[1]"))
}

func TestOr_And(t *testing.T) {
	list := selectNodes(html, "//h1|//footer")
	if len(list) == 0 {
		t.Fatal("//h1|//footer no any node found")
	}
	if list[0].Data != "h1" {
		t.Fatalf("expected first node of node-set is h1,but got %s", list[0].Data)
	}
	if list[1].Data != "footer" {
		t.Fatalf("expected first node of node-set is footer,but got %s", list[1].Data)
	}

	list = selectNodes(html, "//a[@id=1 or @id=2]")
	if list[0] != selectNode(html, "//a[@id=1]") {
		t.Fatal("node is not equal")
	}
	if list[1] != selectNode(html, "//a[@id=2]") {
		t.Fatal("node is not equal")
	}
	list = selectNodes(html, "//a[@id or @href]")
	if list[0] != selectNode(html, "//a[@id=1]") {
		t.Fatal("node is not equal")
	}
	if list[1] != selectNode(html, "//a[@id=2]") {
		t.Fatal("node is not equal")
	}
	testXPath3(t, html, "//a[@id=1 and @href='/']", selectNode(html, "//a[1]"))
	testXPath3(t, html, "//a[text()='Home' and @id='1']", selectNode(html, "//a[1]"))
}

func TestFunction(t *testing.T) {
	testEval(t, html, "boolean(//*[@id])", true)
	testEval(t, html, "boolean(//*[@x])", false)
	testEval(t, html, "name(//title)", "title")
	testXPath2(t, html, "//*[name()='a']", 3)
	testXPath(t, html, "//*[starts-with(name(),'h1')]", "h1")
	testXPath(t, html, "//*[ends-with(name(),'itle')]", "title") // Head title
	testXPath2(t, html, "//*[contains(@href,'a')]", 2)
	testXPath2(t, html, "//*[starts-with(@href,'/a')]", 2)            // a links: `/account`,`/about`
	testXPath2(t, html, "//*[ends-with(@href,'t')]", 2)               // a links: `/account`,`/about`
	testXPath2(t, html, "//*[matches(@href,'(?i)^.*OU[A-Z]?T$')]", 2) // a links: `/account`,`/about`. Note use of `(?i)`
	testXPath3(t, html, "//h1[normalize-space(text())='This is a H1']", selectNode(html, "//h1"))
	testXPath3(t, html, "//title[substring(.,1)='Hello']", selectNode(html, "//title"))
	testXPath3(t, html, "//title[substring(text(),1,4)='Hell']", selectNode(html, "//title"))
	testXPath3(t, html, "//title[substring(self::*,1,4)='Hell']", selectNode(html, "//title"))
	testXPath2(t, html, "//title[substring(child::*,1)]", 0) // Here substring return boolen (false), should it?
	testXPath2(t, html, "//title[substring(child::*,1) = '']", 1)
	testXPath3(t, html, "//li[not(a)]", selectNode(html, "//ul/li[4]"))
	testXPath2(t, html, "//li/a[not(@id='1')]", 2) //  //li/a[@id!=1]
	testXPath2(t, html, "//h1[string-length(normalize-space(' abc ')) = 3]", 1)
	testXPath2(t, html, "//h1[string-length(normalize-space(self::text())) = 12]", 1)
	testXPath2(t, html, "//title[string-length(normalize-space(child::*)) = 0]", 1)
	testXPath2(t, html, "//title[string-length(self::text()) = 5]", 1) // Hello = 5
	testXPath2(t, html, "//title[string-length(child::*) = 5]", 0)
	testXPath2(t, html, "//ul[count(li)=4]", 1)
	testEval(t, html, "true()", true)
	testEval(t, html, "false()", false)
	testEval(t, html, "boolean(0)", false)
	testEval(t, html, "boolean(1)", true)
	testEval(t, html, "sum(1+2)", float64(3))
	testEval(t, html, "string(sum(1+2))", "3")
	testEval(t, html, "sum(1.1+2)", float64(3.1))
	testEval(t, html, "sum(//a/@id)", float64(6)) // 1+2+3
	testEval(t, html, `concat("1","2","3")`, "123")
	testEval(t, html, `concat(" ",//a[@id='1']/@href," ")`, " / ")
	testEval(t, html, "ceiling(5.2)", float64(6))
	testEval(t, html, "floor(5.2)", float64(5))
	testEval(t, html, `substring-before('aa-bb','-')`, "aa")
	testEval(t, html, `substring-before('aa-bb','a')`, "")
	testEval(t, html, `substring-before('aa-bb','b')`, "aa-")
	testEval(t, html, `substring-before('aa-bb','q')`, "")
	testEval(t, html, `substring-after('aa-bb','-')`, "bb")
	testEval(t, html, `substring-after('aa-bb','a')`, "a-bb")
	testEval(t, html, `substring-after('aa-bb','b')`, "b")
	testEval(t, html, `substring-after('aa-bb','q')`, "")
	testEval(t, html, `replace('aa-bb-cc','bb','ee')`, "aa-ee-cc")
	testEval(t, html,
		`translate('The quick brown fox.', 'abcdefghijklmnopqrstuvwxyz', 'ABCDEFGHIJKLMNOPQRSTUVWXYZ')`,
		"THE QUICK BROWN FOX.",
	)
	testEval(t, html,
		`translate('The quick brown fox.', 'brown', 'red')`,
		"The quick red fdx.",
	)
	// preceding-sibling::*
	testXPath3(t, html, "//li[last()]/preceding-sibling::*[2]", selectNode(html, "//li[position()=2]"))
	// preceding::
	testXPath3(t, html, "//li/preceding::*[1]", selectNode(html, "//h1"))
}

func TestTransformFunctionReverse(t *testing.T) {
	nodes := selectNodes(html, "reverse(//li)")
	expectedReversedNodeValues := []string{"", "login", "about", "Home"}
	if len(nodes) != len(expectedReversedNodeValues) {
		t.Fatalf("reverse(//li) should return %d <li> nodes", len(expectedReversedNodeValues))
	}
	for i := 0; i < len(expectedReversedNodeValues); i++ {
		if nodes[i].Value() != expectedReversedNodeValues[i] {
			t.Fatalf("reverse(//li)[%d].Value() should be '%s', instead, got '%s'",
				i, expectedReversedNodeValues[i], nodes[i].Value())
		}
	}

	// Although this xpath itself doesn't make much sense, it does exercise the call path to provide coverage
	// for transformFunctionQuery.Evaluate()
	testXPath2(t, html, "//h1[reverse(.) = reverse(.)]", 1)

	// Test reverse() parsing error: missing node-sets argument.
	assertPanic(t, func() { testXPath2(t, html, "reverse()", 0) })
	// Test reverse() parsing error: invalid node-sets argument.
	assertPanic(t, func() { testXPath2(t, html, "reverse(concat())", 0) })
}

func TestPanic(t *testing.T) {
	// starts-with
	assertPanic(t, func() { testXPath(t, html, "//*[starts-with(0, 0)]", "") })
	assertPanic(t, func() { testXPath(t, html, "//*[starts-with(name(), 0)]", "") })
	//ends-with
	assertPanic(t, func() { testXPath(t, html, "//*[ends-with(0, 0)]", "") })
	assertPanic(t, func() { testXPath(t, html, "//*[ends-with(name(), 0)]", "") })
	// contains
	assertPanic(t, func() { testXPath2(t, html, "//*[contains(0, 0)]", 0) })
	assertPanic(t, func() { testXPath2(t, html, "//*[contains(@href, 0)]", 0) })
	// matches
	assertPanic(t, func() { testXPath2(t, html, "//*[matches()]", 0) })                   // arg len check failure
	assertPanic(t, func() { testXPath2(t, html, "//*[matches(substring(), 0)]", 0) })     // first arg processing failure
	assertPanic(t, func() { testXPath2(t, html, "//*[matches(@href, substring())]", 0) }) // second arg processing failure
	assertPanic(t, func() { testXPath2(t, html, "//*[matches(@href, 0)]", 0) })           // second arg not string
	assertPanic(t, func() { testXPath2(t, html, "//*[matches(@href, '[invalid')]", 0) })  // second arg invalid regexp
	// sum
	assertPanic(t, func() { testXPath3(t, html, "//title[sum('Hello') = 0]", nil) })
	// substring
	assertPanic(t, func() { testXPath3(t, html, "//title[substring(.,'')=0]", nil) })
	assertPanic(t, func() { testXPath3(t, html, "//title[substring(.,4,'')=0]", nil) })
	assertPanic(t, func() { testXPath3(t, html, "//title[substring(.,4,4)=0]", nil) })
	//assertPanic(t, func() { testXPath2(t, html, "//title[substring(child::*,0) = '']", 0) }) // Here substring return boolen (false), should it?

}

func TestEvaluate(t *testing.T) {
	testEval(t, html, "count(//ul/li)", float64(4))
	testEval(t, html, "//html/@lang", []string{"en"})
	testEval(t, html, "//title/text()", []string{"Hello"})
}

func TestOperationOrLogical(t *testing.T) {
	testXPath3(t, html, "//li[1+1]", selectNode(html, "//li[2]"))
	testXPath3(t, html, "//li[5 div 2]", selectNode(html, "//li[2]"))
	testXPath3(t, html, "//li[3 mod 2]", selectNode(html, "//li[1]"))
	testXPath3(t, html, "//li[3 - 2]", selectNode(html, "//li[1]"))
	testXPath2(t, html, "//li[position() mod 2 = 0 ]", 2) // //li[2],li[4]
	testXPath2(t, html, "//a[@id>=1]", 3)                 // //a[@id>=1] == a[1],a[2],a[3]
	testXPath2(t, html, "//a[@id<=2]", 2)                 // //a[@id<=2] == a[1],a[1]
	testXPath2(t, html, "//a[@id<2]", 1)                  // //a[@id>=1] == a[1]
	testXPath2(t, html, "//a[@id!=2]", 2)                 // //a[@id>=1] == a[1],a[3]
	testXPath2(t, html, "//a[@id=1 or @id=3]", 2)         // //a[@id>=1] == a[1],a[3]
	testXPath3(t, html, "//a[@id=1 and @href='/']", selectNode(html, "//a[1]"))
}

func testEval(t *testing.T, root *TNode, expr string, expected interface{}) {
	v := MustCompile(expr).Evaluate(createNavigator(root))
	if it, ok := v.(*NodeIterator); ok {
		exp, ok := expected.([]string)
		if !ok {
			t.Fatalf("expected value, got: %#v", v)
		}
		got := iterateNavs(it)
		if len(exp) != len(got) {
			t.Fatalf("expected: %#v, got: %#v", exp, got)
		}
		for i, n1 := range exp {
			n2 := got[i]
			if n1 != n2.Value() {
				t.Fatalf("expected: %#v, got: %#v", n1, n2)
			}
		}
		return
	}
	if v != expected {
		t.Fatalf("expected: %#v, got: %#v", expected, v)
	}
}

func testXPath(t *testing.T, root *TNode, expr string, expected string) {
	node := selectNode(root, expr)
	if node == nil {
		t.Fatalf("`%s` returns node is nil", expr)
	}
	if node.Data != expected {
		t.Fatalf("`%s` expected node is %s,but got %s", expr, expected, node.Data)
	}
}

func testXPath2(t *testing.T, root *TNode, expr string, expected int) {
	list := selectNodes(root, expr)
	if len(list) != expected {
		t.Fatalf("`%s` expected node numbers is %d,but got %d", expr, expected, len(list))
	}
}

func testXPath3(t *testing.T, root *TNode, expr string, expected *TNode) {
	node := selectNode(root, expr)
	if node == nil {
		t.Fatalf("`%s` returns node is nil", expr)
	}
	if node != expected {
		t.Fatalf("`%s` %s != %s", expr, node.Value(), expected.Value())
	}
}

func testXPath4(t *testing.T, root *TNode, expr string, expected string) {
	node := selectNode(root, expr)
	if node == nil {
		t.Fatalf("`%s` returns node is nil", expr)
	}
	if got := node.Value(); got != expected {
		t.Fatalf("`%s` expected \n%s,but got\n%s", expr, expected, got)
	}
}

func iterateNavs(t *NodeIterator) []*TNodeNavigator {
	var nodes []*TNodeNavigator
	for t.MoveNext() {
		node := t.Current().(*TNodeNavigator)
		nodes = append(nodes, node)
	}
	return nodes
}

func iterateNodes(t *NodeIterator) []*TNode {
	var nodes []*TNode
	for t.MoveNext() {
		node := (t.Current().(*TNodeNavigator)).curr
		nodes = append(nodes, node)
	}
	return nodes
}

func selectNode(root *TNode, expr string) (n *TNode) {
	t := Select(createNavigator(root), expr)
	if t.MoveNext() {
		n = (t.Current().(*TNodeNavigator)).curr
	}
	return n
}

func selectNodes(root *TNode, expr string) []*TNode {
	t := Select(createNavigator(root), expr)
	return iterateNodes(t)
}

func joinValues(nodes []*TNode) string {
	s := make([]string, 0)
	for _, n := range nodes {
		s = append(s, n.Value())
	}
	return strings.Join(s, ",")
}

func createNavigator(n *TNode) *TNodeNavigator {
	return &TNodeNavigator{curr: n, root: n, attr: -1}
}

type Attribute struct {
	Key, Value string
}

type TNode struct {
	Parent, FirstChild, LastChild, PrevSibling, NextSibling *TNode

	Type NodeType
	Data string
	Attr []Attribute
}

func (n *TNode) Value() string {
	if n.Type == TextNode {
		return n.Data
	}

	var buff bytes.Buffer
	var output func(*TNode)
	output = func(node *TNode) {
		if node.Type == TextNode {
			buff.WriteString(node.Data)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			output(child)
		}
	}
	output(n)
	return buff.String()
}

// TNodeNavigator is for navigating TNode.
type TNodeNavigator struct {
	curr, root *TNode
	attr       int
}

func (n *TNodeNavigator) NodeType() NodeType {
	if n.curr.Type == ElementNode && n.attr != -1 {
		return AttributeNode
	}
	return n.curr.Type
}

func (n *TNodeNavigator) LocalName() string {
	if n.attr != -1 {
		return n.curr.Attr[n.attr].Key
	}
	name := n.curr.Data
	if strings.Contains(name, ":") {
		return strings.Split(name, ":")[1]
	}
	return name
}

func (n *TNodeNavigator) Prefix() string {
	if n.attr == -1 && strings.Contains(n.curr.Data, ":") {
		return strings.Split(n.curr.Data, ":")[0]
	}
	return ""
}

func (n *TNodeNavigator) NamespaceURL() string {
	if n.Prefix() != "" {
		for _, a := range n.curr.Attr {
			if a.Key == "xmlns:"+n.Prefix() {
				return a.Value
			}
		}
	}
	return ""
}

func (n *TNodeNavigator) Value() string {
	switch n.curr.Type {
	case CommentNode:
		return n.curr.Data
	case ElementNode:
		if n.attr != -1 {
			return n.curr.Attr[n.attr].Value
		}
		var buf bytes.Buffer
		node := n.curr.FirstChild
		for node != nil {
			if node.Type == TextNode {
				buf.WriteString(strings.TrimSpace(node.Data))
			}
			node = node.NextSibling
		}
		return buf.String()
	case TextNode:
		return n.curr.Data
	}
	return ""
}

func (n *TNodeNavigator) Copy() NodeNavigator {
	n2 := *n
	return &n2
}

func (n *TNodeNavigator) MoveToRoot() {
	n.curr = n.root
}

func (n *TNodeNavigator) MoveToParent() bool {
	if node := n.curr.Parent; node != nil {
		n.curr = node
		return true
	}
	return false
}

func (n *TNodeNavigator) MoveToNextAttribute() bool {
	if n.attr >= len(n.curr.Attr)-1 {
		return false
	}
	n.attr++
	return true
}

func (n *TNodeNavigator) MoveToChild() bool {
	if node := n.curr.FirstChild; node != nil {
		n.curr = node
		return true
	}
	return false
}

func (n *TNodeNavigator) MoveToFirst() bool {
	if n.curr.PrevSibling == nil {
		return false
	}
	for {
		node := n.curr.PrevSibling
		if node == nil {
			break
		}
		n.curr = node
	}
	return true
}

func (n *TNodeNavigator) String() string {
	return n.Value()
}

func (n *TNodeNavigator) MoveToNext() bool {
	if node := n.curr.NextSibling; node != nil {
		n.curr = node
		return true
	}
	return false
}

func (n *TNodeNavigator) MoveToPrevious() bool {
	if node := n.curr.PrevSibling; node != nil {
		n.curr = node
		return true
	}
	return false
}

func (n *TNodeNavigator) MoveTo(other NodeNavigator) bool {
	node, ok := other.(*TNodeNavigator)
	if !ok || node.root != n.root {
		return false
	}

	n.curr = node.curr
	n.attr = node.attr
	return true
}

func createNode(data string, typ NodeType) *TNode {
	return &TNode{Data: data, Type: typ, Attr: make([]Attribute, 0)}
}

func (n *TNode) createChildNode(data string, typ NodeType) *TNode {
	m := createNode(data, typ)
	m.Parent = n
	if n.FirstChild == nil {
		n.FirstChild = m
	} else {
		n.LastChild.NextSibling = m
		m.PrevSibling = n.LastChild
	}
	n.LastChild = m
	return m
}

func (n *TNode) appendNode(data string, typ NodeType) *TNode {
	m := createNode(data, typ)
	m.Parent = n.Parent
	n.NextSibling = m
	m.PrevSibling = n
	if n.Parent != nil {
		n.Parent.LastChild = m
	}
	return m
}

func (n *TNode) addAttribute(k, v string) {
	n.Attr = append(n.Attr, Attribute{k, v})
}

func (n *TNode) getAttribute(key string) string {
	for i := 0; i < len(n.Attr); i++ {
		if n.Attr[i].Key == key {
			return n.Attr[i].Value
		}
	}
	return ""
}
func example2() *TNode {
	/*
		<html lang="en">
		   <head>
			   <title>Hello</title>
			   <meta name="language" content="en"/>
		   </head>
		   <body>
				<h1><span>SPAN</span><a>Anchor</a> This is a H1 </h1>
				<table>
					<tbody>
						<tr>
							<td>row1-val1</td>
							<td>row1-val2</td>
							<td>row1-val3</td>
						</tr>
						<tr>
							<td><para>row2-val1</para></td>
							<td><para>row2-val2</para></td>
							<td><para>row2-val3</para></td>
						</tr>
						<tr>
							<td>row3-val1</td>
							<td><para>row3-val2</para></td>
							<td>row3-val3</td>
						</tr>
					</tbody>
				</table>
		   </body>
		</html>
	*/
	doc := createNode("", RootNode)
	xhtml := doc.createChildNode("html", ElementNode)
	xhtml.addAttribute("lang", "en")

	// The HTML head section.
	head := xhtml.createChildNode("head", ElementNode)
	n := head.createChildNode("title", ElementNode)
	n = n.createChildNode("Hello", TextNode)
	n = head.createChildNode("meta", ElementNode)
	n.addAttribute("name", "language")
	n.addAttribute("content", "en")
	// The HTML body section.
	body := xhtml.createChildNode("body", ElementNode)
	h1 := body.createChildNode("h1", ElementNode)
	n = h1.createChildNode("span", ElementNode)
	n = n.createChildNode("SPAN", TextNode)
	n = h1.createChildNode("a", ElementNode)
	n = n.createChildNode("Anchor", TextNode)
	h1.createChildNode(" This is a H1 ", TextNode)

	n = body.createChildNode("table", ElementNode)
	tbody := n.createChildNode("tbody", ElementNode)
	n = tbody.createChildNode("tr", ElementNode)
	n.createChildNode("td", ElementNode).createChildNode("row1-val1", TextNode)
	n.createChildNode("td", ElementNode).createChildNode("row1-val2", TextNode)
	n.createChildNode("td", ElementNode).createChildNode("row1-val3", TextNode)
	n = tbody.createChildNode("tr", ElementNode)
	n.createChildNode("td", ElementNode).createChildNode("para", ElementNode).createChildNode("row2-val1", TextNode)
	n.createChildNode("td", ElementNode).createChildNode("para", ElementNode).createChildNode("row2-val2", TextNode)
	n.createChildNode("td", ElementNode).createChildNode("para", ElementNode).createChildNode("row2-val3", TextNode)
	n = tbody.createChildNode("tr", ElementNode)
	n.createChildNode("td", ElementNode).createChildNode("row3-val1", TextNode)
	n.createChildNode("td", ElementNode).createChildNode("para", ElementNode).createChildNode("row3-val2", TextNode)
	n.createChildNode("td", ElementNode).createChildNode("row3-val3", TextNode)

	return xhtml
}

func example() *TNode {
	/*
		<html lang="en">
		   <head>
			   <title>Hello</title>
			   <meta name="language" content="en"/>
			   <meta name="description" content="some description"/>
		   </head>
		   <body>
				<h1>
				This is a H1
				</h1>
				<ul>
					<li><a id="1" href="/">Home</a></li>
					<li><a id="2" href="/about">about</a></li>
					<li><a id="3" href="/account">login</a></li>
					<li></li>
				</ul>
				<p>
					Hello,This is an example for gxpath.
				</p>
				<footer>footer script</footer>
		   </body>
		</html>
	*/
	doc := createNode("", RootNode)
	xhtml := doc.createChildNode("html", ElementNode)
	xhtml.addAttribute("lang", "en")

	// The HTML head section.
	head := xhtml.createChildNode("head", ElementNode)
	n := head.createChildNode("title", ElementNode)
	n = n.createChildNode("Hello", TextNode)
	n = head.createChildNode("meta", ElementNode)
	n.addAttribute("name", "language")
	n.addAttribute("content", "en")
	n = head.createChildNode("meta", ElementNode)
	n.addAttribute("name", "description")
	n.addAttribute("content", "some description")
	// The HTML body section.
	body := xhtml.createChildNode("body", ElementNode)
	n = body.createChildNode("h1", ElementNode)
	n = n.createChildNode("\nThis is a H1\n", TextNode)
	ul := body.createChildNode("ul", ElementNode)
	n = ul.createChildNode("li", ElementNode)
	n = n.createChildNode("a", ElementNode)
	n.addAttribute("id", "1")
	n.addAttribute("href", "/")
	n = n.createChildNode("Home", TextNode)
	n = ul.createChildNode("li", ElementNode)
	n = n.createChildNode("a", ElementNode)
	n.addAttribute("id", "2")
	n.addAttribute("href", "/about")
	n = n.createChildNode("about", TextNode)
	n = ul.createChildNode("li", ElementNode)
	n = n.createChildNode("a", ElementNode)
	n.addAttribute("id", "3")
	n.addAttribute("href", "/account")
	n = n.createChildNode("login", TextNode)
	n = ul.createChildNode("li", ElementNode)

	n = body.createChildNode("p", ElementNode)
	n = n.createChildNode("Hello,This is an example for gxpath.", TextNode)

	n = body.createChildNode("footer", ElementNode)
	n = n.createChildNode("footer script", TextNode)

	return xhtml
}
