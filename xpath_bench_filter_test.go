package xpath

import (
	"fmt"
	"testing"
)

func buildFilterBenchDoc(numSections, itemsPerSection int) *TNode {
	root := createNode("root", ElementNode)
	for s := 0; s < numSections; s++ {
		section := addChild(root, createNode("section", ElementNode))
		addChild(section, createNode(fmt.Sprintf("s%d", s), AttributeNode))
		for i := 0; i < itemsPerSection; i++ {
			item := addChild(section, createNode("item", ElementNode))
			cls := "a"
			if i%2 == 1 {
				cls = "b"
			}
			addChild(item, createNode(cls, AttributeNode))
			addChild(item, createNode(fmt.Sprintf("i%d", s*itemsPerSection+i), AttributeNode))
		}
	}
	return root
}

func addChild(parent, child *TNode) *TNode {
	child.Parent = parent
	if parent.FirstChild == nil {
		parent.FirstChild = child
	} else {
		last := parent.FirstChild
		for last.NextSibling != nil {
			last = last.NextSibling
		}
		last.NextSibling = child
		child.PrevSibling = last
	}
	return child
}

func BenchmarkFilterNonPositional(b *testing.B) {
	doc := buildFilterBenchDoc(10, 20) // 200 items across 10 sections, ~430 nodes total
	expr := MustCompile(`//item[@class]`)
	nav := createNavigator(doc)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter := expr.Select(nav)
		for iter.MoveNext() {
		}
	}
}

func BenchmarkFilterNonPositionalAttrValue(b *testing.B) {
	doc := buildFilterBenchDoc(10, 20)
	expr := MustCompile(`//item[@class="a"]`)
	nav := createNavigator(doc)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter := expr.Select(nav)
		for iter.MoveNext() {
		}
	}
}

func BenchmarkFilterPositional(b *testing.B) {
	doc := buildFilterBenchDoc(10, 20)
	expr := MustCompile(`//item[1]`)
	nav := createNavigator(doc)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter := expr.Select(nav)
		for iter.MoveNext() {
		}
	}
}

func buildTranslateBenchDoc(numItems int) *TNode {
	root := createNode("root", ElementNode)
	for i := 0; i < numItems; i++ {
		item := addChild(root, createNode("item", ElementNode))
		style := "color:#605e5c"
		if i%3 == 0 {
			style = "color:#FF0000"
		}
		addChild(item, createNode(style, AttributeNode))
	}
	return root
}

func BenchmarkTranslateConst(b *testing.B) {
	doc := buildTranslateBenchDoc(200)
	expr := MustCompile(`//*[contains(translate(@style, 'ABCDEF', 'abcdef'), 'color:#605e5c')]`)
	nav := createNavigator(doc)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter := expr.Select(nav)
		for iter.MoveNext() {
		}
	}
}
