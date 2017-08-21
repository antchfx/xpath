package xmlquery

import (
	"github.com/antchfx/xpath"
	"github.com/antchfx/xpath/test"
	"github.com/antchfx/xquery/xml"
	"strings"
	"testing"
)

func loadXML(s string) *xmlquery.Node {
	node, err := xmlquery.ParseXML(strings.NewReader(s))
	if err != nil {
		panic(err)
	}
	return node
}

func TestXQuery(t *testing.T) {
	create := func(doc string) (xpath.NodeNavigator, error) {
		return xmlquery.CreateXPathNavigator(loadXML(doc)), nil
	}

	xquerytest.TestAll(t, create, xquerytest.EnableAll)
}
