package build

import (
	"errors"
	"fmt"

	"github.com/antchfx/gxpath/internal/parse"
	"github.com/antchfx/gxpath/internal/query"
	"github.com/antchfx/gxpath/xpath"
)

type flag int

const (
	noneFlag flag = iota
	filterFlag
)

// builder provides building an XPath expressions.
type builder struct {
	depth      int
	flag       flag
	firstInput query.Query
}

// axisPredicate creates a predicate to predicating for this axis node.
func axisPredicate(root *parse.AxisNode) func(xpath.NodeNavigator) bool {
	// get current axix node type.
	typ := xpath.ElementNode
	if root.AxeType == "attribute" {
		typ = xpath.AttributeNode
	} else {
		switch root.Prop {
		case "comment":
			typ = xpath.CommentNode
		case "text":
			typ = xpath.TextNode
		case "processing-instruction":
			typ = xpath.ProcessingInstructionNode
		case "node":
			typ = xpath.ElementNode
		}
	}
	predicate := func(n xpath.NodeNavigator) bool {
		if typ == n.NodeType() {
			if root.LocalName == "" || (root.LocalName == n.LocalName() && root.Prefix == n.Prefix()) {
				return true
			}
		}
		return false
	}

	return predicate
}

// processAxisNode buildes a query for the XPath axis node.
func (b *builder) processAxisNode(root *parse.AxisNode) (query.Query, error) {
	var (
		err       error
		qyInput   query.Query
		qyOutput  query.Query
		predicate = axisPredicate(root)
	)

	if root.Input == nil {
		qyInput = &query.ContextQuery{}
	} else {
		if b.flag&filterFlag == 0 {
			if root.AxeType == "child" && (root.Input.Type() == parse.NodeAxis) {
				if input := root.Input.(*parse.AxisNode); input.AxeType == "descendant-or-self" {
					var qyGrandInput query.Query
					if input.Input != nil {
						qyGrandInput, _ = b.processNode(input.Input)
					} else {
						qyGrandInput = &query.ContextQuery{}
					}
					qyOutput = &query.DescendantQuery{Input: qyGrandInput, Predicate: predicate, Self: true}
					return qyOutput, nil
				}
			}
		}
		qyInput, err = b.processNode(root.Input)
		if err != nil {
			return nil, err
		}
	}

	switch root.AxeType {
	case "ancestor":
		qyOutput = &query.AncestorQuery{Input: qyInput, Predicate: predicate}
	case "ancestor-or-self":
		qyOutput = &query.AncestorQuery{Input: qyInput, Predicate: predicate, Self: true}
	case "attribute":
		qyOutput = &query.AttributeQuery{Input: qyInput, Predicate: predicate}
	case "child":
		qyOutput = &query.ChildQuery{Input: qyInput, Predicate: predicate}
	case "descendant":
		qyOutput = &query.DescendantQuery{Input: qyInput, Predicate: predicate}
	case "descendant-or-self":
		qyOutput = &query.DescendantQuery{Input: qyInput, Predicate: predicate, Self: true}
	case "following":
		qyOutput = &query.FollowingQuery{Input: qyInput, Predicate: predicate}
	case "following-sibling":
		qyOutput = &query.FollowingQuery{Input: qyInput, Predicate: predicate, Sibling: true}
	case "parent":
		qyOutput = &query.ParentQuery{Input: qyInput, Predicate: predicate}
	case "preceding":
		qyOutput = &query.PrecedingQuery{Input: qyInput, Predicate: predicate}
	case "preceding-sibling":
		qyOutput = &query.PrecedingQuery{Input: qyInput, Predicate: predicate, Sibling: true}
	case "self":
		qyOutput = &query.SelfQuery{Input: qyInput}
	case "namespace":
		// haha,what will you do someting??
	default:
		err = fmt.Errorf("unknown axe type: %s", root.AxeType)
		return nil, err
	}
	return qyOutput, nil
}

// processFilterNode builds query.Query for the XPath filter predicate.
func (b *builder) processFilterNode(root *parse.FilterNode) (query.Query, error) {
	b.flag |= filterFlag

	qyInput, err := b.processNode(root.Input)
	if err != nil {
		return nil, err
	}
	qyCond, err := b.processNode(root.Condition)
	if err != nil {
		return nil, err
	}
	qyOutput := &query.FilterQuery{Input: qyInput, Predicate: qyCond}
	return qyOutput, nil
}

// processFunctionNode buildes query.Query for the XPath function node.
func (b *builder) processFunctionNode(root *parse.FunctionNode) (query.Query, error) {
	var qyOutput query.Query
	switch root.FuncName {
	case "last":
		qyOutput = &query.XPathFunction{Input: b.firstInput, Func: lastFunc}
	case "position":
		qyOutput = &query.XPathFunction{Input: b.firstInput, Func: positionFunc}
	case "count":
		if b.firstInput == nil {
			return nil, errors.New("xpath: expression must evaluate to node-set")
		}
		if len(root.Args) == 0 {
			return nil, fmt.Errorf("xpath: count(node-sets) function must with have parameters node-sets")
		}
		argQuery, err := b.processNode(root.Args[0])
		if err != nil {
			return nil, err
		}
		qyOutput = &query.XPathFunction{Input: argQuery, Func: countFunc}
	default:
		return nil, fmt.Errorf("not yet support this function %s()", root.FuncName)
	}
	return qyOutput, nil
}

func (b *builder) processOperatorNode(root *parse.OperatorNode) (query.Query, error) {
	left, err := b.processNode(root.Left)
	if err != nil {
		return nil, err
	}
	right, err := b.processNode(root.Right)
	if err != nil {
		return nil, err
	}
	var qyOutput query.Query
	switch root.Op {
	case "+", "-", "div", "mod": // Numeric operator
		var exprFunc func(interface{}, interface{}) interface{}
		switch root.Op {
		case "+":
			exprFunc = plusFunc
		case "-":
			exprFunc = minusFunc
		case "div":
			exprFunc = divFunc
		case "mod":
			exprFunc = modFunc
		}
		qyOutput = &query.NumericExpr{Left: left, Right: right, Do: exprFunc}
	case "=", ">", ">=", "<", "<=", "!=":
		var exprFunc func(query.Iterator, interface{}, interface{}) interface{}
		switch root.Op {
		case "=":
			exprFunc = eqFunc
		case ">":
			exprFunc = gtFunc
		case ">=":
			exprFunc = geFunc
		case "<":
			exprFunc = ltFunc
		case "<=":
			exprFunc = leFunc
		case "!=":
		}
		qyOutput = &query.LogicalExpr{Left: left, Right: right, Do: exprFunc}
	case "or", "and":
		isOr := root.Op == "or"
		qyOutput = &query.BooleanExpr{Left: left, Right: right, IsOr: isOr}
	}
	return qyOutput, nil
}

func (b *builder) processNode(root parse.Node) (q query.Query, err error) {
	if b.depth = b.depth + 1; b.depth > 1024 {
		err = errors.New("the xpath expressions is too complex")
		return
	}

	switch root.Type() {
	case parse.NodeConstantOperand:
		n := root.(*parse.OperandNode)
		q = &query.XPathConstant{Val: n.Val}
	case parse.NodeRoot:
		q = &query.ContextQuery{Root: true}
	case parse.NodeAxis:
		q, err = b.processAxisNode(root.(*parse.AxisNode))
		b.firstInput = q
	case parse.NodeFilter:
		q, err = b.processFilterNode(root.(*parse.FilterNode))
	case parse.NodeFunction:
		q, err = b.processFunctionNode(root.(*parse.FunctionNode))
	case parse.NodeOperator:
		q, err = b.processOperatorNode(root.(*parse.OperatorNode))
	}
	return
}

// Build builds a specified XPath expressions expr.
func Build(expr string) (query.Query, error) {
	root := parse.Parse(expr)
	b := &builder{}
	return b.processNode(root)
}
