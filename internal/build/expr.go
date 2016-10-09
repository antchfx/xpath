package build

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/antchfx/gxpath/internal/query"
)

// valueType is a return value type.
type valueType int

const (
	booleanType valueType = iota
	numberType
	stringType
	nodeSetType
)

func getValueType(i interface{}) valueType {
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Float64:
		return numberType
	case reflect.String:
		return stringType
	case reflect.Bool:
		return booleanType
	default:
		if _, ok := i.(query.Query); ok {
			return nodeSetType
		}
	}
	panic(fmt.Errorf("xpath unknown value type: %v", v.Kind()))
}

type logical func(query.Iterator, string, interface{}, interface{}) bool

var logicalFuncs = [][]logical{
	[]logical{cmpBoolean_Boolean, nil, nil, nil},
	[]logical{nil, cmpNumeric_Numeric, cmpNumeric_String, cmpNumeric_NodeSet},
	[]logical{nil, cmpString_Numeric, cmpString_String, cmpString_NodeSet},
	[]logical{nil, cmpNodeSet_Numeric, cmpNodeSet_String, cmpNodeSet_NodeSet},
}

// number vs number
func cmpNumberNumberF(op string, a, b float64) bool {
	switch op {
	case "=":
		return a == b
	case ">":
		return a > b
	case "<":
		return a < b
	case ">=":
		return a >= b
	case "<=":
		return a <= b
	case "!=":
		return a != b
	}
	return false
}

// string vs string
func cmpStringStringF(op string, a, b string) bool {
	switch op {
	case "=":
		return a == b
	case ">":
		return a > b
	case "<":
		return a < b
	case ">=":
		return a >= b
	case "<=":
		return a <= b
	case "!=":
		return a != b
	}
	return false
}

func cmpBooleanBooleanF(op string, a, b bool) bool {
	switch op {
	case "or":
		return a || b
	case "and":
		return a && b
	}
	return false
}

func cmpNumeric_Numeric(t query.Iterator, op string, m, n interface{}) bool {
	a := m.(float64)
	b := m.(float64)
	return cmpNumberNumberF(op, a, b)
}

func cmpNumeric_String(t query.Iterator, op string, m, n interface{}) bool {
	a := m.(float64)
	b := m.(string)
	num, err := strconv.ParseFloat(b, 64)
	if err != nil {
		panic(err)
	}
	return cmpNumberNumberF(op, a, num)
}

func cmpNumeric_NodeSet(t query.Iterator, op string, m, n interface{}) bool {
	a := m.(float64)
	b := n.(query.Query)

	for {
		node := b.Select(t)
		if node == nil {
			break
		}
		num, err := strconv.ParseFloat(node.Value(), 64)
		if err != nil {
			panic(err)
		}
		if cmpNumberNumberF(op, a, num) {
			return true
		}
	}
	return false
}

func cmpNodeSet_Numeric(t query.Iterator, op string, m, n interface{}) bool {
	a := m.(query.Query)
	b := n.(float64)
	for {
		node := a.Select(t)
		if node == nil {
			break
		}
		num, err := strconv.ParseFloat(node.Value(), 64)
		if err != nil {
			panic(err)
		}
		if cmpNumberNumberF(op, num, b) {
			return true
		}
	}
	return false
}

func cmpNodeSet_String(t query.Iterator, op string, m, n interface{}) bool {
	a := m.(query.Query)
	b := n.(string)
	for {
		node := a.Select(t)
		if node == nil {
			break
		}
		if cmpStringStringF(op, b, node.Value()) {
			return true
		}
	}
	return false
}

func cmpNodeSet_NodeSet(t query.Iterator, op string, m, n interface{}) bool {
	return false
}

func cmpString_Numeric(t query.Iterator, op string, m, n interface{}) bool {
	a := m.(string)
	b := n.(float64)
	num, err := strconv.ParseFloat(a, 64)
	if err != nil {
		panic(err)
	}
	return cmpNumberNumberF(op, b, num)
}

func cmpString_String(t query.Iterator, op string, m, n interface{}) bool {
	a := m.(string)
	b := n.(string)
	return cmpStringStringF(op, a, b)
}

func cmpString_NodeSet(t query.Iterator, op string, m, n interface{}) bool {
	a := m.(string)
	b := n.(query.Query)
	for {
		node := b.Select(t)
		if node == nil {
			break
		}
		if cmpStringStringF(op, a, node.Value()) {
			return true
		}
	}
	return false
}

func cmpBoolean_Boolean(t query.Iterator, op string, m, n interface{}) bool {
	a := m.(bool)
	b := m.(bool)
	return cmpBooleanBooleanF(op, a, b)
}

// eqFunc is an `=` operator.
func eqFunc(t query.Iterator, m, n interface{}) interface{} {
	t1 := getValueType(m)
	t2 := getValueType(n)
	return logicalFuncs[t1][t2](t, "=", m, n)
}

// gtFunc is an `>` operator.
func gtFunc(t query.Iterator, m, n interface{}) interface{} {
	t1 := getValueType(m)
	t2 := getValueType(n)
	return logicalFuncs[t1][t2](t, ">", m, n)
}

// geFunc is an `>=` operator.
func geFunc(t query.Iterator, m, n interface{}) interface{} {
	t1 := getValueType(m)
	t2 := getValueType(n)
	return logicalFuncs[t1][t2](t, ">=", m, n)
}

// ltFunc is an `<` operator.
func ltFunc(t query.Iterator, m, n interface{}) interface{} {
	t1 := getValueType(m)
	t2 := getValueType(n)
	return logicalFuncs[t1][t2](t, "<", m, n)
}

// leFunc is an `<=` operator.
func leFunc(t query.Iterator, m, n interface{}) interface{} {
	t1 := getValueType(m)
	t2 := getValueType(n)
	return logicalFuncs[t1][t2](t, "<=", m, n)
}

// neFunc is an `!=` operator.
func neFunc(t query.Iterator, m, n interface{}) interface{} {
	t1 := getValueType(m)
	t2 := getValueType(n)
	return logicalFuncs[t1][t2](t, "!=", m, n)
}

// orFunc is an `or` operator.
var orFunc = func(t query.Iterator, m, n interface{}) interface{} {
	t1 := getValueType(m)
	t2 := getValueType(n)
	return logicalFuncs[t1][t2](t, "or", m, n)
}

func numericExpr(m, n interface{}, cb func(float64, float64) float64) float64 {
	typ := reflect.TypeOf(float64(0))
	a := reflect.ValueOf(m).Convert(typ)
	b := reflect.ValueOf(n).Convert(typ)
	return cb(a.Float(), b.Float())
}

// plusFunc is an `+` operator.
var plusFunc = func(m, n interface{}) interface{} {
	return numericExpr(m, n, func(a, b float64) float64 {
		return a + b
	})
}

// minusFunc is an `-` operator.
var minusFunc = func(m, n interface{}) interface{} {
	return numericExpr(m, n, func(a, b float64) float64 {
		return a - b
	})
}

// mulFunc is an `*` operator.
var mulFunc = func(m, n interface{}) interface{} {
	return numericExpr(m, n, func(a, b float64) float64 {
		return a * b
	})
}

// divFunc is an `DIV` operator.
var divFunc = func(m, n interface{}) interface{} {
	return numericExpr(m, n, func(a, b float64) float64 {
		return a / b
	})
}

// modFunc is an 'MOD' operator.
var modFunc = func(m, n interface{}) interface{} {
	return numericExpr(m, n, func(a, b float64) float64 {
		return float64(int(a) % int(b))
	})
}
