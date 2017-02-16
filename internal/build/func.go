package build

import (
	"errors"
	"strings"

	"github.com/antchfx/gxpath/internal/query"
	"github.com/antchfx/gxpath/xpath"
)

func predicate(q query.Query) func(xpath.NodeNavigator) bool {
	type Predicater interface {
		Test(xpath.NodeNavigator) bool
	}
	if p, ok := q.(Predicater); ok {
		return p.Test
	}
	return func(xpath.NodeNavigator) bool { return true }
}

// positionFunc is a XPath Node Set functions postion().
var positionFunc = func(q query.Query, t query.Iterator) interface{} {
	var (
		count = 1
		node  = t.Current()
	)
	test := predicate(q)
	for node.MoveToPrevious() {
		if test(node) {
			count++
		}
	}
	return float64(count)
}

// lastFunc is a XPath Node Set functions last().
var lastFunc = func(q query.Query, t query.Iterator) interface{} {
	var (
		count = 0
		node  = t.Current()
	)
	node.MoveToFirst()
	test := predicate(q)
	for {
		if test(node) {
			count++
		}
		if !node.MoveToNext() {
			break
		}
	}
	return float64(count)
}

// countFunc is a XPath Node Set functions count(node-set).
var countFunc = func(q query.Query, t query.Iterator) interface{} {
	var (
		count = 0
		node  = t.Current()
	)
	node.MoveToFirst()
	test := predicate(q)
	for {
		if test(node) {
			count++
		}
		if !node.MoveToNext() {
			break
		}
	}
	return float64(count)
}

// nameFunc is a XPath functions name([node-set]).
var nameFunc = func(q query.Query, t query.Iterator) interface{} {
	return t.Current().LocalName()
}

// startwithFunc is a XPath functions starts-with(string, string).
type startwithFunc struct {
	arg1, arg2 query.Query
}

func (s *startwithFunc) do(q query.Query, t query.Iterator) interface{} {
	var (
		m, n string
		ok   bool
	)
	switch typ := s.arg1.Evaluate(t).(type) {
	case string:
		m = typ
	case query.Query:
		node := typ.Select(t)
		if node == nil {
			return false
		}
		m = node.Value()
	default:
		panic(errors.New("starts-with() function argument type must be string"))
	}
	n, ok = s.arg2.Evaluate(t).(string)
	if !ok {
		panic(errors.New("starts-with() function argument type must be string"))
	}
	return strings.HasPrefix(m, n)
}

// normalizespaceFunc is XPath functions normalize-space(string?)
var normalizespaceFunc = func(q query.Query, t query.Iterator) interface{} {
	var m string
	switch typ := q.Evaluate(t).(type) {
	case string:
		m = typ
	case query.Query:
		node := typ.Select(t)
		if node == nil {
			return false
		}
		m = node.Value()
	}
	return strings.TrimSpace(m)
}

// substringFunc is XPath functions substring function returns a part of a given string.
type substringFunc struct {
	arg1, arg2, arg3 query.Query
}

func (f *substringFunc) do(q query.Query, t query.Iterator) interface{} {
	var m string
	switch typ := f.arg1.Evaluate(t).(type) {
	case string:
		m = typ
	case query.Query:
		node := typ.Select(t)
		if node == nil {
			return false
		}
		m = node.Value()
	}

	var start, length float64
	var ok bool

	if start, ok = f.arg2.Evaluate(t).(float64); !ok {
		panic(errors.New("substring() function first argument type must be int"))
	}
	if f.arg3 != nil {
		if length, ok = f.arg3.Evaluate(t).(float64); !ok {
			panic(errors.New("substring() function second argument type must be int"))
		}
	}
	if (len(m) - int(start)) < int(length) {
		panic(errors.New("substring() function start and length argument out of range"))
	}
	if length > 0 {
		return m[int(start):int(length+start)]
	}
	return m[int(start):]
}

// notFunc is XPATH functions not(expression) function operation.
var notFunc = func(q query.Query, t query.Iterator) interface{} {
	switch v := q.Evaluate(t).(type) {
	case bool:
		return !v
	case query.Query:
		node := v.Select(t)
		return node == nil
	default:
		return false
	}
}
