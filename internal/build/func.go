package build

import "github.com/antchfx/gxpath/internal/query"

// positionFunc is a XPath Node Set functions postion().
var positionFunc = func(q query.Query, t query.Iterator) interface{} {
	count := 0
	node := t.Current()
	curr := node.Current()
	node.MoveToFirst()
	for {
		if q.Test(node) {
			count++
			if node.Current() == curr {
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
	count := 0
	node := t.Current()
	node.MoveToFirst()
	for {
		if q.Test(node) {
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
	count := 0
	node := t.Current()
	node.MoveToFirst()
	for {
		if q.Test(node) {
			count++
		}
		if !node.MoveToNext() {
			break
		}
	}
	return float64(count)
}
