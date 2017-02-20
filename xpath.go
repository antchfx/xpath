package xpath

// A type of XPath node.
type NodeType int

const (
	// A root node of the XML document or node tree.
	RootNode NodeType = iota

	// An element, such as <element>.
	ElementNode

	// An attribute, such as id='123'.
	AttributeNode

	// The text content of a node.
	TextNode

	// A comment node, such as <!-- my comment -->
	CommentNode
)

// NodeNavigator provides cursor model for navigating XML data.
type NodeNavigator interface {
	// NodeType returns the XPathNodeType of the current node.
	NodeType() NodeType

	// LocalName gets the Name of the current node.
	LocalName() string

	// Prefix returns namespace prefix associated with the current node.
	Prefix() string

	// Value gets the value of current node.
	Value() string

	// Copy does a deep copy of the NodeNavigator and all its components.
	Copy() NodeNavigator

	// MoveToRoot moves the NodeNavigator to the root node of the current node.
	MoveToRoot()

	// MoveToParent moves the NodeNavigator to the parent node of the current node.
	MoveToParent() bool

	// MoveToNextAttribute moves the NodeNavigator to the next attribute on current node.
	MoveToNextAttribute() bool

	// MoveToChild moves the NodeNavigator to the first child node of the current node.
	MoveToChild() bool

	// MoveToFirst moves the NodeNavigator to the first sibling node of the current node.
	MoveToFirst() bool

	// MoveToNext moves the NodeNavigator to the next sibling node of the current node.
	MoveToNext() bool

	// MoveToPrevious moves the NodeNavigator to the previous sibling node of the current node.
	MoveToPrevious() bool

	// MoveTo moves the NodeNavigator to the same position as the specified NodeNavigator.
	MoveTo(NodeNavigator) bool
}

// NodeIterator holds all matched Node object.
type NodeIterator struct {
	node  NodeNavigator
	query query
}

// Current returns current node which matched.
func (t *NodeIterator) Current() NodeNavigator {
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
func Select(root NodeNavigator, expr string) *NodeIterator {
	qy, err := build(expr)
	if err != nil {
		panic(err)
	}
	t := &NodeIterator{query: qy, node: root}
	return t
}
