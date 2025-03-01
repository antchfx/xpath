package xpath_test

import (
	"bytes"
	"fmt"

	"github.com/antchfx/xpath"
)

type NodeType uint

const (
	DocumentNode NodeType = iota
	ElementNode
	TextNode
)

type Node struct {
	Parent, FirstChild, LastChild, PrevSibling, NextSibling *Node

	Type NodeType
	Data string
}

func (n *Node) Value() string {
	if n.Type == TextNode {
		return n.Data
	}

	var buff bytes.Buffer
	var output func(*Node)
	output = func(node *Node) {
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

func (parent *Node) AddChild(n *Node) {
	n.Parent = parent
	n.NextSibling = nil
	if parent.FirstChild == nil {
		parent.FirstChild = n
		n.PrevSibling = nil
	} else {
		parent.LastChild.NextSibling = n
		n.PrevSibling = parent.LastChild
	}

	parent.LastChild = n
}

type NodeNavigator struct {
	curr, root *Node
}

func (n *NodeNavigator) NodeType() xpath.NodeType {
	switch n.curr.Type {
	case TextNode:
		return xpath.TextNode
	case DocumentNode:
		return xpath.RootNode
	}
	return xpath.ElementNode
}

func (n *NodeNavigator) LocalName() string {
	return n.curr.Data
}

func (n *NodeNavigator) Prefix() string {
	return ""
}

func (n *NodeNavigator) NamespaceURL() string {
	return ""
}

func (n *NodeNavigator) Value() string {
	switch n.curr.Type {
	case ElementNode:
		return n.curr.Value()
	case TextNode:
		return n.curr.Data
	}
	return ""
}

func (n *NodeNavigator) Copy() xpath.NodeNavigator {
	n2 := *n
	return &n2
}

func (n *NodeNavigator) MoveToRoot() {
	n.curr = n.root
}

func (n *NodeNavigator) MoveToParent() bool {
	if node := n.curr.Parent; node != nil {
		n.curr = node
		return true
	}
	return false
}

func (n *NodeNavigator) MoveToNextAttribute() bool {
	return true
}

func (n *NodeNavigator) MoveToChild() bool {
	if node := n.curr.FirstChild; node != nil {
		n.curr = node
		return true
	}
	return false
}

func (n *NodeNavigator) MoveToFirst() bool {
	for {
		node := n.curr.PrevSibling
		if node == nil {
			break
		}
		n.curr = node
	}
	return true
}

func (n *NodeNavigator) String() string {
	return n.Value()
}

func (n *NodeNavigator) MoveToNext() bool {
	if node := n.curr.NextSibling; node != nil {
		n.curr = node
		return true
	}
	return false
}

func (n *NodeNavigator) MoveToPrevious() bool {
	if node := n.curr.PrevSibling; node != nil {
		n.curr = node
		return true
	}
	return false
}

func (n *NodeNavigator) MoveTo(other xpath.NodeNavigator) bool {
	node, ok := other.(*NodeNavigator)
	if !ok || node.root != n.root {
		return false
	}

	n.curr = node.curr
	return true
}

// XPath package example. See more xpath implements package:
// https://github.com/antchfx/htmlquery
// https://github.com/antchfx/xmlquery
// https://github.com/antchfx/jsonquery
func Example() {
	/***
	?xml version="1.0" encoding="UTF-8"?>
	<bookstore>
		<book>
			<title>Everyday Italian</title>
			<author>Giada De Laurentiis</author>
			<year>2005</year>
			<price>30.00</price>
		</book>
		<book>
			<title>Harry Potter</title>
			<author>J K. Rowling</author>
			<year>2005</year>
			<price>29.99</price>
		</book>
	</bookstore>
	**/

	// Here, for begin test, we should create a document
	books := []struct {
		title  string
		author string
		year   int
		price  float64
	}{
		{title: "Everyday Italian", author: "Giada De Laurentiis", year: 2005, price: 30.00},
		{title: "Harry Potter", author: "J K. Rowling", year: 2005, price: 29.99},
	}
	bookstore := &Node{Data: "bookstore", Type: ElementNode}
	for _, v := range books {
		book := &Node{Data: "book", Type: ElementNode}
		title := &Node{Data: "title", Type: ElementNode}
		title.AddChild(&Node{Data: v.title, Type: TextNode})
		book.AddChild(title)
		author := &Node{Data: "author", Type: ElementNode}
		author.AddChild(&Node{Data: v.author, Type: TextNode})
		book.AddChild(author)
		year := &Node{Data: "year", Type: ElementNode}
		year.AddChild(&Node{Data: fmt.Sprintf("%d", v.year), Type: TextNode})
		book.AddChild(year)
		price := &Node{Data: "price", Type: ElementNode}
		price.AddChild(&Node{Data: fmt.Sprintf("%f", v.price), Type: TextNode})
		book.AddChild(price)
		bookstore.AddChild(book)
	}
	var doc = &Node{}
	doc.AddChild(bookstore)
	var root xpath.NodeNavigator = &NodeNavigator{curr: doc, root: doc}
	expr, err := xpath.Compile("count(//book)")
	// using Evaluate() method
	if err != nil {
		panic(err)
	}
	val := expr.Evaluate(root) // it returns float64 type
	fmt.Println(val.(float64))

	// using Evaluate() method
	expr = xpath.MustCompile("sum(//price)")
	val = expr.Evaluate(root) // output total price
	fmt.Println(val.(float64))

	// using Select() method
	expr = xpath.MustCompile("//book")
	iter := expr.Select(root) // it always returns NodeIterator object.
	for iter.MoveNext() {
		fmt.Println(iter.Current().Value())
	}
}
