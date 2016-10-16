package xpath

// A Node is an element node that can navigating to
// an node attribute and another node.
//type Node interface{}

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
