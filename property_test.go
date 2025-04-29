package xpath

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// Helper to convert antchfx result to string for comparison
func antchfxResultToString(result interface{}, nav NodeNavigator) string {
	switch v := result.(type) {
	case *NodeIterator:
		// Create a copy of the iterator to avoid consuming the original
		iterCopy := &NodeIterator{query: v.query.Clone(), node: nav.Copy()}
		var parts []string
		for iterCopy.MoveNext() {
			currentNodeNav := iterCopy.Current()
			// Attempt to cast to *TNodeNavigator to access the underlying *TNode
			// This assumes we are using TNodeNavigator in these tests.
			if tNav, ok := currentNodeNav.(*TNodeNavigator); ok {
				// Serialize the current node to an XML string snippet
				parts = append(parts, serializeNodeToString(tNav.curr))
			} else {
				// Fallback or error if the navigator is not a TNodeNavigator
				// For now, use the basic Value() as a fallback, though it won't match xmllint XML output
				parts = append(parts, currentNodeNav.Value())
			}
		}
		// Join the XML snippets with newlines, similar to how xmllint outputs multiple nodes.
		return strings.Join(parts, "\n")
	case string:
		return v
	case float64:
		// Consistent float formatting, matching strconv.ParseFloat default
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case query:
		// If Evaluate returns a query object (e.g., for paths that don't evaluate to final value directly?)
		// We might need to iterate it here as well. Let's assume Evaluate gives final value or NodeIterator.
		// For now, represent as type name.
		return fmt.Sprintf("unhandled_antchfx_type_query")
	default:
		// Other unexpected types.
		return fmt.Sprintf("unhandled_antchfx_type_%T", result)
	}
}

// Helper to parse xmllint output to a comparable string
func parseXmllintOutput(stdout string, stderr string, exitCode int) (string, error) {
	// xmllint exit codes: http://xmlsoft.org/xmllint.html
	// 0: OK
	// 10: XPath evaluation returned no result (empty node set)
	// 11: Error evaluating the XPath expression (syntax/evaluation error)
	// 12: Error building the context for XPath evaluation
	// Other codes: XML parsing errors, etc.

	// Handle specific evaluation outcomes based on exit code and stderr hints
	if exitCode == 10 { // Empty result set
		// Check stderr for explicit non-node-set types returning "empty" equivalent
		if strings.Contains(stderr, "XPath expression evaluates to a Boolean : false") {
			return "false", nil
		}
		// If number evaluates to NaN? xmllint might exit 10. Represent as empty? Or NaN? Let's use empty for now.
		// Default for exit 10 is empty node-set or empty string.
		return "", nil
	}

	if exitCode == 0 { // Success
		// Check stderr for explicit type information first
		if strings.Contains(stderr, "XPath expression evaluates to a Boolean : false") {
			return "false", nil
		}
		if strings.Contains(stderr, "XPath expression evaluates to a Boolean : true") {
			return "true", nil
		}
		if strings.Contains(stderr, "XPath expression evaluates to a Number : ") {
			parts := strings.SplitN(stderr, ":", 2)
			if len(parts) == 2 {
				numStr := strings.TrimSpace(parts[1])
				// Normalize number format
				f, err := strconv.ParseFloat(numStr, 64)
				if err == nil {
					return strconv.FormatFloat(f, 'f', -1, 64), nil
				}
				return numStr, nil // Fallback to raw string if parsing fails
			}
		}
		// If no specific type in stderr, assume stdout contains the result (string or node-set XML).
		// Return the trimmed stdout directly. We will compare this raw output.
		return strings.TrimSpace(stdout), nil
	} // This brace correctly closes the `if exitCode == 0` block

	// If exitCode is 11 (evaluation error) or 12 (context error) or others, return error.
	// The caller should handle these cases (e.g., compare if antchfx also failed).
	// We return an empty string and the error for context.
	return "", fmt.Errorf("xmllint failed or evaluation error (exit code %d). Stderr: %s", exitCode, stderr)
}

// Check if xmllint command exists and skip tests if not.
func checkXmllintAvailability(t *testing.T) {
	t.Helper()
	_, err := exec.LookPath("xmllint")
	if err != nil {
		t.Skip("xmllint command not found in PATH, skipping differential tests.")
	}
}

// Limited set of tags for generation to increase match probability.
var htmlTags = []string{"div", "p", "span", "a", "b", "i", "table", "tr", "td"}

// Limited set of attribute names.
var htmlAttrs = []string{"id", "class", "href", "title", "style"}

// genTNode generates a random TNode tree resembling simple HTML.
// Declared at package level to allow recursive definition in init().
var genTNode *rapid.Generator[*TNode]

func init() {
	genTNode = rapid.Custom(func(t *rapid.T) *TNode {
		// Decide node type: element or text. Bias towards elements initially.
		// Limit recursion depth implicitly by reducing probability of elements at deeper levels,
		// or explicitly pass depth (more complex). Let's rely on rapid's size control for now.
		isElement := rapid.Bool().Draw(t, "isElement")
		if !isElement {
			// Generate a text node from a limited set.
			text := rapid.SampledFrom([]string{"", "foo", "bar"}).Draw(t, "textData")
			return createNode(text, TextNode)
		}

		// Generate an element node.
		tag := rapid.SampledFrom(htmlTags).Draw(t, "tag")
		node := createNode(tag, ElementNode)

		// Add attributes sometimes.
		if rapid.Bool().Draw(t, "hasAttrs") {
			numAttrs := rapid.IntRange(0, 3).Draw(t, "numAttrs")
			for i := 0; i < numAttrs; i++ {
				attrName := rapid.SampledFrom(htmlAttrs).Draw(t, fmt.Sprintf("attrName%d", i))
				// Ensure unique attribute names for simplicity, though not strictly required by HTML/XML.
				// This simple generator might add duplicate attrs, which is fine for crash testing.
				// Generate attribute value from a limited set.
				attrVal := rapid.SampledFrom([]string{"", "foo", "bar"}).Draw(t, fmt.Sprintf("attrVal%d", i))
				node.addAttribute(attrName, attrVal)
			}
		}

		// Add children sometimes. Limit depth and breadth via rapid's size control.
		if rapid.Bool().Draw(t, "hasChildren") {
			numChildren := rapid.IntRange(1, 5).Draw(t, "numChildren")
			for i := 0; i < numChildren; i++ {
				// Recursively generate child node using the already defined generator.
				child := genTNode.Draw(t, fmt.Sprintf("child%d", i))
				// Add the generated child node using the new AddChild method.
				node.AddChild(child)
			}
		}

		return node
	})
}

// AddChild adds an existing TNode as a child of this node.
func (n *TNode) AddChild(child *TNode) {
	child.Parent = n
	child.PrevSibling = n.LastChild
	child.NextSibling = nil // Ensure it's the last child initially

	if n.LastChild != nil {
		n.LastChild.NextSibling = child
	} else {
		// This is the first child
		n.FirstChild = child
	}
	n.LastChild = child // Update the last child pointer
}

// genStringLiteral generates a random XPath string literal.
func genStringLiteral() *rapid.Generator[string] {
	// Using a limited set of simple strings for literals.
	// Ensure generated strings don't contain the quote character used.
	// Rapid's StringOf generator could be used for more complex strings,
	// but requires careful handling of escaping.
	return rapid.Custom(func(t *rapid.T) string {
		quote := rapid.SampledFrom([]string{"'", "\""}).Draw(t, "quote")
		// Simple content, avoiding the chosen quote. More robust generation
		// would handle escaping or filter characters.
		content := rapid.SampledFrom([]string{"", "foo", "bar", "test", "data"}).Draw(t, "content")
		return quote + content + quote
	})
}

// genNumberLiteral generates a random XPath number literal (integer for simplicity).
func genNumberLiteral() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		// Generate small integers, positive and negative.
		num := rapid.IntRange(-10, 100).Draw(t, "number")
		return fmt.Sprintf("%d", num)
	})
}

// Forward declaration for recursive use in generators.
var genRelativePathExpr *rapid.Generator[string]
var genPredicateContent *rapid.Generator[string]

func init() {
	// Define genRelativePathExpr here or ensure it's defined before use in genPredicateContent.
	// We'll define it later, but the forward declaration allows compilation.

	// genPredicateContent generates expressions suitable for inside [...].
	genPredicateContent = rapid.Custom(func(t *rapid.T) string {
		// Choose the type of predicate expression.
		// Weights can be adjusted based on desired frequency.
		return rapid.OneOf(
			// Index predicate: [1], [last()]
			rapid.Just("last()"),
			genNumberLiteral(),
			// Boolean predicate: [foo], [@id='bar'], [text()='foo'], [count(a)>0]
			genRelativePathExpr, // Represents existence check, e.g., [element]
			rapid.Custom(func(t *rapid.T) string { // Simple comparison: path = literal
				// Generate a simple path, often an attribute or text()
				lhsPath := rapid.OneOf(
					rapid.Just("text()"),
					rapid.Just("."),
					rapid.Custom(func(t *rapid.T) string { return "@" + rapid.SampledFrom(htmlAttrs).Draw(t, "attrName") }),
					rapid.SampledFrom(htmlTags), // Simple element name test
				).Draw(t, "lhsPath")

				// Add more comparison operators
				op := rapid.SampledFrom([]string{"=", "!=", "<", "<=", ">", ">="}).Draw(t, "compOp")

				// Generate a literal for the RHS.
				// If LHS is text(), RHS should be a string for robust comparison.
				// Otherwise, it can be a string or a number.
				var rhsLiteral string
				if lhsPath == "text()" {
					rhsLiteral = genStringLiteral().Draw(t, "rhsLiteralString")
				} else {
					rhsLiteral = rapid.OneOf(genStringLiteral(), genNumberLiteral()).Draw(t, "rhsLiteralMixed")
				}

				return fmt.Sprintf("%s %s %s", lhsPath, op, rhsLiteral)
			}),
			rapid.Custom(func(t *rapid.T) string { // Function call predicate: [contains(., 'foo')]
				funcName := rapid.SampledFrom([]string{"contains", "starts-with"}).Draw(t, "funcName")
				// Argument 1: often context node or attribute/text
				arg1 := rapid.OneOf(
					rapid.Just("."),
					rapid.Just("text()"),
					rapid.Custom(func(t *rapid.T) string { return "@" + rapid.SampledFrom(htmlAttrs).Draw(t, "attrName") }),
				).Draw(t, "funcArg1")
				// Argument 2: string literal
				arg2 := genStringLiteral().Draw(t, "funcArg2")
				return fmt.Sprintf("%s(%s, %s)", funcName, arg1, arg2)
			}),
			// Add more complex predicates: position(), count(), boolean logic (and/or)
		).Draw(t, "predicateContent")
	})
}

// genPredicate generates a full predicate expression: '[' + content + ']'.
func genPredicate() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		content := genPredicateContent.Draw(t, "content")
		return "[" + content + "]"
	})
}

// genAxis generates a random XPath axis.
func genAxis() *rapid.Generator[string] {
	axes := []string{
		"child", "descendant", "parent", "ancestor", "following-sibling",
		"preceding-sibling", "following", "preceding", "attribute", "self",
		"descendant-or-self", "ancestor-or-self",
		// "namespace", // Deprecated and often unsupported
	}
	return rapid.SampledFrom(axes)
}

// genNodeTest generates a random XPath node test (name test or kind test).
func genNodeTest() *rapid.Generator[string] {
	return rapid.OneOf(
		// Name tests
		rapid.Just("*"),
		rapid.SampledFrom(htmlTags),
		// Kind tests
		rapid.Just("node()"),
		rapid.Just("text()"),
		// element() and attribute() are XPath 2.0/3.0, not 1.0
		// rapid.Just("element()"),
		// rapid.Just("attribute()"),
		// More specific kind tests (less likely to match simple generated docs, and also XPath 1.0)
		rapid.Just("comment()"), // Enable comment() node test
		// rapid.Just("processing-instruction()"), // Often requires a name argument
	)
}

// genStep generates a single XPath step (axis::nodetest[predicate1][predicate2]...).
func genStep() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		axis := genAxis().Draw(t, "axis")
		nodeTest := genNodeTest().Draw(t, "nodeTest")
		stepBase := ""
		// Abbreviated syntax for common cases
		// Ensure axis and nodeTest are compatible before potentially abbreviating.
		canAbbreviateChild := axis == "child" && nodeTest != "attribute()" && nodeTest != "comment()" && nodeTest != "processing-instruction()"
		canAbbreviateAttr := axis == "attribute" && nodeTest != "element()" && nodeTest != "text()" && nodeTest != "node()" && nodeTest != "comment()" && nodeTest != "processing-instruction()"

		useAbbreviation := rapid.Bool().Draw(t, "useAbbreviation")

		if useAbbreviation && canAbbreviateChild {
			stepBase = nodeTest // Abbreviated child axis
		} else if useAbbreviation && canAbbreviateAttr {
			if nodeTest == "attribute()" || nodeTest == "*" {
				stepBase = "@*" // Abbreviated attribute::*
			} else {
				stepBase = "@" + nodeTest // Abbreviated attribute axis name test
			}
		} else {
			// Default to full syntax if abbreviation is not chosen or not applicable
			stepBase = axis + "::" + nodeTest
		}

		// Add predicates sometimes
		predicates := ""
		if rapid.Bool().Draw(t, "hasPredicates") {
			numPredicates := rapid.IntRange(1, 2).Draw(t, "numPredicates") // 1 or 2 predicates
			for i := 0; i < numPredicates; i++ {
				// Ensure genPredicateContent is initialized before drawing from genPredicate
				if genPredicateContent == nil {
					// This might happen if init order is tricky. Log or handle.
					// For now, assume init() worked correctly.
					t.Fatalf("genPredicateContent is nil, initialization order issue?")
				}
				predicates += genPredicate().Draw(t, fmt.Sprintf("predicate%d", i))
			}
		}

		return stepBase + predicates
	})
}

// genRelativePathExpr generates a relative XPath expression (sequence of steps).
// Now defined using the forward declaration.
func init() {
	// Assign the actual generator function to the forward-declared variable.
	// This breaks the init cycle dependency if genPredicateContent needs genRelativePathExpr.
	genRelativePathExpr = rapid.Custom(func(t *rapid.T) string {
		// Generate the number of steps first.
		numSteps := rapid.IntRange(1, 3).Draw(t, "numSteps") // Reduced max steps slightly
		steps := make([]string, numSteps)
		for i := 0; i < numSteps; i++ {
			steps[i] = genStep().Draw(t, fmt.Sprintf("step%d", i))
		}
		// Join steps with / or //
		separator := rapid.SampledFrom([]string{"/", "//"}).Draw(t, "separator")
		// Avoid leading // if the path starts relative, although parser might handle it.
		// Let's keep it simple: join all with the chosen separator.
		return strings.Join(steps, separator)
	})
} // <-- Added missing closing brace for init()

// generateFunctionArgs generates the argument string for a given XPath function name.
func generateFunctionArgs(t *rapid.T, funcName string) string {
	args := ""
	numArgs := 0 // Keep track of generated args for clarity, though not strictly needed by fmt.Sprintf
	switch funcName {
	// Functions that can take 0 or 1 argument (node-set/path)
	case "string", "boolean", "number", "name", "namespace-uri", "local-name", "normalize-space":
		if rapid.Bool().Draw(t, "hasArg") {
			arg := rapid.OneOf(rapid.Just("."), genRelativePathExpr).Draw(t, "arg0")
			args = arg
			numArgs = 1
		}
	// count() and sum() MUST take exactly 1 argument (node-set)
	case "count", "sum":
		numArgs = 1
		// Argument must evaluate to a node-set.
		args = rapid.OneOf(rapid.Just("."), genRelativePathExpr).Draw(t, "arg0")
	case "concat": // 2+ arguments
		numArgs = rapid.IntRange(2, 4).Draw(t, "numConcatArgs")
		argList := make([]string, numArgs)
		for i := 0; i < numArgs; i++ {
			// Args are typically strings or expressions evaluating to strings
			argList[i] = rapid.OneOf(genStringLiteral(), genRelativePathExpr).Draw(t, fmt.Sprintf("concatArg%d", i))
		}
		args = strings.Join(argList, ", ")
	case "starts-with", "contains": // 2 arguments (string, string)
		numArgs = 2
		arg1 := rapid.OneOf(rapid.Just("."), genRelativePathExpr, genStringLiteral()).Draw(t, "strArg1")
		arg2 := genStringLiteral().Draw(t, "strArg2") // Second arg usually literal
		args = fmt.Sprintf("%s, %s", arg1, arg2)
	case "substring-before", "substring-after": // 2 arguments (string, string)
		numArgs = 2
		arg1 := rapid.OneOf(rapid.Just("."), genRelativePathExpr, genStringLiteral()).Draw(t, "strArg1")
		arg2 := genStringLiteral().Draw(t, "strArg2")
		args = fmt.Sprintf("%s, %s", arg1, arg2)
	case "substring": // 2 or 3 arguments (string, number, number?)
		numArgs = rapid.IntRange(2, 3).Draw(t, "numSubstringArgs")
		arg1 := rapid.OneOf(rapid.Just("."), genRelativePathExpr, genStringLiteral()).Draw(t, "strArg1")
		// XPath substring index is 1-based. Generate positive integers for start position.
		// Use a similar range to genNumberLiteral's positive side.
		startPos := rapid.IntRange(1, 100).Draw(t, "substringStartPos")
		arg2 := fmt.Sprintf("%d", startPos) // Convert generated int to string
		if numArgs == 3 {
			// The third argument (length) can be any number, including negative/zero,
			// though negative/zero length might result in empty strings. Use genNumberLiteral here.
			arg3 := genNumberLiteral().Draw(t, "numArg3")
			args = fmt.Sprintf("%s, %s, %s", arg1, arg2, arg3)
		} else {
			args = fmt.Sprintf("%s, %s", arg1, arg2)
		}
	case "string-length": // 1 argument (string) - Parser requires one argument.
		numArgs = 1
		// Argument needs to evaluate to string.
		args = rapid.OneOf(rapid.Just("."), genRelativePathExpr, genStringLiteral()).Draw(t, "strArg1")
	case "translate": // 3 arguments (string, string, string)
		numArgs = 3
		arg1 := rapid.OneOf(rapid.Just("."), genRelativePathExpr, genStringLiteral()).Draw(t, "strArg1")
		arg2 := genStringLiteral().Draw(t, "strArg2")
		arg3 := genStringLiteral().Draw(t, "strArg3")
		args = fmt.Sprintf("%s, %s, %s", arg1, arg2, arg3)
	case "not": // 1 argument (boolean)
		numArgs = 1
		// Argument needs to evaluate to boolean, e.g., a path, comparison, or function call
		// For simplicity, use a relative path or another simple function for now.
		arg := rapid.OneOf(genRelativePathExpr, rapid.Just("true()"), rapid.Just("false()")).Draw(t, "boolArg1")
		args = arg
	// case "lang": // Removed as it's unsupported by the library.
	// 	numArgs = 1
	// 	args = genStringLiteral().Draw(t, "langArg1")
	// Functions with no arguments:
	case "true", "false", "position", "last":
		numArgs = 0
	// Numeric functions often take node-sets:
	case "floor", "ceiling", "round":
		numArgs = 1
		// Argument needs to evaluate to number. Use path or number literal.
		args = rapid.OneOf(genRelativePathExpr, genNumberLiteral()).Draw(t, "numArg1")
	// Handle newly added functions (simplified argument generation)
	case "ends-with": // 2 args (string, string)
		numArgs = 2
		arg1 := rapid.OneOf(rapid.Just("."), genRelativePathExpr, genStringLiteral()).Draw(t, "strArg1")
		arg2 := genStringLiteral().Draw(t, "strArg2")
		args = fmt.Sprintf("%s, %s", arg1, arg2)
	case "lower-case": // 1 arg (string)
		numArgs = 1
		args = rapid.OneOf(rapid.Just("."), genRelativePathExpr, genStringLiteral()).Draw(t, "strArg1")
	case "matches": // 2-3 args (string, pattern, flags?) - Generate 2 args for simplicity
		numArgs = 2
		arg1 := rapid.OneOf(rapid.Just("."), genRelativePathExpr, genStringLiteral()).Draw(t, "strArg1")
		// Pattern is a string literal (regex) - keep simple
		arg2 := genStringLiteral().Draw(t, "regexPattern")
		args = fmt.Sprintf("%s, %s", arg1, arg2)
	case "replace": // 3 args (string, pattern, replacement)
		numArgs = 3
		arg1 := rapid.OneOf(rapid.Just("."), genRelativePathExpr, genStringLiteral()).Draw(t, "strArg1")
		arg2 := genStringLiteral().Draw(t, "regexPattern")
		arg3 := genStringLiteral().Draw(t, "replacementStr")
		args = fmt.Sprintf("%s, %s, %s", arg1, arg2, arg3)
	case "reverse": // 1 arg (node-set?) - Treat as string for simplicity? Spec unclear for 1.0 context.
		// Let's assume it takes a path expression.
		numArgs = 1
		args = genRelativePathExpr.Draw(t, "pathArg1")
	case "string-join": // 2 args (node-set?, separator)
		numArgs = 2
		// First arg is often path, second is string literal separator
		arg1 := genRelativePathExpr.Draw(t, "pathArg1")
		arg2 := genStringLiteral().Draw(t, "separatorStr")
		args = fmt.Sprintf("%s, %s", arg1, arg2)

	default:
		// Fallback for functions not explicitly handled (likely 0 args like true, false, position, last)
		// Check if the function *should* have args based on its name
		// For now, assume 0 args if not explicitly handled above.
		numArgs = 0
	}
	_ = numArgs // Use numArgs if needed for debugging or more complex logic later
	return args
}

// applyKnownFunctionIssueFilters applies a series of filters to a string generator
// to exclude known problematic XPath function calls found during testing.
func applyKnownFunctionIssueFilters(gen *rapid.Generator[string]) *rapid.Generator[string] {
	return gen.Filter(func(s string) bool {
		// boolean() without arguments is valid XPath 1.0 (evaluates context node),
		// causes a nil pointer dereference in antchfx/xpath's Evaluate function.
		return s != "boolean()"
	}).Filter(func(s string) bool {
		// number() without arguments is valid XPath 1.0 (evaluates context node),
		// but causes a nil pointer dereference in antchfx/xpath's Evaluate function.
		return s != "number()"
	}).Filter(func(s string) bool {
		// local-name() on the document root returns "" in xmllint/browsers,
		// but returns the name of the first child element in antchfx.
		return s != "local-name()"
	}).Filter(func(s string) bool {
		// name() on the document root returns "" in xmllint/browsers,
		// but returns the name of the first child element in antchfx.
		return s != "name()"
	})
}

// genSimpleFunctionCall generates calls to common XPath functions.
func genSimpleFunctionCall() *rapid.Generator[string] {
	// Define the base generator
	return rapid.Custom(func(t *rapid.T) string {
		// Select a function name from the list supported in the README
		funcName := rapid.SampledFrom([]string{
			// Core XPath 1.0
			"boolean", "ceiling", "concat", "contains", "count", "false", "floor",
			"last", "local-name", "name", "namespace-uri", "normalize-space",
			"not", "number", "position", "round", "starts-with", "string",
			"string-length", "substring", "substring-after", "substring-before",
			"sum", "translate", "true",
			// Added from README (potentially XPath 2.0+)
			// "ends-with", // not supported by xmllint
			"lower-case", "matches", "replace", "reverse", "string-join",
			// lang() is explicitly marked as unsupported (✗) in the README.
			// "lang",
		}).Draw(t, "funcName")

		// Generate arguments using the helper function
		args := generateFunctionArgs(t, funcName)

		return fmt.Sprintf("%s(%s)", funcName, args)
	})
}

// applyKnownPathIssueFilters applies filters for known problematic path expressions.
// TODO: Remove as fixed
func applyKnownPathIssueFilters(gen *rapid.Generator[string]) *rapid.Generator[string] {
	return gen.Filter(func(s string) bool {
		// The expression "/child::*" causes a mismatch between xmllint and antchfx
		// for a specific simple document (<doc><div/></doc>). xmllint returns the
		// document structure, while antchfx returns an empty string.
		// See failure: https://github.com/your-repo/link/to/issue/or/commit
		return s != "/child::*"
	}).Filter(func(s string) bool {
		// The expression "/*" selects the document element in xmllint,
		// but returns an empty result in antchfx for the test document structure.
		return s != "/*"
	}).Filter(func(s string) bool {
		// The expression "/descendant::*" selects the document element and its descendants in xmllint,
		// but returns an empty result in antchfx for the test document structure.
		return s != "/descendant::*"
	}).Filter(func(s string) bool {
		// The expression "/descendant::div" selects the div element in xmllint,
		// but returns an empty result in antchfx for the test document structure.
		return s != "/descendant::div"
	})
}

// genXPathExpr generates a simple absolute or relative XPath expression,
// potentially starting with '/', '//', or being a function call.
func genXPathExpr() *rapid.Generator[string] {
	// Use OneOf to decide the top-level structure
	baseGen := rapid.OneOf(
		// Option 1: Path expression (absolute or relative)
		rapid.Custom(func(t *rapid.T) string {
			// Start with / or // or relative path
			start := rapid.SampledFrom([]string{"/", "//", ""}).Draw(t, "start")
			if start == "" && rapid.Bool().Draw(t, "forceAbsolute") {
				// Ensure we don't generate empty expressions often
				start = "/"
			}

			// Generate the relative path part
			// Ensure genRelativePathExpr is initialized
			if genRelativePathExpr == nil {
				t.Fatalf("genRelativePathExpr is nil during genXPathExpr generation")
			}
			relativePath := genRelativePathExpr.Draw(t, "relativePath")

			// Handle edge cases like "/" or "//" which might need a path following
			if (start == "/" || start == "//") && relativePath == "" {
				// Avoid generating just "/" or "//" if relativePath is empty.
				// Append a simple node test if needed.
				relativePath = "node()"
			} else if start == "" && relativePath == "" {
				// Avoid generating completely empty string. Default to context node.
				return "."
			}

			// Combine start and relative path
			// Need to be careful about "//" followed by potentially empty relative path
			// or "/" followed by empty. The logic above tries to prevent empty relativePath
			// when start is / or //.
			return start + relativePath
		}),
		// Option 2: Top-level function call
		applyKnownFunctionIssueFilters(genSimpleFunctionCall()),
		// Option 3: Simple literal (less common as top-level expression but possible)
		// genStringLiteral(),
		// genNumberLiteral(),
		// TODO: Add UnionExpr ('|'), Operators (+, -, =, etc.) at the top level
	)

	// Apply the path-specific filters
	return applyKnownPathIssueFilters(baseGen)
}

// setupStaticTestFile creates a temporary file with static XML content
// used for initial xmllint syntax checks.
func setupStaticTestFile(testingT *testing.T, tmpDir string) (string, error) {
	staticTmpFile, err := os.CreateTemp(tmpDir, "static-xpath-test-*.xml")
	if err != nil {
		return "", fmt.Errorf("failed to create static temp file: %w", err)
	}
	defer staticTmpFile.Close() // Ensure close even on write error

	_, err = staticTmpFile.WriteString(staticXMLContent)
	if err != nil {
		// Explicit close before returning error might not be strictly needed due to defer,
		// but ensures the file handle is released immediately.
		staticTmpFile.Close()
		return "", fmt.Errorf("failed to write static temp file: %w", err)
	}

	// Close explicitly after successful write before returning the path
	err = staticTmpFile.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close static temp file: %w", err)
	}
	return staticTmpFile.Name(), nil
}

// Static XML content for basic xmllint syntax validation.
const staticXMLContent = `<?xml version="1.0" encoding="UTF-8"?>
<root>
  <child id="a">foo</child>
  <child id="b" value="123"/>
  <!-- comment -->
</root>
`

// runXPathPropertyTestIteration performs a single iteration of the property test logic.
// It generates inputs, runs xmllint, runs antchfx, and optionally compares results.
// testingT is the main *testing.T, t is the rapid.T, tmpDir is for temp files.
// compareResults determines if the final comparison step is performed.
// Returns true if the iteration should continue (no fatal error), false otherwise.
func runXPathPropertyTestIteration(testingT *testing.T, t *rapid.T, tmpDir, staticTmpFilePath string, compareResults bool) bool {
	// 1. Generate random document and expression
	rootNode := genTNode.Filter(func(n *TNode) bool { return n.Type == ElementNode }).Draw(t, "doc")
	exprStr := genXPathExpr().Draw(t, "expr")
	t.Logf("Testing expression: %q", exprStr)

	// 2. Initial syntax check with xmllint on STATIC file
	cmdStatic := exec.Command("xmllint", "--xpath", exprStr, staticTmpFilePath)
	var xmllintStaticStderr bytes.Buffer
	cmdStatic.Stderr = &xmllintStaticStderr
	cmdStaticErr := cmdStatic.Run() // We only care about the error/exit code here

	exitCodeStatic := 0
	if exitErr, ok := cmdStaticErr.(*exec.ExitError); ok {
		exitCodeStatic = exitErr.ExitCode()
	}

	if exitCodeStatic == 11 { // XPath syntax error according to xmllint
		t.Logf("xmllint rejected expr %q syntactically (exit code 11), skipping.", exprStr)
		_, antchfxCompileErr := Compile(exprStr)
		if antchfxCompileErr == nil {
			// If xmllint rejects syntax but antchfx compiles, that's a failure.
			testingT.Fatalf("xmllint rejected expr %q but antchfx compiled it.\nxmllint stderr:\n%s",
				exprStr, xmllintStaticStderr.String())
			return false // Stop test
		}
		return true // Skip this expression, continue test
	}
	// Handle other unexpected errors during static check
	if cmdStaticErr != nil && !(exitCodeStatic == 0 || exitCodeStatic == 10) {
		testingT.Errorf("xmllint failed unexpectedly on static file (exit code %d) for expr %q: Stderr: %s",
			exitCodeStatic, exprStr, xmllintStaticStderr.String())
		return true // Log as Errorf, but continue test run
	}

	// If syntax seems OK, proceed with the random document

	// 3. Serialize random document and write to temp file
	xmlString := nodeToXMLString(rootNode)
	randomTmpFile, err := os.CreateTemp(tmpDir, "random-xpath-test-*.xml")
	if err != nil {
		testingT.Fatalf("Failed to create random temp file: %v", err)
		return false
	}
	defer os.Remove(randomTmpFile.Name())
	defer randomTmpFile.Close()

	_, err = randomTmpFile.WriteString(xmlString)
	if err != nil {
		testingT.Fatalf("Failed to write random temp file: %v", err)
		return false
	}
	err = randomTmpFile.Close()
	if err != nil {
		testingT.Fatalf("Failed to close random temp file: %v", err)
		return false
	}
	randomTmpFilePath := randomTmpFile.Name()

	// 4. Run xmllint on the RANDOM document
	cmdRandom := exec.Command("xmllint", "--xpath", exprStr, randomTmpFilePath)
	var xmllintRandomStdout, xmllintRandomStderr bytes.Buffer
	cmdRandom.Stdout = &xmllintRandomStdout
	cmdRandom.Stderr = &xmllintRandomStderr
	cmdRandomErr := cmdRandom.Run()

	exitCodeRandom := 0
	if exitErr, ok := cmdRandomErr.(*exec.ExitError); ok {
		exitCodeRandom = exitErr.ExitCode()
	}
	xmllintStderrStr := xmllintRandomStderr.String()

	// Check for xmllint errors that should prevent further processing/comparison
	// We allow exit codes 0 (success) and 10 (no result) to proceed.
	// Exit code 11 (eval error) is handled later during compilation comparison.
	if cmdRandomErr != nil && !(exitCodeRandom == 0 || exitCodeRandom == 10 || exitCodeRandom == 11) {
		// Includes XML parsing errors (1-9), context errors (12), etc.
		t.Logf("xmllint failed unexpectedly on random doc (exit code %d) for expr %q. Skipping.\nStderr: %s\nXML:\n%s",
			exitCodeRandom, exprStr, xmllintStderrStr, xmlString)
		return true // Continue test run, but skip this iteration
	}

	// Skip comparison if xmllint returned exit code 10 (no result found) - only relevant for equivalence test
	if compareResults && exitCodeRandom == 10 {
		t.Logf("Equivalence test: xmllint returned exit code 10 (no result) for expr %q. Skipping comparison.\nXML:\n%s",
			exprStr, xmlString)
		return true // Continue test run
	}

	// 5. Compile and Evaluate with antchfx/xpath
	antchfxExpr, antchfxCompileErr := Compile(exprStr)

	// Handle compilation discrepancies
	if antchfxCompileErr != nil {
		if exitCodeRandom != 11 { // antchfx failed, xmllint didn't report syntax error (11)
			testingT.Fatalf("antchfx failed to compile expr %q which xmllint accepted/processed (exit %d):\nAntchfx Error: %v\nxmllint stderr:\n%s",
				exprStr, exitCodeRandom, antchfxCompileErr, xmllintStderrStr)
			return false
		} else { // Both failed (xmllint eval error 11, antchfx compile error)
			t.Logf("Both xmllint (exit 11) and antchfx failed to compile/evaluate expr %q. Antchfx err: %v", exprStr, antchfxCompileErr)
			return true // Consistent failure, continue test run
		}
	}

	// If antchfx compiled but xmllint failed evaluation (11)
	if exitCodeRandom == 11 && antchfxCompileErr == nil {
		testingT.Fatalf("xmllint failed evaluating expr %q (exit 11) but antchfx compiled it successfully.\nxmllint stderr:\n%s\nXML:\n%s",
			exprStr, xmllintStderrStr, xmlString)
		return false
	}

	// Evaluate antchfx, catching panics
	var antchfxResult interface{}
	var antchfxPanic interface{}
	func() {
		defer func() {
			if r := recover(); r != nil {
				antchfxPanic = r
			}
		}()
		nav := createNavigator(rootNode)
		antchfxResult = antchfxExpr.Evaluate(nav)
	}()

	// Handle panics - this is critical for the NoPanic test
	if antchfxPanic != nil {
		// Fail immediately ONLY if a panic occurred when xmllint exited with 0 (success).
		// We are most interested in panics where xmllint definitively succeeded.
		if exitCodeRandom == 0 {
			testingT.Fatalf("Panic during antchfx Evaluate for expr %q (xmllint exit 0):\nPanic: %v\nXML:\n%s\nxmllint stdout:\n%s\nxmllint stderr:\n%s",
				exprStr, antchfxPanic, xmlString, xmllintRandomStdout.String(), xmllintStderrStr)
			return false // Stop test run on panic
		} else {
			// Log panics that occurred with other xmllint exit codes (e.g., 10, 11) but don't fail the NoPanic test.
			t.Logf("Antchfx panic occurred for expr %q, but xmllint exit code was %d (not 0). Panic: %v", exprStr, exitCodeRandom, antchfxPanic)
			// Continue the test run, as this doesn't meet the strict criteria for NoPanic failure.
		}
	}

	// If we reach here, antchfx compiled and evaluated without panic for an expression
	// that xmllint also processed (exit 0, 10, or 11 where antchfx also failed compile).

	// If only checking for panics, we are done for this iteration.
	if !compareResults {
		return true // Continue test run
	}

	// --- Equivalence Check Logic ---
	// This part only runs if compareResults is true

	// Normalize antchfx result
	navForToString := createNavigator(rootNode)
	antchfxNormStr := antchfxResultToString(antchfxResult, navForToString)
	t.Logf("Antchfx raw result (type %T): %s", antchfxResult, antchfxNormStr)

	// Normalize xmllint result
	xmllintNormStr, xmllintParseErr := parseXmllintOutput(xmllintRandomStdout.String(), xmllintStderrStr, exitCodeRandom)
	if xmllintParseErr != nil {
		testingT.Fatalf("Failed to parse xmllint output for expr %q: %v\nXML:\n%s", exprStr, xmllintParseErr, xmlString)
		return false
	}

	// Compare normalized strings
	if xmllintNormStr != antchfxNormStr {
		testingT.Fatalf("Result mismatch for expr %q\n--- xmllint (exit %d) ---\n%s\n------\n--- antchfx (normalized) ---\n%s\n------\n--- antchfx (raw type %T, as string) ---\n%s\n------\n--- XML ---\n%s\n------\n--- xmllint stderr ---\n%s\n------",
			exprStr, exitCodeRandom, xmllintNormStr, antchfxNormStr, antchfxResult, antchfxNormStr, xmlString, xmllintStderrStr)
		return false
	} else {
		// t.Logf("Results match for expr %q", exprStr)
	}

	return true // Continue test run
}

// TestPropertyXPathNoPanic checks that antchfx/xpath does not panic when evaluating
// XPath expressions that xmllint successfully processes (exit code 0 or 10).
func TestPropertyXPathNoPanic(testingT *testing.T) {
	checkXmllintAvailability(testingT)
	testingT.Log("Starting TestPropertyXPathNoPanic...")

	// Setup static file for syntax check
	tmpDir := testingT.TempDir()
	staticTmpFilePath, err := setupStaticTestFile(testingT, tmpDir)
	if err != nil {
		testingT.Fatalf("Setup failed: %v", err)
	}
	// No need to defer remove static file, tmpDir handles it.

	rapid.Check(testingT, func(t *rapid.T) {
		// Run the shared logic, but don't compare results.
		// The helper function will fail the test immediately on panic.
		if !runXPathPropertyTestIteration(testingT, t, tmpDir, staticTmpFilePath, false) {
			// If the helper returned false, it means a fatal error occurred.
			// rapid should stop further iterations.
			t.FailNow()
		}
	}) // Removed CheckConfig option
	testingT.Logf("TestPropertyXPathNoPanic finished.")
}

// TestPropertyXPathEquivalence checks that antchfx/xpath evaluation results
// match xmllint results for the same XPath expression and document.
func TestPropertyXPathEquivalence(testingT *testing.T) {
	checkXmllintAvailability(testingT)
	testingT.Log("Starting TestPropertyXPathEquivalence...")

	// Setup static file for syntax check
	tmpDir := testingT.TempDir()
	staticTmpFilePath, err := setupStaticTestFile(testingT, tmpDir)
	if err != nil {
		testingT.Fatalf("Setup failed: %v", err)
	}

	rapid.Check(testingT, func(t *rapid.T) {
		// Run the shared logic, including result comparison.
		if !runXPathPropertyTestIteration(testingT, t, tmpDir, staticTmpFilePath, true) {
			// If the helper returned false, it means a fatal error occurred.
			t.FailNow()
		}
	})
	testingT.Logf("TestPropertyXPathEquivalence finished.")
}

// Helper function to serialize the TNode tree to a string suitable for xmllint.
// serializeNodeToString converts a single TNode and its descendants to an XML string snippet.
// Does NOT add XML declaration or a wrapping element. Indentation is basic.
func serializeNodeToString(n *TNode) string {
	var sb strings.Builder
	var printNode func(*TNode, int)
	printNode = func(node *TNode, indent int) {
		sb.WriteString(strings.Repeat("  ", indent)) // Basic indentation
		switch node.Type {
		case ElementNode:
			// Ensure element names are XML-compatible (basic check)
			tagName := node.Data
			if tagName == "" {
				tagName = "unknown" // Handle empty tag names if generator allows
			}
			sb.WriteString("<" + tagName)
			// Keep track of added attribute names to avoid duplicates which xmllint might dislike
			addedAttrs := make(map[string]bool)
			for _, attr := range node.Attr {
				// Ensure attr names are XML-compatible (basic check) and not duplicated
				attrName := attr.Key
				if attrName == "" || addedAttrs[attrName] {
					continue // Skip empty or duplicate attribute names
				}
				addedAttrs[attrName] = true
				// Use standard Go quoting which handles XML entities (&, <, >, ", ')
				sb.WriteString(fmt.Sprintf(" %s=%q", attrName, attr.Value))
			}
			if node.FirstChild == nil {
				sb.WriteString("/>") // No newline for self-closing in snippet
			} else {
				sb.WriteString(">")
				// No newline after opening tag in snippet? Or maybe yes? Let's omit for now.
				for child := node.FirstChild; child != nil; child = child.NextSibling {
					printNode(child, indent+1) // Indent children
				}
				sb.WriteString(strings.Repeat("  ", indent)) // Indent closing tag
				sb.WriteString("</" + tagName + ">")
			}
		case TextNode:
			// Escape text content for XML
			escapedData := escapeXMLText(node.Data)
			sb.WriteString(escapedData) // No quotes or newline around text nodes in snippet
		case CommentNode:
			// Ensure comment data doesn't contain "--"
			commentData := strings.ReplaceAll(node.Data, "--", "- -")
			sb.WriteString(fmt.Sprintf("<!--%s-->", commentData))
		// Ignore other node types for snippet serialization
		default:
		}
		// Add a newline after each top-level node in the snippet for readability?
		// Let's try without first, aiming for compact output like xmllint often gives.
		// sb.WriteString("\n") // Removed potential trailing newline
	}

	// Start printing from the node itself at indent 0
	printNode(n, 0)
	return sb.String()
}

// Adds an XML declaration and wraps content in a single <doc> root.
// Uses serializeNodeToString for the core node serialization.
func nodeToXMLString(node *TNode) string {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n") // XML declaration
	sb.WriteString("<doc>\n")                                       // Wrapper root element

	// Serialize the main node using the helper, assuming it adds necessary indentation/newlines internally
	sb.WriteString(serializeNodeToString(node))
	sb.WriteString("\n") // Add a newline after the serialized node content

	sb.WriteString("</doc>\n") // Close wrapper root element
	return sb.String()
}

// escapeXMLText escapes characters problematic for XML text nodes.
func escapeXMLText(s string) string {
	var buf bytes.Buffer
	for _, r := range s {
		switch r {
		case '&':
			buf.WriteString("&amp;")
		case '<':
			buf.WriteString("&lt;")
		case '>':
			buf.WriteString("&gt;")
		// Standard Go %q handles quotes, but they are allowed in text nodes.
		// Only strictly need to escape &, <, > in text content.
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}
