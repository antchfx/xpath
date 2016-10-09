package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/antchfx/gxpath"
	"github.com/antchfx/gxpath/xpath"
	"github.com/antchfx/xml"
)

type xmlNodeNavigator struct {
	doc      *xml.Node
	currnode *xml.Node
	attindex int
}

func (n *xmlNodeNavigator) LocalName() string {
	if n.attindex != -1 && len(n.currnode.Attr) > 0 {
		return n.currnode.Attr[n.attindex].Name.Local
	} else {
		return n.currnode.Data
	}
}

func (n *xmlNodeNavigator) Current() xpath.Node {
	return n.currnode
}

func (n *xmlNodeNavigator) Value() string {
	switch n.currnode.Type {
	case xml.CommentNode:
		return n.currnode.Data
	case xml.DocumentNode:
		return ""
	case xml.ElementNode:
		if n.attindex != -1 {
			return n.currnode.Attr[n.attindex].Value
		}
		return XmlNodeInnerText(n.currnode)
	case xml.TextNode:
		return n.currnode.Data
	default:
		panic(fmt.Sprintf("unknowed XmlNodeType: %v", n.currnode.Type))
	}
}

func (n *xmlNodeNavigator) Prefix() string {
	return ""
}

func (n *xmlNodeNavigator) NodeType() xpath.NodeType {
	switch n.currnode.Type {
	case xml.CommentNode:
		return xpath.CommentNode
	case xml.TextNode:
		return xpath.TextNode
	case xml.DeclarationNode, xml.DocumentNode:
		return xpath.RootNode
	case xml.ElementNode:
		if n.attindex != -1 {
			return xpath.AttributeNode
		}
		return xpath.ElementNode
	default:
		panic(fmt.Sprintf("unknowed XmlNodeType: %v", n.currnode.Type))
	}
}

func (n *xmlNodeNavigator) Copy() xpath.NodeNavigator {
	nav := *n
	return &nav
}

func (n *xmlNodeNavigator) MoveTo(other xpath.NodeNavigator) bool {
	nav, ok := other.(*xmlNodeNavigator)
	if !ok {
		return false
	}
	if nav.doc == n.doc {
		n.currnode = nav.currnode
		n.attindex = nav.attindex
		return true
	}
	return false
}

func (n *xmlNodeNavigator) MoveToRoot() {
	n.currnode = n.doc
}

func (n *xmlNodeNavigator) MoveToPrevious() bool {
	if n.currnode.PrevSibling == nil {
		return false
	}
	n.currnode = n.currnode.PrevSibling
	return true
}

func (n *xmlNodeNavigator) MoveToParent() bool {
	if n.currnode.Parent == nil {
		return false
	}
	n.currnode = n.currnode.Parent
	return true
}

func (n *xmlNodeNavigator) MoveToFirst() bool {
	if n.currnode.PrevSibling == nil {
		return false
	}
	for n.currnode.PrevSibling != nil {
		n.currnode = n.currnode.PrevSibling
	}
	return true
}

func (n *xmlNodeNavigator) String() string {
	return n.Value()
}

func (n *xmlNodeNavigator) MoveToNext() bool {
	if cur := n.currnode.NextSibling; cur == nil {
		return false
	} else {
		n.currnode = cur
	}
	return true
}

func (n *xmlNodeNavigator) MoveToFirstAttribute() bool {
	if len(n.currnode.Attr) == 0 {
		return false
	}
	n.attindex = 0
	return true
}

func (n *xmlNodeNavigator) MoveToNextAttribute() bool {
	if n.attindex >= len(n.currnode.Attr)-1 {
		return false
	}
	n.attindex++
	return true
}

func (n *xmlNodeNavigator) MoveToChild() bool {
	if cur := n.currnode.FirstChild; cur == nil {
		return false
	} else {
		n.currnode = cur
	}
	return true
}

func XmlNodeInnerText(n *xml.Node) string {
	var b bytes.Buffer
	var output func(*xml.Node)
	output = func(node *xml.Node) {
		if node.Type == xml.TextNode {
			b.WriteString(node.Data)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			output(child)
		}
	}
	output(n)
	return b.String()
}

func main() {
	s := `<?xml version="1.0" encoding="UTF-8"?>
<bookstore>
<book>
  <title lang="en" id="1">Harry Potter</title>
  <price>29.99</price>
</book>
<book>
  <title lang="en" id="2">Learning XML</title>
  <price>39.95</price>
</book>
</bookstore>
`
	root, err := xml.Parse(strings.NewReader(s))
	if err != nil {
		panic(err)
	}
	nav := &xmlNodeNavigator{doc: root, currnode: root, attindex: -1}
	fmt.Println("enter XPath express and press enter key to test.")
	scan := bufio.NewScanner(os.Stdin)
	for {
		scan.Scan()
		iter, err := gxpath.Select(nav, strings.TrimSpace(scan.Text()))
		if err != nil {
			fmt.Println(fmt.Sprintf("got error: %v", err))
		} else {
			for iter.MoveNext() {
				fmt.Println(">>")
				fmt.Println(iter.Current().Value())
			}
		}
		fmt.Println("==========")
	}
}
