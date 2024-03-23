package xpath

import "testing"

func TestXPathNode_self(t *testing.T) {
	s := `//name/self::*`
	nodes := selectNodes(employee_example, s)
	assertEqual(t, 3, len(nodes))

	expecteds := []struct {
		lines int
		value string
	}{
		{lines: 4, value: "Opal Kole"},
		{lines: 9, value: "Max Miller"},
		{lines: 14, value: "Beccaa Moss"},
	}
	for i := 0; i < 3; i++ {
		n := nodes[i]
		assertEqual(t, "name", n.Data)
		e := expecteds[i]
		assertEqual(t, e.lines, n.lines)
		assertEqual(t, e.value, n.Value())
	}
}

func TestXPathNode_child(t *testing.T) {
	s := `/empinfo/child::*`
	nodes := selectNodes(employee_example, s)
	assertEqual(t, 3, len(nodes))
	expected_1 := []struct {
		lines int
		value string
	}{{lines: 3}, {lines: 8}, {lines: 13}}
	for i := 0; i < 3; i++ {
		n := nodes[i]
		assertEqual(t, "employee", n.Data)
		assertEqual(t, expected_1[i].lines, n.lines)
	}

	s = `/empinfo/child::node()`
	nodes = selectNodes(employee_example, s)
	assertEqual(t, 3, len(nodes))
	for _, n := range nodes {
		assertEqual(t, ElementNode, n.Type)
		assertEqual(t, "employee", n.Data)
	}

	s = `//name/child::text()`
	nodes = selectNodes(employee_example, s)
	assertEqual(t, 3, len(nodes))
	expected_2 := []string{"Opal Kole", "Max Miller", "Beccaa Moss"}
	for i := 0; i < 3; i++ {
		n := nodes[i]
		assertEqual(t, TextNode, n.Type)
		assertEqual(t, expected_2[i], n.Value())
	}

	s = `//child::employee/child::email`
	nodes = selectNodes(employee_example, s)
	assertEqual(t, 3, len(nodes))
	for i := 0; i < 3; i++ {
		n := nodes[i]
		assertEqual(t, "email", n.Data)
	}
}

func TestXPathNode_descendant(t *testing.T) {
	s := `//employee/descendant::*`
	nodes := selectNodes(employee_example, s)
	assertEqual(t, 9, len(nodes))
	expecteds := []struct {
		lines int
		tag   string
	}{
		{4, "name"},
		{5, "designation"},
		{6, "email"},
		{9, "name"},
		{10, "designation"},
		{11, "email"},
		{14, "name"},
		{15, "designation"},
		{16, "email"},
	}
	for i := 0; i < 9; i++ {
		n := nodes[i]
		assertEqual(t, expecteds[i].lines, n.lines)
		assertEqual(t, expecteds[i].tag, n.Data)
	}

	s = `//descendant::employee`
	nodes = selectNodes(employee_example, s)
	assertEqual(t, 3, len(nodes))
	for _, n := range nodes {
		assertEqual(t, "employee", n.Data)
	}
}

func TestXPathNode_descendant_or_self(t *testing.T) {
	s := `//employee/descendant-or-self::*`
	nodes := selectNodes(employee_example, s)
	assertEqual(t, 12, len(nodes))
	expected_1 := []struct {
		lines int
		tag   string
	}{
		{3, "employee"},
		{4, "name"},
		{5, "designation"},
		{6, "email"},
		{8, "employee"},
		{9, "name"},
		{10, "designation"},
		{11, "email"},
		{13, "employee"},
		{14, "name"},
		{15, "designation"},
		{16, "email"},
	}
	for i := 0; i < 12; i++ {
		n := nodes[i]
		assertEqual(t, expected_1[i].lines, n.lines)
		assertEqual(t, expected_1[i].tag, n.Data)
	}

	s = `//descendant-or-self::employee`
	nodes = selectNodes(employee_example, s)
	// Not Passed
	assertEqual(t, 3, len(nodes))
	for _, n := range nodes {
		assertEqual(t, "employee", n.Data)
	}
}

func TestXPathNode_ancestor(t *testing.T) {
	s := `//employee/ancestor::*`
	nodes := selectNodes(employee_example, s)
	assertEqual(t, 1, len(nodes))
	n := nodes[0]
	assertEqual(t, "empinfo", n.Data)

	s = `//ancestor::name`
	nodes = selectNodes(employee_example, s)
	assertEqual(t, 3, len(nodes))
	for _, n := range nodes {
		assertEqual(t, "name", n.Data)
	}
}

func TestXPathNode_ancestor_or_self(t *testing.T) {
	s := `//employee/ancestor-or-self::*`
	nodes := selectNodes(employee_example, s)
	assertEqual(t, 4, len(nodes))

	expecteds := []struct {
		lines int
		tag   string
	}{
		{2, "empinfo"},
		{3, "employee"},
		{8, "employee"},
		{13, "employee"},
	}

	for i := 0; i < 4; i++ {
		n := nodes[i]
		assertEqual(t, expecteds[i].lines, n.lines)
		assertEqual(t, expecteds[i].tag, n.Data)
	}

	s = `//name/ancestor-or-self::employee`
	nodes = selectNodes(employee_example, s)
	assertEqual(t, 3, len(nodes))
	for i := 0; i < 3; i++ {
		n := nodes[i]
		assertEqual(t, expecteds[i+1].lines, n.lines)
		assertEqual(t, expecteds[i+1].tag, n.Data)
	}
}

func TestXPathNode_parent(t *testing.T) {
	s := `//name/parent::*`
	nodes := selectNodes(employee_example, s)
	assertEqual(t, 3, len(nodes))

	expecteds := []struct {
		lines int
		tag   string
	}{
		{3, "employee"},
		{8, "employee"},
		{13, "employee"},
	}
	for i := 0; i < 3; i++ {
		n := nodes[i]
		assertEqual(t, expecteds[i].lines, n.lines)
		assertEqual(t, expecteds[i].tag, n.Data)
	}

	s = `//name/parent::employee`
	nodes = selectNodes(employee_example, s)
	assertEqual(t, 3, len(nodes))
	for i := 0; i < 3; i++ {
		n := nodes[i]
		assertEqual(t, expecteds[i].lines, n.lines)
		assertEqual(t, expecteds[i].tag, n.Data)
	}
}

func TestXPathNode_attribute(t *testing.T) {
	s := `//attribute::id`
	nodes := selectNodes(employee_example, s)
	assertEqual(t, 3, len(nodes))

	expecteds := []string{"1", "2", "3"}
	for i := 0; i < 3; i++ {
		n := nodes[i]
		assertEqual(t, AttributeNode.String(), n.Type.String())
		assertEqual(t, expecteds[i], n.Data)
	}

	s = `//attribute::*`
	nodes = selectNodes(employee_example, s)
	assertEqual(t, 9, len(nodes))
	expected_attributes := []string{"id", "discipline", "experience", "id", "from", "discipline", "experience", "id", "discipline"}
	for i := 0; i < 9; i++ {
		n := nodes[i]
		assertEqual(t, expected_attributes[i], n.Data)
	}
}

func TestXPathNode_following(t *testing.T) {
	s := `//employee[@id=1]/following::*`
	nodes := selectNodes(employee_example, s)
	assertEqual(t, 8, len(nodes))

	expecteds := []struct {
		lines int
		tag   string
	}{
		{8, "employee"},
		{9, "name"},
		{10, "designation"},
		{11, "email"},
		{13, "employee"},
		{14, "name"},
		{15, "designation"},
		{16, "email"},
	}

	for i := 0; i < 8; i++ {
		n := nodes[i]
		assertEqual(t, expecteds[i].lines, n.lines)
		assertEqual(t, expecteds[i].tag, n.Data)
	}
}

func TestXPathNode_following_sibling(t *testing.T) {
	s := `//employee[@id=1]/following-sibling::*`
	nodes := selectNodes(employee_example, s)
	assertEqual(t, 2, len(nodes))

	expecteds := []struct {
		lines int
		tag   string
	}{
		{8, "employee"},
		{13, "employee"},
	}

	for i := 0; i < 2; i++ {
		n := nodes[i]
		assertEqual(t, expecteds[i].lines, n.lines)
		assertEqual(t, expecteds[i].tag, n.Data)
	}
}

func TestXPathNode_preceding(t *testing.T) {
	s := `//employee[@id=3]/preceding::*`
	nodes := selectNodes(employee_example, s)
	// Warning: The sorted result nodes is incorrect. [3, 4, ..] should be at first.
	assertEqual(t, 8, len(nodes))

	expecteds := []struct {
		lines int
		tag   string
	}{
		{8, "employee"},
		{9, "name"},
		{10, "designation"},
		{11, "email"},

		{3, "employee"},
		{4, "name"},
		{5, "designation"},
		{6, "email"},
	}

	for i := 0; i < 8; i++ {
		n := nodes[i]
		assertEqual(t, expecteds[i].lines, n.lines)
		assertEqual(t, expecteds[i].tag, n.Data)
	}
}

func TestXPathNode_preceding_sibling(t *testing.T) {
	s := `//employee[@id=3]/preceding-sibling::*`
	nodes := selectNodes(employee_example, s)
	// Warning: The sorted result nodes is incorrect. [3, 8] should be at first.
	assertEqual(t, 2, len(nodes))

	expecteds := []struct {
		lines int
		tag   string
	}{
		{8, "employee"},
		{3, "employee"},
	}

	for i := 0; i < 8; i++ {
		n := nodes[i]
		assertEqual(t, expecteds[i].lines, n.lines)
		assertEqual(t, expecteds[i].tag, n.Data)
	}

}
