package parse

import (
	"bytes"
	"fmt"
)

// A Node is an element in the parse tree.
type Node interface {
	Type() NodeType
	String() string
}

// NodeType identifies the type of a parse tree node.
type NodeType int

func (t NodeType) Type() NodeType {
	return t
}

const (
	NodeRoot NodeType = iota
	NodeAxis
	NodeFilter
	NodeFunction
	NodeOperator
	NodeVariable
	NodeConstantOperand
)

// RootNode holds a top-level node of tree.
type RootNode struct {
	NodeType
	slash string
}

func (r *RootNode) String() string {
	return r.slash
}

// OperatorNode holds two Nodes operator.
type OperatorNode struct {
	NodeType
	Op    string
	Left  Node
	Right Node
}

func (o *OperatorNode) String() string {
	return fmt.Sprintf("%v%s%v", o.Left, o.Op, o.Right)
}

// AxisNode holds a location step.
type AxisNode struct {
	NodeType
	Input     Node
	Prop      string // node-test name.[comment|text|processing-instruction|node]
	AxeType   string // name of the axes.[attribute|ancestor|child|....]
	LocalName string // local part name of node.
	Prefix    string // prefix name of node.
}

func (a *AxisNode) String() string {
	var b bytes.Buffer
	if a.AxeType != "" {
		b.Write([]byte(a.AxeType + "::"))
	}
	if a.Prefix != "" {
		b.Write([]byte(a.Prefix + ":"))
	}
	b.Write([]byte(a.LocalName))
	if a.Prop != "" {
		b.Write([]byte("/" + a.Prop + "()"))
	}
	return b.String()
}

// OperandNode holds a constant operand.
type OperandNode struct {
	NodeType
	Val interface{}
}

func (o *OperandNode) String() string {
	return fmt.Sprintf("%v", o.Val)
}

// FilterNode holds a condition filter.
type FilterNode struct {
	NodeType
	Input, Condition Node
}

func (f *FilterNode) String() string {
	return fmt.Sprintf("%s[%s]", f.Input, f.Condition)
}

// VariableNode holds a variable.
type VariableNode struct {
	NodeType
	Name, Prefix string
}

func (v *VariableNode) String() string {
	if v.Prefix == "" {
		return v.Name
	}
	return fmt.Sprintf("%s:%s", v.Prefix, v.Name)
}

// FunctionNode holds a function call.
type FunctionNode struct {
	NodeType
	Args     []Node
	Prefix   string
	FuncName string // function name
}

func (f *FunctionNode) String() string {
	var b bytes.Buffer
	// fun(arg1, ..., argn)
	b.Write([]byte(f.FuncName))
	b.Write([]byte("("))
	for i, arg := range f.Args {
		if i > 0 {
			b.Write([]byte(","))
		}
		b.Write([]byte(fmt.Sprintf("%s", arg)))
	}
	b.Write([]byte(")"))
	return b.String()
}
