package build

import (
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
		count = 0
		node  = t.Current()
		val   = node.Current()
	)
	node.MoveToFirst()
	test := predicate(q)
	for {
		if test(node) {
			count++
			if node.Current() == val {
				break
			}
		}
		if !node.MoveToNext() {
			break
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
