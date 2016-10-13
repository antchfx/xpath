// https://www.w3.org/TR/xpath/

package parse

import "fmt"

type parser struct {
	r *scanner
	d int
}

// newOperatorNode returns new operator node OperatorNode.
func newOperatorNode(op string, left, right Node) Node {
	return &OperatorNode{NodeType: NodeOperator, Op: op, Left: left, Right: right}
}

// newOperand returns new constant operand node OperandNode.
func newOperandNode(v interface{}) Node {
	return &OperandNode{NodeType: NodeConstantOperand, Val: v}
}

// newAxisNode returns new axis node AxisNode.
func newAxisNode(axeTyp, localName, prefix, prop string, n Node) Node {
	return &AxisNode{
		NodeType:  NodeAxis,
		LocalName: localName,
		Prefix:    prefix,
		AxeType:   axeTyp,
		Prop:      prop,
		Input:     n,
	}
}

// newVariableNode returns new variable node VariableNode.
func newVariableNode(prefix, name string) Node {
	return &VariableNode{NodeType: NodeVariable, Name: name, Prefix: prefix}
}

// newFilterNode returns a new filter node FilterNode.
func newFilterNode(n, m Node) Node {
	return &FilterNode{NodeType: NodeFilter, Input: n, Condition: m}
}

// newRootNode returns a root node.
func newRootNode(s string) Node {
	return &RootNode{NodeType: NodeRoot, slash: s}
}

// newFunctionNode returns function call node.
func newFunctionNode(name, prefix string, args []Node) Node {
	return &FunctionNode{NodeType: NodeFunction, Prefix: prefix, FuncName: name, Args: args}
}

// testOp reports whether current item name is an operand op.
func testOp(r *scanner, op string) bool {
	return r.typ == itemName && r.prefix == "" && r.name == op
}

func isPrimaryExpr(r *scanner) bool {
	switch r.typ {
	case itemString, itemNumber, itemDollar, itemLParens:
		return true
	case itemName:
		return r.canBeFunc && !isNodeType(r)
	}
	return false
}

func isNodeType(r *scanner) bool {
	switch r.name {
	case "node", "text", "processing-instruction", "comment":
		return r.prefix == ""
	}
	return false
}

func isStep(item itemType) bool {
	switch item {
	case itemDot, itemDotDot, itemAt, itemAxe, itemStar, itemName:
		return true
	}
	return false
}

func checkItem(r *scanner, typ itemType) {
	if r.typ != typ {
		panic(fmt.Sprintf("%s has an invalid token", r.text))
	}
}

// parseExpression parsing the expression with input Node n.
func (p *parser) parseExpression(n Node) Node {
	if p.d = p.d + 1; p.d > 200 {
		panic("the xpath query is too complex(depth > 200)")
	}
	n = p.parseOrExpr(n)
	p.d--
	return n
}

// next scanning next item on forward.
func (p *parser) next() bool {
	return p.r.nextItem()
}

func (p *parser) skipItem(typ itemType) {
	checkItem(p.r, typ)
	p.next()
}

// OrExpr ::= AndExpr | OrExpr 'or' AndExpr
func (p *parser) parseOrExpr(n Node) Node {
	opnd := p.parseAndExpr(n)
	for {
		if !testOp(p.r, "or") {
			break
		}
		p.next()
		opnd = newOperatorNode("or", opnd, p.parseAndExpr(n))
	}
	return opnd
}

// AndExpr ::= EqualityExpr	| AndExpr 'and' EqualityExpr
func (p *parser) parseAndExpr(n Node) Node {
	opnd := p.parseEqualityExpr(n)
	for {
		if !testOp(p.r, "and") {
			break
		}
		p.next()
		opnd = newOperatorNode("and", opnd, p.parseEqualityExpr(n))
	}
	return opnd
}

// EqualityExpr ::= RelationalExpr | EqualityExpr '=' RelationalExpr | EqualityExpr '!=' RelationalExpr
func (p *parser) parseEqualityExpr(n Node) Node {
	opnd := p.parseRelationalExpr(n)
Loop:
	for {
		var op string
		switch p.r.typ {
		case itemEq:
			op = "="
		case itemNe:
			op = "!="
		default:
			break Loop
		}
		p.next()
		opnd = newOperatorNode(op, opnd, p.parseRelationalExpr(n))
	}
	return opnd
}

// RelationalExpr ::= AdditiveExpr	| RelationalExpr '<' AdditiveExpr | RelationalExpr '>' AdditiveExpr
//					| RelationalExpr '<=' AdditiveExpr
//					| RelationalExpr '>=' AdditiveExpr
func (p *parser) parseRelationalExpr(n Node) Node {
	opnd := p.parseAdditiveExpr(n)
Loop:
	for {
		var op string
		switch p.r.typ {
		case itemLt:
			op = "<"
		case itemGt:
			op = ">"
		case itemLe:
			op = "<="
		case itemGe:
			op = ">="
		default:
			break Loop
		}
		p.next()
		opnd = newOperatorNode(op, opnd, p.parseAdditiveExpr(n))
	}
	return opnd
}

// AdditiveExpr	::= MultiplicativeExpr	| AdditiveExpr '+' MultiplicativeExpr | AdditiveExpr '-' MultiplicativeExpr
func (p *parser) parseAdditiveExpr(n Node) Node {
	opnd := p.parseMultiplicativeExpr(n)
Loop:
	for {
		var op string
		switch p.r.typ {
		case itemPlus:
			op = "+"
		case itemMinus:
			op = "-"
		default:
			break Loop
		}
		p.next()
		opnd = newOperatorNode(op, opnd, p.parseMultiplicativeExpr(n))
	}
	return opnd
}

// MultiplicativeExpr ::= UnaryExpr	| MultiplicativeExpr MultiplyOperator(*) UnaryExpr
//						| MultiplicativeExpr 'div' UnaryExpr | MultiplicativeExpr 'mod' UnaryExpr
func (p *parser) parseMultiplicativeExpr(n Node) Node {
	opnd := p.parseUnaryExpr(n)
Loop:
	for {
		var op string
		if p.r.typ == itemStar {
			op = "*"
		} else if testOp(p.r, "div") || testOp(p.r, "mod") {
			op = p.r.name
		} else {
			break Loop
		}
		p.next()
		opnd = newOperatorNode(op, opnd, p.parseUnaryExpr(n))
	}
	return opnd
}

// UnaryExpr ::= UnionExpr | '-' UnaryExpr
func (p *parser) parseUnaryExpr(n Node) Node {
	minus := false
	// ignore '-' sequence
	for p.r.typ == itemMinus {
		p.next()
		minus = !minus
	}
	opnd := p.parseUnionExpr(n)
	if minus {
		opnd = newOperatorNode("*", opnd, newOperandNode(float64(-1)))
	}
	return opnd
}

// 	UnionExpr ::= PathExpr | UnionExpr '|' PathExpr
func (p *parser) parseUnionExpr(n Node) Node {
	opnd := p.parsePathExpr(n)
Loop:
	for {
		if p.r.typ != itemUnion {
			break Loop
		}
		p.next()
		opnd2 := p.parsePathExpr(n)
		// Checking the node type that must be is node set type?
		opnd = newOperatorNode("|", opnd, opnd2)
	}
	return opnd
}

// PathExpr ::= LocationPath | FilterExpr | FilterExpr '/' RelativeLocationPath	| FilterExpr '//' RelativeLocationPath
func (p *parser) parsePathExpr(n Node) Node {
	var opnd Node
	if isPrimaryExpr(p.r) {
		opnd = p.parseFilterExpr(n)
		switch p.r.typ {
		case itemSlash:
			p.next()
			opnd = p.parseRelativeLocationPath(opnd)
		case itemSlashSlash:
			p.next()
			opnd = p.parseRelativeLocationPath(newAxisNode("descendant-or-self", "", "", "", opnd))
		}
	} else {
		opnd = p.parseLocationPath(nil)
	}
	return opnd
}

// FilterExpr ::= PrimaryExpr | FilterExpr Predicate
func (p *parser) parseFilterExpr(n Node) Node {
	opnd := p.parsePrimaryExpr(n)
	if p.r.typ == itemLBracket {
		opnd = newFilterNode(opnd, p.parsePredicate(opnd))
	}
	return opnd
}

// 	Predicate ::=  '[' PredicateExpr ']'
func (p *parser) parsePredicate(n Node) Node {
	p.skipItem(itemLBracket)
	opnd := p.parseExpression(n)
	p.skipItem(itemRBracket)
	return opnd
}

// LocationPath ::= RelativeLocationPath | AbsoluteLocationPath
func (p *parser) parseLocationPath(n Node) (opnd Node) {
	switch p.r.typ {
	case itemSlash:
		p.next()
		opnd = newRootNode("/")
		if isStep(p.r.typ) {
			opnd = p.parseRelativeLocationPath(opnd) // ?? child:: or self ??
		}
	case itemSlashSlash:
		p.next()
		opnd = newRootNode("//")
		opnd = p.parseRelativeLocationPath(newAxisNode("descendant-or-self", "", "", "", opnd))
	default:
		opnd = p.parseRelativeLocationPath(n)
	}
	return opnd
}

// RelativeLocationPath	 ::= Step | RelativeLocationPath '/' Step | AbbreviatedRelativeLocationPath
func (p *parser) parseRelativeLocationPath(n Node) Node {
	opnd := n
Loop:
	for {
		opnd = p.parseStep(opnd)
		switch p.r.typ {
		case itemSlashSlash:
			p.next()
			opnd = newAxisNode("descendant-or-self", "", "", "", opnd)
		case itemSlash:
			p.next()
		default:
			break Loop
		}
	}
	return opnd
}

// Step	::= AxisSpecifier NodeTest Predicate* | AbbreviatedStep
func (p *parser) parseStep(n Node) Node {
	axeTyp := "child" // default axes value.
	if p.r.typ == itemDot || p.r.typ == itemDotDot {
		if p.r.typ == itemDot {
			axeTyp = "self"
		} else {
			axeTyp = "parent"
		}
		p.next()
		return newAxisNode(axeTyp, "", "", "", n)
	}
	switch p.r.typ {
	case itemAt:
		p.next()
		axeTyp = "attribute"
	case itemAxe:
		axeTyp = p.r.name
		p.next()
	}
	opnd := p.parseNodeTest(n, axeTyp)
	if p.r.typ == itemLBracket {
		opnd = newFilterNode(opnd, p.parsePredicate(opnd))
	}
	return opnd
}

// 	NodeTest ::= NameTest | NodeType '(' ')' | 'processing-instruction' '(' Literal ')'
func (p *parser) parseNodeTest(n Node, axeTyp string) (opnd Node) {
	switch p.r.typ {
	case itemName:
		if p.r.canBeFunc && isNodeType(p.r) {
			var prop string
			switch p.r.name {
			case "comment", "text", "processing-instruction", "node":
				prop = p.r.name
			}
			var name string
			p.next()
			p.skipItem(itemLParens)
			if prop == "processing-instruction" && p.r.typ != itemRParens {
				checkItem(p.r, itemString)
				name = p.r.strval
				p.next()
			}
			p.skipItem(itemRParens)
			opnd = newAxisNode(axeTyp, name, "", prop, n)
		} else {
			prefix := p.r.prefix
			name := p.r.name
			p.next()
			if p.r.name == "*" {
				name = ""
			}
			opnd = newAxisNode(axeTyp, name, prefix, "", n)
		}
	case itemStar:
		opnd = newAxisNode(axeTyp, "", "", "", n)
		p.next()
	default:
		panic("expression must evaluate to a node-set")
	}
	return opnd
}

// PrimaryExpr ::= VariableReference | '(' Expr ')'	| Literal | Number | FunctionCall
func (p *parser) parsePrimaryExpr(n Node) (opnd Node) {
	switch p.r.typ {
	case itemString:
		opnd = newOperandNode(p.r.strval)
		p.next()
	case itemNumber:
		opnd = newOperandNode(p.r.numval)
		p.next()
	case itemDollar:
		p.next()
		checkItem(p.r, itemName)
		opnd = newVariableNode(p.r.prefix, p.r.name)
		p.next()
	case itemLParens:
		p.next()
		opnd = p.parseExpression(n)
		p.skipItem(itemRParens)
	case itemName:
		if p.r.canBeFunc && !isNodeType(p.r) {
			opnd = p.parseMethod(nil)
		}
	}
	return opnd
}

// FunctionCall	 ::=  FunctionName '(' ( Argument ( ',' Argument )* )? ')'
func (p *parser) parseMethod(n Node) Node {
	var args []Node
	name := p.r.name
	prefix := p.r.prefix

	p.skipItem(itemName)
	p.skipItem(itemLParens)
	if p.r.typ != itemRParens {
		for {
			args = append(args, p.parseExpression(n))
			if p.r.typ == itemRParens {
				break
			}
			p.skipItem(itemComma)
		}
	}
	p.skipItem(itemRParens)
	return newFunctionNode(name, prefix, args)
}

// Parse parsing the XPath express string expr and returns a tree Node.
func Parse(expr string) Node {
	r := &scanner{text: expr}
	r.nextChar()
	r.nextItem()
	p := &parser{r: r}
	return p.parseExpression(nil)
}
