package xpath

import (
	"errors"
	"fmt"
	"strings"
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
	firstInput query
}

type Predicate interface {
	Test(n NodeNavigator) bool
}

type PredicateFunc func(NodeNavigator) bool

func (f PredicateFunc) Test(n NodeNavigator) bool {
	return f(n)
}

type BoolPredicate bool

func (p BoolPredicate) Test(n NodeNavigator) bool {
	return bool(p)
}

func (p *BoolPredicate) String() string {
	if *p {
		return "True"
	} else {
		return "False"
	}
}

var TruePredicate BoolPredicate = BoolPredicate(true)
var FalsePredicate BoolPredicate = BoolPredicate(false)

type TypePredicate []NodeType

func (p *TypePredicate) Test(n NodeNavigator) bool {
	if len(*p) == 0 {
		return true
	}
	typ := n.NodeType()
	for _, t := range *p {
		if typ == t {
			return true
		}
	}
	return false
}

func (p *TypePredicate) String() string {
	return fmt.Sprintf("TypePredicate{%v}", []NodeType(*p))
}

type AndPredicate []Predicate

func (p *AndPredicate) Test(n NodeNavigator) bool {
	for _, pp := range *p {
		if !pp.Test(n) {
			return false
		}
	}
	return true
}

func (p *AndPredicate) string() string {
	var res []string
	for _, p := range *p {
		res = append(res, fmt.Sprintf("%s", p))
	}
	return strings.Join(res, " && ")
}

type OrPredicate []Predicate

func (p *OrPredicate) Test(n NodeNavigator) bool {
	for _, pp := range *p {
		if pp.Test(n) {
			return true
		}
	}
	return false
}

func (p *OrPredicate) string() string {
	var res []string
	for _, p := range *p {
		res = append(res, fmt.Sprintf("%s", p))
	}
	return strings.Join(res, " || ")
}

type NamePredicate struct {
	Prefix    string
	LocalName string
}

func (p *NamePredicate) Test(n NodeNavigator) bool {
	return p.LocalName == "" || (p.LocalName == n.LocalName() && p.Prefix == n.Prefix())
}

func (p *NamePredicate) string() string {
	return fmt.Sprintf("NamePredicate{%#v, %#v}", p.Prefix, p.LocalName)
}

// axisPredicate creates a predicate to predicating for this axis node.
func axisPredicate(root *axisNode) (*TypePredicate, *NamePredicate) {
	// get current axix node type.
	var typ *TypePredicate = &TypePredicate{}
	var nam *NamePredicate = &NamePredicate{root.Prefix, root.LocalName}
	var allNode bool = false
	switch root.AxeType {
	case "attribute":
		*typ = append(*typ, AttributeNode)
	case "self", "parent":
		allNode = true
	default:
		switch root.Prop {
		case "comment":
			*typ = append(*typ, CommentNode)
		case "text":
			allNode = true
			//*typ = append(*typ, TextNode)
			//	case "processing-instruction":
		//	typ = ProcessingInstructionNode
		case "node":
			allNode = true
		}
	}
	if len(*typ) == 0 && !allNode {
		*typ = append(*typ, ElementNode)
	}
	return typ, nam
}

// processAxisNode processes a query for the XPath axis node.
func (b *builder) processAxisNode(root *axisNode) (query, error) {
	var (
		err                error
		qyInput            query
		qyOutput           query
		typePred, namePred = axisPredicate(root)
		predicate          = &AndPredicate{typePred, namePred}
	)

	if root.Input == nil {
		qyInput = &contextQuery{}
	} else {
		if root.AxeType == "child" && (root.Input.Type() == nodeAxis) {
			if input := root.Input.(*axisNode); input.AxeType == "descendant-or-self" {
				var qyGrandInput query
				if input.Input != nil {
					qyGrandInput, _ = b.processNode(input.Input)
				} else {
					qyGrandInput = &contextQuery{}
				}
				qyOutput = &descendantQuery{Input: qyGrandInput, Predicate: predicate, Self: true}
				return qyOutput, nil
			}
		}
		qyInput, err = b.processNode(root.Input)
		if err != nil {
			return nil, err
		}
	}

	switch root.AxeType {
	case "ancestor":
		qyOutput = &ancestorQuery{Input: qyInput, Predicate: predicate}
	case "ancestor-or-self":
		qyOutput = &ancestorQuery{Input: qyInput, Predicate: predicate, Self: true}
	case "attribute":
		qyOutput = &attributeQuery{Input: qyInput, Predicate: predicate}
	case "child":
		filter := func(n NodeNavigator) bool {
			v := namePred.Test(n)
			switch root.Prop {
			case "text":
				v = v && n.NodeType() == TextNode
			case "node":
				v = v && (n.NodeType() == ElementNode || n.NodeType() == TextNode)
			case "comment":
				v = v && n.NodeType() == CommentNode
			default:
				return predicate.Test(n)
			}
			return v
		}
		qyOutput = &childQuery{Input: qyInput, Predicate: PredicateFunc(filter)}
	case "descendant":
		qyOutput = &descendantQuery{Input: qyInput, Predicate: predicate}
	case "descendant-or-self":
		qyOutput = &descendantQuery{Input: qyInput, Predicate: predicate, Self: true}
	case "following":
		qyOutput = &followingQuery{Input: qyInput, Predicate: predicate}
	case "following-sibling":
		qyOutput = &followingQuery{Input: qyInput, Predicate: predicate, Sibling: true}
	case "parent":
		qyOutput = &parentQuery{Input: qyInput, Predicate: predicate}
	case "preceding":
		qyOutput = &precedingQuery{Input: qyInput, Predicate: predicate}
	case "preceding-sibling":
		qyOutput = &precedingQuery{Input: qyInput, Predicate: predicate, Sibling: true}
	case "self":
		qyOutput = &selfQuery{Input: qyInput, Predicate: predicate}
	case "namespace":
		// haha,what will you do someting??
	default:
		err = fmt.Errorf("unknown axe type: %s", root.AxeType)
		return nil, err
	}
	return qyOutput, nil
}

// processFilterNode builds query for the XPath filter predicate.
func (b *builder) processFilterNode(root *filterNode) (query, error) {
	b.flag |= filterFlag

	qyInput, err := b.processNode(root.Input)
	if err != nil {
		return nil, err
	}
	qyCond, err := b.processNode(root.Condition)
	if err != nil {
		return nil, err
	}
	qyOutput := &filterQuery{Input: qyInput, Predicate: qyCond}
	return qyOutput, nil
}

// processFunctionNode processes query for the XPath function node.
func (b *builder) processFunctionNode(root *functionNode) (query, error) {
	var qyOutput query
	switch root.FuncName {
	case "starts-with":
		arg1, err := b.processNode(root.Args[0])
		if err != nil {
			return nil, err
		}
		arg2, err := b.processNode(root.Args[1])
		if err != nil {
			return nil, err
		}
		qyOutput = &functionQuery{Input: b.firstInput, Func: startwithFunc(arg1, arg2)}
	case "contains":
		arg1, err := b.processNode(root.Args[0])
		if err != nil {
			return nil, err
		}
		arg2, err := b.processNode(root.Args[1])
		if err != nil {
			return nil, err
		}

		qyOutput = &functionQuery{Input: b.firstInput, Func: containsFunc(arg1, arg2)}
	case "substring":
		//substring( string , start [, length] )
		if len(root.Args) < 2 {
			return nil, errors.New("xpath: substring function must have at least two parameter")
		}
		var (
			arg1, arg2, arg3 query
			err              error
		)
		if arg1, err = b.processNode(root.Args[0]); err != nil {
			return nil, err
		}
		if arg2, err = b.processNode(root.Args[1]); err != nil {
			return nil, err
		}
		if len(root.Args) == 3 {
			if arg3, err = b.processNode(root.Args[2]); err != nil {
				return nil, err
			}
		}
		qyOutput = &functionQuery{Input: b.firstInput, Func: substringFunc(arg1, arg2, arg3)}
	case "string-length":
		// string-length( [string] )
		if len(root.Args) < 1 {
			return nil, errors.New("xpath: string-length function must have at least one parameter")
		}
		arg1, err := b.processNode(root.Args[0])
		if err != nil {
			return nil, err
		}
		qyOutput = &functionQuery{Input: b.firstInput, Func: stringLengthFunc(arg1)}
	case "normalize-space":
		if len(root.Args) == 0 {
			return nil, errors.New("xpath: normalize-space function must have at least one parameter")
		}
		argQuery, err := b.processNode(root.Args[0])
		if err != nil {
			return nil, err
		}
		qyOutput = &functionQuery{Input: argQuery, Func: normalizespaceFunc}
	case "not":
		if len(root.Args) == 0 {
			return nil, errors.New("xpath: not function must have at least one parameter")
		}
		argQuery, err := b.processNode(root.Args[0])
		if err != nil {
			return nil, err
		}
		qyOutput = &functionQuery{Input: argQuery, Func: notFunc}
	case "name":
		qyOutput = &functionQuery{Input: b.firstInput, Func: nameFunc}
	case "last":
		qyOutput = &functionQuery{Input: b.firstInput, Func: lastFunc}
	case "position":
		qyOutput = &functionQuery{Input: b.firstInput, Func: positionFunc}
	case "count":
		//if b.firstInput == nil {
		//	return nil, errors.New("xpath: expression must evaluate to node-set")
		//}
		if len(root.Args) == 0 {
			return nil, fmt.Errorf("xpath: count(node-sets) function must with have parameters node-sets")
		}
		argQuery, err := b.processNode(root.Args[0])
		if err != nil {
			return nil, err
		}
		qyOutput = &functionQuery{Input: argQuery, Func: countFunc}
	case "sum":
		if len(root.Args) == 0 {
			return nil, fmt.Errorf("xpath: sum(node-sets) function must with have parameters node-sets")
		}
		argQuery, err := b.processNode(root.Args[0])
		if err != nil {
			return nil, err
		}
		qyOutput = &functionQuery{Input: argQuery, Func: sumFunc}
	case "concat":
		if len(root.Args) < 2 {
			return nil, fmt.Errorf("xpath: concat() must have at least two arguments")
		}
		var args []query
		for _, v := range root.Args {
			q, err := b.processNode(v)
			if err != nil {
				return nil, err
			}
			args = append(args, q)
		}
		qyOutput = &functionQuery{Input: b.firstInput, Func: concatFunc(args...)}
	default:
		return nil, fmt.Errorf("not yet support this function %s()", root.FuncName)
	}
	return qyOutput, nil
}

func (b *builder) processOperatorNode(root *operatorNode) (query, error) {
	left, err := b.processNode(root.Left)
	if err != nil {
		return nil, err
	}
	right, err := b.processNode(root.Right)
	if err != nil {
		return nil, err
	}
	var qyOutput query
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
		qyOutput = &numericQuery{Left: left, Right: right, Do: exprFunc}
	case "=", ">", ">=", "<", "<=", "!=":
		var exprFunc func(iterator, interface{}, interface{}) interface{}
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
			exprFunc = neFunc
		}
		qyOutput = &logicalQuery{Left: left, Right: right, Do: exprFunc}
	case "or", "and", "|":
		isOr := false
		if root.Op == "or" || root.Op == "|" {
			isOr = true
		}
		qyOutput = &booleanQuery{Left: left, Right: right, IsOr: isOr}
	}
	return qyOutput, nil
}

func (b *builder) processNode(root node) (q query, err error) {
	if b.depth = b.depth + 1; b.depth > 1024 {
		err = errors.New("the xpath expressions is too complex")
		return
	}

	switch root.Type() {
	case nodeConstantOperand:
		n := root.(*operandNode)
		q = &constantQuery{Val: n.Val}
	case nodeRoot:
		q = &contextQuery{Root: true}
	case nodeAxis:
		q, err = b.processAxisNode(root.(*axisNode))
		b.firstInput = q
	case nodeFilter:
		q, err = b.processFilterNode(root.(*filterNode))
	case nodeFunction:
		q, err = b.processFunctionNode(root.(*functionNode))
	case nodeOperator:
		q, err = b.processOperatorNode(root.(*operatorNode))
	}
	return
}

// build builds a specified XPath expressions expr.
func build(expr string) (q query, err error) {
	defer func() {
		if e := recover(); e != nil {
			switch x := e.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("unknown panic")
			}
		}
	}()
	root := parse(expr)
	b := &builder{}
	return b.processNode(root)
}
