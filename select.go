package gxpath

import (
	"github.com/antchfx/gxpath/internal/build"
	"github.com/antchfx/gxpath/internal/query"
	"github.com/antchfx/gxpath/xpath"
)

// NodeIterator holds all matched Node object.
type NodeIterator struct {
	node  xpath.NodeNavigator
	query query.Query
}

// Current returns current node which matched.
func (t *NodeIterator) Current() xpath.NodeNavigator {
	return t.node
}

// MoveNext moves Navigator to the next match node.
func (t *NodeIterator) MoveNext() bool {
	n := t.query.Select(t)
	if n != nil {
		if !t.node.MoveTo(n) {
			t.node = n.Copy()
		}
		return true
	}
	return false
}

// Select selects a node set using the specified XPath expression.
func Select(root xpath.NodeNavigator, expr string) *NodeIterator {
	qy, err := build.Build(expr)
	if err != nil {
		panic(err)
	}
	t := &NodeIterator{query: qy, node: root}
	return t
}
