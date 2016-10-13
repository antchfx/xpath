package query

import (
	"reflect"

	"github.com/antchfx/gxpath/xpath"
)

// An XPath query interface.
type Query interface {
	// Select traversing Iterator returns a query matched node xpath.NodeNavigator.
	Select(Iterator) xpath.NodeNavigator

	// Evaluate evaluates query and returns values of the current query.
	Evaluate(Iterator) interface{}

	// Test checks a specified xpath.NodeNavigator can passed by the current query.
	//Test(xpath.NodeNavigator) bool
}

// ContextQuery is returns current node on the Iterator object query.
type ContextQuery struct {
	count int
	Root  bool // Moving to root-level node in the current context Iterator.
}

func (c *ContextQuery) Select(t Iterator) (n xpath.NodeNavigator) {
	if c.count == 0 {
		c.count++
		n = t.Current().Copy()
		if c.Root {
			n.MoveToRoot()
		}
	}
	return n
}

func (c *ContextQuery) Evaluate(Iterator) interface{} {
	c.count = 0
	return c
}

// AncestorQuery is an XPath ancestor node query.(ancestor::*|ancestor-self::*)
type AncestorQuery struct {
	iterator func() xpath.NodeNavigator

	Self      bool
	Input     Query
	Predicate func(xpath.NodeNavigator) bool
}

func (a *AncestorQuery) Select(t Iterator) xpath.NodeNavigator {
	for {
		if a.iterator == nil {
			node := a.Input.Select(t)
			if node == nil {
				return nil
			}
			first := true
			a.iterator = func() xpath.NodeNavigator {
				if first && a.Self {
					first = false
					if a.Predicate(node) {
						return node
					}
				}
				for node.MoveToParent() {
					if !a.Predicate(node) {
						break
					}
					return node
				}
				return nil
			}
		}

		if node := a.iterator(); node != nil {
			return node
		}
		a.iterator = nil
	}
}

func (a *AncestorQuery) Evaluate(t Iterator) interface{} {
	a.Input.Evaluate(t)
	return a
}

func (a *AncestorQuery) Test(n xpath.NodeNavigator) bool {
	return a.Predicate(n)
}

// AttributeQuery is an XPath attribute node query.(@*)
type AttributeQuery struct {
	iterator func() xpath.NodeNavigator

	Input     Query
	Predicate func(xpath.NodeNavigator) bool
}

func (a *AttributeQuery) Select(t Iterator) xpath.NodeNavigator {
	for {
		if a.iterator == nil {
			node := a.Input.Select(t)
			if node == nil {
				return nil
			}
			node = node.Copy()
			a.iterator = func() xpath.NodeNavigator {
				for {
					onAttr := node.MoveToNextAttribute()
					if !onAttr {
						return nil
					}
					if a.Predicate(node) {
						return node
					}
				}
			}
		}

		if node := a.iterator(); node != nil {
			return node
		}
		a.iterator = nil
	}
}

func (a *AttributeQuery) Evaluate(t Iterator) interface{} {
	a.Input.Evaluate(t)
	return a
}

func (a *AttributeQuery) Test(n xpath.NodeNavigator) bool {
	return a.Predicate(n)
}

// ChildQuery is an XPath child node query.(child::*)
type ChildQuery struct {
	posit    int
	iterator func() xpath.NodeNavigator

	Input     Query
	Predicate func(xpath.NodeNavigator) bool
}

func (c *ChildQuery) Select(t Iterator) xpath.NodeNavigator {
	for {
		if c.iterator == nil {
			c.posit = 0
			node := c.Input.Select(t)
			if node == nil {
				return nil
			}
			node = node.Copy()
			first := true
			c.iterator = func() xpath.NodeNavigator {
				for {
					if (first && !node.MoveToChild()) || (!first && !node.MoveToNext()) {
						return nil
					}
					first = false
					if c.Predicate(node) {
						return node
					}
				}
			}
		}

		if node := c.iterator(); node != nil {
			c.posit++
			return node
		}
		c.iterator = nil
	}
}

func (c *ChildQuery) Evaluate(t Iterator) interface{} {
	c.Input.Evaluate(t)
	c.iterator = nil
	return c
}

func (c *ChildQuery) Test(n xpath.NodeNavigator) bool {
	return c.Predicate(n)
}

// position returns a position of current xpath.NodeNavigator.
func (c *ChildQuery) position() int {
	return c.posit
}

// DescendantQuery is an XPath descendant node query.(descendant::* | descendant-or-self::*)
type DescendantQuery struct {
	iterator func() xpath.NodeNavigator

	Self      bool
	Input     Query
	Predicate func(xpath.NodeNavigator) bool
}

func (d *DescendantQuery) Select(t Iterator) xpath.NodeNavigator {
	for {
		if d.iterator == nil {
			node := d.Input.Select(t)
			if node == nil {
				return nil
			}
			node = node.Copy()
			level := 0
			first := true
			d.iterator = func() xpath.NodeNavigator {
				if first && d.Self {
					first = false
					if d.Predicate(node) {
						return node
					}
				}

				for {
					if node.MoveToChild() {
						level++
					} else {
						for {
							if level == 0 {
								return nil
							}
							if node.MoveToNext() {
								break
							}
							node.MoveToParent()
							level--
						}
					}
					if d.Predicate(node) {
						return node
					}
				}
			}
		}

		if node := d.iterator(); node != nil {
			return node
		}
		d.iterator = nil
	}
}

func (d *DescendantQuery) Evaluate(t Iterator) interface{} {
	d.Input.Evaluate(t)
	d.iterator = nil
	return d
}

func (d *DescendantQuery) Test(n xpath.NodeNavigator) bool {
	return d.Predicate(n)
}

// FollowingQuery is an XPath following node query.(following::*|following-sibling::*)
type FollowingQuery struct {
	iterator func() xpath.NodeNavigator

	Input     Query
	Sibling   bool // The matching sibling node of current node.
	Predicate func(xpath.NodeNavigator) bool
}

func (f *FollowingQuery) Select(t Iterator) xpath.NodeNavigator {
	for {
		if f.iterator == nil {
			node := f.Input.Select(t)
			if node == nil {
				return nil
			}
			node = node.Copy()
			if f.Sibling {
				f.iterator = func() xpath.NodeNavigator {
					for {
						if !node.MoveToNext() {
							return nil
						}
						if f.Predicate(node) {
							return node
						}
					}
				}
			} else {
				var q Query // descendant query
				f.iterator = func() xpath.NodeNavigator {
					for {
						if q == nil {
							for !node.MoveToNext() {
								if !node.MoveToParent() {
									return nil
								}
							}
							q = &DescendantQuery{
								Self:      true,
								Input:     &ContextQuery{},
								Predicate: f.Predicate,
							}
							t.Current().MoveTo(node)
						}
						if node := q.Select(t); node != nil {
							return node
						}
						q = nil
					}
				}
			}
		}

		if node := f.iterator(); node != nil {
			return node
		}
		f.iterator = nil
	}
}

func (f *FollowingQuery) Evaluate(t Iterator) interface{} {
	f.Input.Evaluate(t)
	return f
}

func (f *FollowingQuery) Test(n xpath.NodeNavigator) bool {
	return f.Predicate(n)
}

// PrecedingQuery is an XPath preceding node query.(preceding::*)
type PrecedingQuery struct {
	iterator  func() xpath.NodeNavigator
	Input     Query
	Sibling   bool // The matching sibling node of current node.
	Predicate func(xpath.NodeNavigator) bool
}

func (p *PrecedingQuery) Select(t Iterator) xpath.NodeNavigator {
	for {
		if p.iterator == nil {
			node := p.Input.Select(t)
			if node == nil {
				return nil
			}
			node = node.Copy()
			if p.Sibling {
				p.iterator = func() xpath.NodeNavigator {
					for {
						for !node.MoveToPrevious() {
							return nil
						}
						if p.Predicate(node) {
							return node
						}
					}
				}
			} else {
				var q Query
				p.iterator = func() xpath.NodeNavigator {
					for {
						if q == nil {
							for !node.MoveToPrevious() {
								if !node.MoveToParent() {
									return nil
								}
							}
							q = &DescendantQuery{
								Self:      true,
								Input:     &ContextQuery{},
								Predicate: p.Predicate,
							}
							t.Current().MoveTo(node)
						}
						if node := q.Select(t); node != nil {
							return node
						}
						q = nil
					}
				}
			}
		}
		if node := p.iterator(); node != nil {
			return node
		}
		p.iterator = nil
	}
}

func (p *PrecedingQuery) Evaluate(t Iterator) interface{} {
	p.Input.Evaluate(t)
	return p
}

func (p *PrecedingQuery) Test(n xpath.NodeNavigator) bool {
	return p.Predicate(n)
}

// ParentQuery is an XPath parent node query.(parent::*)
type ParentQuery struct {
	Input     Query
	Predicate func(xpath.NodeNavigator) bool
}

func (p *ParentQuery) Select(t Iterator) xpath.NodeNavigator {
	for {
		node := p.Input.Select(t)
		if node == nil {
			return nil
		}
		node = node.Copy()
		if node.MoveToParent() && p.Predicate(node) {
			return node
		}
	}
}

func (p *ParentQuery) Evaluate(t Iterator) interface{} {
	p.Input.Evaluate(t)
	return p
}

func (p *ParentQuery) Test(n xpath.NodeNavigator) bool {
	return p.Predicate(n)
}

// SelfQuery is an Self node query.(self::*)
type SelfQuery struct {
	Input     Query
	Predicate func(xpath.NodeNavigator) bool
}

func (s *SelfQuery) Select(t Iterator) xpath.NodeNavigator {
	for {
		node := s.Input.Select(t)
		if node == nil {
			return nil
		}

		if s.Predicate(node) {
			return node
		}
	}
}

func (s *SelfQuery) Evaluate(t Iterator) interface{} {
	s.Input.Evaluate(t)
	return s
}

func (s *SelfQuery) Test(n xpath.NodeNavigator) bool {
	return s.Predicate(n)
}

// FilterQuery is an XPath query for predicate filter.
type FilterQuery struct {
	Input     Query
	Predicate Query
}

func (f *FilterQuery) do(t Iterator) bool {
	val := reflect.ValueOf(f.Predicate.Evaluate(t))
	switch val.Kind() {
	case reflect.Bool:
		return val.Bool()
	case reflect.String:
		return len(val.String()) > 0
	case reflect.Float64:
		pt := float64(getNodePosition(f.Input))
		return val.Float() == pt
	default:
		if q, ok := f.Predicate.(Query); ok {
			return q.Select(t) != nil
		}
	}
	return false
}

func (f *FilterQuery) Select(t Iterator) xpath.NodeNavigator {
	for {
		node := f.Input.Select(t)
		if node == nil {
			return node
		}
		node = node.Copy()
		//fmt.Println(node.LocalName())

		t.Current().MoveTo(node)
		if f.do(t) {
			return node
		}
	}
}

func (f *FilterQuery) Evaluate(t Iterator) interface{} {
	f.Input.Evaluate(t)
	return f
}

// FunctionQuery is an XPath function that call a function to returns
// value of current xpath.NodeNavigator node.
type XPathFunction struct {
	Input Query                             // Node Set
	Func  func(Query, Iterator) interface{} // The xpath function.
}

func (f *XPathFunction) Select(t Iterator) xpath.NodeNavigator {
	return nil
}

// Evaluate call a specified function that will returns the
// following value type: number,string,boolean.
func (f *XPathFunction) Evaluate(t Iterator) interface{} {
	return f.Func(f.Input, t)
}

// XPathConstant is an XPath constant operand.
type XPathConstant struct {
	Val interface{}
}

func (c *XPathConstant) Select(t Iterator) xpath.NodeNavigator {
	return nil
}

func (c *XPathConstant) Evaluate(t Iterator) interface{} {
	return c.Val
}

// LogicalExpr is an XPath logical expression.
type LogicalExpr struct {
	Left, Right Query

	Do func(Iterator, interface{}, interface{}) interface{}
}

func (l *LogicalExpr) Select(t Iterator) xpath.NodeNavigator {
	// When a XPath expr is logical expression.
	node := t.Current().Copy()
	val := l.Evaluate(t)
	switch val.(type) {
	case bool:
		if val.(bool) == true {
			return node
		}
	}
	return nil
}

func (l *LogicalExpr) Evaluate(t Iterator) interface{} {
	m := l.Left.Evaluate(t)
	n := l.Right.Evaluate(t)
	return l.Do(t, m, n)
}

// NumericExpr is an XPath numeric operator expression.
type NumericExpr struct {
	Left, Right Query

	Do func(interface{}, interface{}) interface{}
}

func (n *NumericExpr) Select(t Iterator) xpath.NodeNavigator {
	return nil
}

func (n *NumericExpr) Evaluate(t Iterator) interface{} {
	m := n.Left.Evaluate(t)
	k := n.Right.Evaluate(t)
	return n.Do(m, k)
}

type BooleanExpr struct {
	IsOr        bool
	Left, Right Query
}

func (b *BooleanExpr) Select(t Iterator) xpath.NodeNavigator {
	return nil
}

func (b *BooleanExpr) Evaluate(t Iterator) interface{} {
	m := b.Left.Evaluate(t)
	if m.(bool) == b.IsOr {
		return m
	}
	return b.Right.Evaluate(t)
}

func getNodePosition(q Query) int {
	type Position interface {
		position() int
	}
	if count, ok := q.(Position); ok {
		return count.position()
	}
	return 1
}
