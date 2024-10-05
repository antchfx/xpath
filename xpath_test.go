package xpath

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"strings"
	"testing"
)

var (
	employee_example = createEmployeeExample()
	book_example     = createBookExample()
	html_example     = createHtmlExample()
	empty_example    = createNode("", RootNode)
	mybook_example   = createMyBookExample()
)

type testQuery string

func (t testQuery) Select(_ iterator) NodeNavigator {
	panic("implement me")
}

func (t testQuery) Clone() query {
	return t
}

func (t testQuery) Evaluate(_ iterator) interface{} {
	return string(t)
}

func (t testQuery) ValueType() resultType {
	return xpathResultType.Any
}

func (t testQuery) Properties() queryProp {
	return queryProps.None
}

func test_xpath_elements(t *testing.T, root *TNode, expr string, expected ...int) {
	result := selectNodes(root, expr)
	assertEqual(t, len(expected), len(result))

	for i := 0; i < len(expected); i++ {
		assertEqual(t, expected[i], result[i].lines)
	}
}

func test_xpath_values(t testing.TB, root *TNode, expr string, expected ...string) {
	result := selectNodes(root, expr)
	assertEqual(t, len(expected), len(result))

	for i := 0; i < len(expected); i++ {
		assertEqual(t, expected[i], result[i].Value())
	}
}

func test_xpath_tags(t *testing.T, root *TNode, expr string, expected ...string) {
	result := selectNodes(root, expr)
	assertEqual(t, len(expected), len(result))

	for i := 0; i < len(expected); i++ {
		assertEqual(t, expected[i], result[i].Data)
	}
}

func test_xpath_count(t *testing.T, root *TNode, expr string, expected int) {
	result := selectNodes(root, expr)
	assertEqual(t, expected, len(result))
}

func test_xpath_eval(t *testing.T, root *TNode, expr string, expected ...interface{}) {
	e, err := Compile(expr)
	assertNoErr(t, err)

	v := e.Evaluate(createNavigator(root))
	// if is a node-set
	if iter, ok := v.(*NodeIterator); ok {
		got := iterateNavs(iter)
		assertEqual(t, len(expected), len(got))
		for i := 0; i < len(expected); i++ {
			assertEqual(t, expected[i], got[i])
		}
		return
	}
	assertEqual(t, expected[0], v)
}

func Test_Predicates_MultiParent(t *testing.T) {
	// https://github.com/antchfx/xpath/issues/75
	/*
	   <measCollecFile xmlns="http://www.3gpp.org/ftp/specs/archive/32_series/32.435#measCollec">
	   		<measData>
	   			<measInfo>
	   				<measType p="1">field1</measType>
	   				<measType p="2">field2</measType>
	   				<measValue>
	   					<r p="1">31854</r>
	   					<r p="2">159773</r>
	   				</measValue>
	   			</measInfo>
	   			<measInfo measInfoId="metric_name2">
	   				<measType p="1">field3</measType>
	   				<measType p="2">field4</measType>
	   				<measValue>
	   					<r p="1">1234</r>
	   					<r p="2">567</r>
	   				</measValue>
	   			</measInfo>
	   		</measData>
	   	</measCollecFile>
	*/
	doc := createNode("", RootNode)
	measCollecFile := doc.createChildNode("measCollecFile", ElementNode)
	measData := measCollecFile.createChildNode("measData", ElementNode)
	data := []struct {
		measType  map[string]string
		measValue map[string]string
	}{
		{measType: map[string]string{"1": "field1", "2": "field2"}, measValue: map[string]string{"1": "31854", "2": "159773"}},
		{measType: map[string]string{"1": "field3", "2": "field4"}, measValue: map[string]string{"1": "1234", "2": "567"}},
	}
	for j := 0; j < len(data); j++ {
		d := data[j]
		measInfo := measData.createChildNode("measInfo", ElementNode)
		measType := measInfo.createChildNode("measType", ElementNode)

		var keys []string
		for k := range d.measType {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			measType.addAttribute("p", k)
			measType.createChildNode(d.measType[k], TextNode)
		}

		measValue := measInfo.createChildNode("measValue", ElementNode)
		keys = make([]string, 0)
		for k := range d.measValue {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			r := measValue.createChildNode("r", ElementNode)
			r.addAttribute("p", k)
			r.createChildNode(d.measValue[k], TextNode)
		}
	}
	test_xpath_values(t, doc, `//r[@p=../../measType/@p]`, "31854", "159773", "1234", "567")

}

func TestCompile(t *testing.T) {
	var err error
	_, err = Compile("//a")
	assertNil(t, err)
	_, err = Compile("//a[id=']/span")
	assertErr(t, err)
	_, err = Compile("//ul/li/@class")
	assertNil(t, err)
	_, err = Compile("/a/b/(c, .[not(c)])")
	assertNil(t, err)
	_, err = Compile("/pre:foo")
	assertNil(t, err)
}

func TestInvalidXPath(t *testing.T) {
	var err error
	_, err = Compile("()")
	assertErr(t, err)
	_, err = Compile("(1,2,3)")
	assertErr(t, err)
}

func TestCompileWithNS(t *testing.T) {
	_, err := CompileWithNS("/foo", nil)
	assertNil(t, err)
	_, err = CompileWithNS("/foo", map[string]string{})
	assertNil(t, err)
	_, err = CompileWithNS("/foo", map[string]string{"a": "b"})
	assertNil(t, err)
	_, err = CompileWithNS("/a:foo", map[string]string{"a": "b"})
	assertNil(t, err)
	_, err = CompileWithNS("/u:foo", map[string]string{"a": "b"})
	assertErr(t, err)
}

func TestNamespacePrefixQuery(t *testing.T) {
	/*
		<?xml version="1.0" encoding="UTF-8"?>
		<books>
			<book>book1</book>
			<b:book xmlns:b="ns">book2</b:book>
			<c:book xmlns:c="ns">book3</c:book>
		</books>
	*/
	doc := createNode("", RootNode)
	books := doc.createChildNode("books", ElementNode)
	books.lines = 2
	book1 := books.createChildNode("book", ElementNode)
	book1.lines = 3
	book1.createChildNode("book1", TextNode)
	book2 := books.createChildNode("b:book", ElementNode)
	book2.lines = 4
	book2.addAttribute("xmlns:b", "ns")
	book2.createChildNode("book2", TextNode)
	book3 := books.createChildNode("c:book", ElementNode)
	book3.lines = 5
	book3.addAttribute("xmlns:c", "ns")
	book3.createChildNode("book3", TextNode)

	test_xpath_elements(t, doc, `//b:book`, 4) // expected [4 , 5]

	// With namespace bindings:
	exp, _ := CompileWithNS("//x:book", map[string]string{"x": "ns"})
	nodes := iterateNodes(exp.Select(createNavigator(doc)))
	assertEqual(t, 2, len(nodes))
	assertEqual(t, "book2", nodes[0].Value())
	assertEqual(t, "book3", nodes[1].Value())
}

func TestMustCompile(t *testing.T) {
	expr := MustCompile("//")
	assertTrue(t, expr != nil)

	if wanted := (nopQuery{}); expr.q != wanted {
		t.Fatalf("wanted nopQuery object but got %s", expr)
	}
	iter := expr.Select(createNavigator(html_example))
	if iter.MoveNext() {
		t.Fatal("should be an empty node list but got one")
	}
}

func Test_plusFunc(t *testing.T) {
	// 1+1
	assertEqual(t, float64(2), plusFunc(nil, float64(1), float64(1)))
	// string +
	assertEqual(t, float64(2), plusFunc(nil, "1", "1"))
	// invalid string
	v := plusFunc(nil, "a", 1)
	assertTrue(t, math.IsNaN(v.(float64)))
	// Nodeset
	// TODO
}

func Test_minusFunc(t *testing.T) {
	// 1 - 1
	assertEqual(t, float64(0), minusFunc(nil, float64(1), float64(1)))
	// string
	assertEqual(t, float64(0), minusFunc(nil, "1", "1"))
	// invalid string
	v := minusFunc(nil, "a", 1)
	assertTrue(t, math.IsNaN(v.(float64)))
}

func TestNodeType(t *testing.T) {
	tests := []struct {
		expr     string
		expected NodeType
	}{
		{`//employee`, ElementNode},
		{`//name[text()]`, ElementNode},
		{`//name/text()`, TextNode},
		{`//employee/@id`, AttributeNode},
	}
	for _, test := range tests {
		v := selectNode(employee_example, test.expr)
		assertTrue(t, v != nil)
		assertEqual(t, test.expected, v.Type)
	}

	doc := createNode("", RootNode)
	doc.createChildNode("<!-- This is a comment -->", CommentNode)
	n := selectNode(doc, "//comment()")
	assertTrue(t, n != nil)
	assertEqual(t, CommentNode, n.Type)
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
		n := t.Current().(*TNodeNavigator)
		if n.NodeType() == AttributeNode {
			childNode := &TNode{
				Type: TextNode,
				Data: n.Value(),
			}
			nodes = append(nodes, &TNode{
				Parent:     n.curr,
				Type:       AttributeNode,
				Data:       n.LocalName(),
				FirstChild: childNode,
				LastChild:  childNode,
			})
		} else {
			nodes = append(nodes, n.curr)
		}

	}
	return nodes
}

func selectNode(root *TNode, expr string) *TNode {
	list := selectNodes(root, expr)
	if len(list) == 0 {
		return nil
	}
	return list[0]
}

func selectNodes(root *TNode, expr string) []*TNode {
	t := Select(createNavigator(root), expr)
	c := make(map[uint64]bool)
	var list []*TNode
	for _, n := range iterateNodes(t) {
		m := getHashCode(createNavigator(n))
		if _, ok := c[m]; ok {
			continue
		}
		c[m] = true
		list = append(list, n)
	}
	return list
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

	Type         NodeType
	Data         string
	Attr         []Attribute
	NamespaceURL string
	Prefix       string
	lines        int
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
	switch n.curr.Type {
	case CommentNode:
		return CommentNode
	case TextNode:
		return TextNode
	case ElementNode:
		if n.attr != -1 {
			return AttributeNode
		}
		return ElementNode
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
	return n.curr.Prefix
}

func (n *TNodeNavigator) NamespaceURL() string {
	if n.Prefix() != "" {
		for _, a := range n.curr.Attr {
			if a.Key == "xmlns:"+n.Prefix() {
				return a.Value
			}
		}
	}
	return n.curr.NamespaceURL
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
	if n.attr != -1 {
		n.attr = -1
		return true
	} else if node := n.curr.Parent; node != nil {
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
	if n.attr != -1 {
		return false
	}
	if node := n.curr.FirstChild; node != nil {
		n.curr = node
		return true
	}
	return false
}

func (n *TNodeNavigator) MoveToFirst() bool {
	if n.attr != -1 || n.curr.PrevSibling == nil {
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
	if n.attr != -1 {
		return false
	}
	if node := n.curr.NextSibling; node != nil {
		n.curr = node
		return true
	}
	return false
}

func (n *TNodeNavigator) MoveToPrevious() bool {
	if n.attr != -1 {
		return false
	}
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

func createBookExample() *TNode {
	/*
	   <?xml version="1.0" encoding="UTF-8"?>
	   <bookstore>
	   <book category="cooking">
	     <title lang="en">Everyday Italian</title>
	     <author>Giada De Laurentiis</author>
	     <year>2005</year>
	     <price>30.00</price>
	   </book>
	   <book category="children">
	     <title lang="en">Harry Potter</title>
	     <author>J K. Rowling</author>
	     <year>2005</year>
	     <price>29.99</price>
	   </book>
	   <book category="web">
	     <title lang="en">XQuery Kick Start</title>
	     <author>James McGovern</author>
	     <author>Per Bothner</author>
	     <author>Kurt Cagle</author>
	     <author>James Linn</author>
	     <author>Vaidyanathan Nagarajan</author>
	     <year>2003</year>
	     <price>49.99</price>
	   </book>
	   <book category="web">
	     <title lang="en">Learning XML</title>
	     <author>Erik T. Ray</author>
	     <year>2003</year>
	     <price>39.95</price>
	   </book>
	   </bookstore>
	*/
	type Element struct {
		Data       string
		Attributes map[string]string
	}
	books := []struct {
		category string
		title    Element
		year     int
		price    float64
		authors  []string
	}{
		{
			category: "cooking",
			title:    Element{"Everyday Italian", map[string]string{"lang": "en"}},
			year:     2005,
			price:    30.00,
			authors:  []string{"Giada De Laurentiis"},
		},
		{
			category: "children",
			title:    Element{"Harry Potter", map[string]string{"lang": "en"}},
			year:     2005,
			price:    29.99,
			authors:  []string{"J K. Rowling"},
		},
		{
			category: "web",
			title:    Element{"XQuery Kick Start", map[string]string{"lang": "en"}},
			year:     2003,
			price:    49.99,
			authors:  []string{"James McGovern", "Per Bothner", "Kurt Cagle", "James Linn", "Vaidyanathan Nagarajan"},
		},
		{
			category: "web",
			title:    Element{"Learning XML", map[string]string{"lang": "en"}},
			year:     2003,
			price:    39.95,
			authors:  []string{"Erik T. Ray"},
		},
	}

	var lines = 0
	doc := createNode("", RootNode)
	lines++
	bookstore := doc.createChildNode("bookstore", ElementNode)
	lines++
	bookstore.lines = lines
	for i := 0; i < len(books); i++ {
		v := books[i]
		lines++

		book := bookstore.createChildNode("book", ElementNode)
		book.lines = lines
		lines++
		book.addAttribute("category", v.category)
		// title
		title := book.createChildNode("title", ElementNode)
		title.lines = lines
		lines++
		for k, v := range v.title.Attributes {
			title.addAttribute(k, v)
		}
		title.createChildNode(v.title.Data, TextNode)
		// authors
		for j := 0; j < len(v.authors); j++ {
			author := book.createChildNode("author", ElementNode)
			author.lines = lines
			lines++
			author.createChildNode(v.authors[j], TextNode)
		}
		// year
		year := book.createChildNode("year", ElementNode)
		year.lines = lines
		lines++
		year.createChildNode(fmt.Sprintf("%d", v.year), TextNode)
		// price
		price := book.createChildNode("price", ElementNode)
		price.lines = lines
		lines++
		price.createChildNode(fmt.Sprintf("%.2f", v.price), TextNode)
	}
	return doc
}

// The example document from https://way2tutorial.com/xml/xpath-node-test.php
func createEmployeeExample() *TNode {
	/*
	   <?xml version="1.0" standalone="yes"?>
	   <empinfo>
	     <employee id="1">
	       <name>Opal Kole</name>
	       <designation discipline="web" experience="3 year">Senior Engineer</designation>
	       <email>OpalKole@myemail.com</email>
	     </employee>
	     <employee id="2">
	       <name from="CA">Max Miller</name>
	       <designation discipline="DBA" experience="2 year">DBA Engineer</designation>
	       <email>maxmiller@email.com</email>
	     </employee>
	     <employee id="3">
	       <name>Beccaa Moss</name>
	       <designation discipline="appdev">Application Developer</designation>
	       <email>beccaamoss@email.com</email>
	     </employee>
	   </empinfo>
	*/

	type Element struct {
		Data       string
		Attributes map[string]string
	}
	var lines = 0
	doc := createNode("", RootNode)
	lines++ // 1
	empinfo := doc.createChildNode("empinfo", ElementNode)
	lines++
	empinfo.lines = lines
	var employees = []struct {
		name        Element
		designation Element
		email       Element
	}{
		{
			name: Element{Data: "Opal Kole"},
			designation: Element{Data: "Senior Engineer", Attributes: map[string]string{
				"discipline": "web",
				"experience": "3 year",
			}},
			email: Element{Data: "OpalKole@myemail.com"},
		},
		{
			name: Element{Data: "Max Miller", Attributes: map[string]string{"from": "CA"}},
			designation: Element{Data: "DBA Engineer", Attributes: map[string]string{
				"discipline": "DBA",
				"experience": "2 year",
			}},
			email: Element{Data: "maxmiller@email.com"},
		},
		{
			name: Element{Data: "Beccaa Moss"},
			designation: Element{Data: "Application Developer", Attributes: map[string]string{
				"discipline": "appdev",
			}},
			email: Element{Data: "beccaamoss@email.com"},
		},
	}
	for i := 0; i < len(employees); i++ {
		v := employees[i]
		lines++
		// employee
		employee := empinfo.createChildNode("employee", ElementNode)
		employee.addAttribute("id", fmt.Sprintf("%d", i+1))
		employee.lines = lines
		lines++
		// name
		name := employee.createChildNode("name", ElementNode)
		name.createChildNode(v.name.Data, TextNode)
		for k, n := range v.name.Attributes {
			name.addAttribute(k, n)
		}
		name.lines = lines
		lines++
		// designation
		designation := employee.createChildNode("designation", ElementNode)
		designation.createChildNode(v.designation.Data, TextNode)
		for k, n := range v.designation.Attributes {
			designation.addAttribute(k, n)
		}
		designation.lines = lines
		lines++
		// email
		email := employee.createChildNode("email", ElementNode)
		email.createChildNode(v.email.Data, TextNode)
		for k, n := range v.email.Attributes {
			email.addAttribute(k, n)
		}
		email.lines = lines
		// skiping closed tag
		lines++
	}
	return doc
}

func createHtmlExample() *TNode {
	/*
		<html lang="en">
		  <head>
		    <title>My page</title>
		    <meta name="language" content="en" />
		  </head>
		  <body>
		    <h2>Welcome to my page</h2>
		    <ul>
		      <li>
		        <a href="/">Home</a>
		      </li>
		      <li>
		        <a href="/about">About</a>
		      </li>
		      <li>
		        <a href="/account">Login</a>
		      </li>
			  <li></li>
		    </ul>
		    <p>This is the first paragraph.</p>
		    <!-- this is the end -->
		  </body>
		</html>
	*/
	lines := 0
	doc := createNode("", RootNode)
	lines++
	xhtml := doc.createChildNode("html", ElementNode)
	xhtml.lines = lines
	xhtml.addAttribute("lang", "en")
	lines++
	// head container
	head := xhtml.createChildNode("head", ElementNode)
	head.lines = lines
	lines++
	title := head.createChildNode("title", ElementNode)
	title.lines = lines
	title.createChildNode("My page", TextNode)
	lines++
	meta := head.createChildNode("meta", ElementNode)
	meta.lines = lines
	meta.addAttribute("name", "language")
	meta.addAttribute("content", "en")
	// skip the head
	lines++
	lines++
	body := xhtml.createChildNode("body", ElementNode)
	body.lines = lines
	lines++
	h2 := body.createChildNode("h2", ElementNode)
	h2.lines = lines
	h2.createChildNode("Welcome to my page", TextNode)
	lines++
	links := []struct {
		text string
		href string
	}{
		{text: "Home", href: "/"},
		{text: "About", href: "/About"},
		{text: "Login", href: "/account"},
	}

	ul := body.createChildNode("ul", ElementNode)
	ul.lines = lines
	lines++
	for i := 0; i < len(links); i++ {
		link := links[i]
		li := ul.createChildNode("li", ElementNode)
		li.lines = lines
		lines++
		a := li.createChildNode("a", ElementNode)
		a.lines = lines
		a.addAttribute("href", link.href)
		a.createChildNode(link.text, TextNode)
		lines++
		// skip the <li>
		lines++
	}
	// skip the last ul
	lines++
	p := body.createChildNode("p", ElementNode)
	p.lines = lines
	lines++
	p.createChildNode("This is the first paragraph.", TextNode)
	lines++
	comment := body.createChildNode("<!-- this is the end -->", CommentNode)
	comment.lines = lines
	lines++
	return doc
}

func createMyBookExample() *TNode {
	/*
		<?xml version="1.0" encoding="utf-8"?>
		<books xmlns:mybook="http://www.contoso.com/books">
			<mybook:book id="bk101">
				<title>XML Developer's Guide</title>
				<author>Gambardella, Matthew</author>
				<price>44.95</price>
				<publish_date>2000-10-01</publish_date>
			</mybook:book>
			<mybook:book id="bk102">
				<title>Midnight Rain</title>
				<author>Ralls, Kim</author>
				<price>5.95</price>
				<publish_date>2000-12-16</publish_date>
			</mybook:book>
		</books>
	*/
	var (
		prefix       string = "mybook"
		namespaceURL string = "http://www.contoso.com/books"
	)
	lines := 1
	doc := createNode("", RootNode)
	doc.lines = lines
	lines++
	books := doc.createChildNode("books", ElementNode)
	books.addAttribute("xmlns:mybook", namespaceURL)
	books.lines = lines
	lines++
	data := []struct {
		id      string
		title   string
		author  string
		price   float64
		publish string
	}{
		{id: "bk101", title: "XML Developer's Guide", author: "Gambardella, Matthew", price: 44.95, publish: "2000-10-01"},
		{id: "bk102", title: "Midnight Rain", author: "Ralls, Kim", price: 5.95, publish: "2000-12-16"},
	}
	for i := 0; i < len(data); i++ {
		v := data[i]
		book := books.createChildNode("book", ElementNode)
		book.addAttribute("id", v.id)
		book.Prefix = prefix
		book.NamespaceURL = namespaceURL
		book.lines = lines
		lines++
		title := book.createChildNode("title", ElementNode)
		title.createChildNode(v.title, TextNode)
		title.lines = lines
		lines++
		author := book.createChildNode("author", ElementNode)
		author.createChildNode(v.author, TextNode)
		author.lines = lines
		lines++
		price := book.createChildNode("price", ElementNode)
		price.createChildNode(fmt.Sprintf("%.2f", v.price), TextNode)
		price.lines = lines
		lines++
		publish_date := book.createChildNode("publish_date", ElementNode)
		publish_date.createChildNode(v.publish, TextNode)
		publish_date.lines = lines
		lines++
		// skip the last of book element
		lines++
	}
	return doc
}
