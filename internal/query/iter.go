package query

import "github.com/antchfx/gxpath/xpath"

type Iterator interface {
	Current() xpath.NodeNavigator
}
