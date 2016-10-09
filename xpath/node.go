package xpath

// A Node is an element node that can navigating to
// an node attribute and another node.
type Node interface{}

// A type of XPath node.
type NodeType int

const (
	// A root node of the XML document or node tree.
	RootNode NodeType = iota

	// An element, such as <element>.
	ElementNode

	// An attribute, such as id='123'.
	AttributeNode

	// A namespace, such as xmlns="namespace".
	NamespaceNode

	// The text content of a node.
	TextNode

	// A comment node, such as <!-- my comment -->
	CommentNode

	// A node with only white space characters and no significant
	// white space. White space characters are #x20, #x9, #xD, or #xA.
	//WhitespaceNode

	// A processing instruction, such as <?pi test?>.
	ProcessingInstructionNode
)
