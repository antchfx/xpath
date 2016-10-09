package gxpath

import (
	"errors"

	"github.com/antchfx/gxpath/internal/build"
	"github.com/antchfx/gxpath/internal/query"
	"github.com/antchfx/gxpath/xpath"
)

// NodeIterator holds all matched Node object.
type NodeIterator struct {
	node  xpath.NodeNavigator
	query query.Query
}

// Current returns current matched node.
func (t *NodeIterator) Current() xpath.NodeNavigator {
	return t.node
}

// MoveNext moving to a next matched node.
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
func Select(root xpath.Node, expr string) (*NodeIterator, error) {
	node, ok := root.(xpath.NodeNavigator)
	if !ok {
		return nil, errors.New("gxpath: Node root does not implement a xpath.NodeNavigator interface")
	}
	qy, err := build.Build(expr)
	if err != nil {
		return nil, err
	}
	return &NodeIterator{query: qy, node: node}, nil
}
