package xpath

import "testing"

func TestXPathPredicate_Positions(t *testing.T) {
	s := `/empinfo/employee[2]`
	nodes := selectNodes(employee_example, s)
	assertEqual(t, 1, len(nodes))
	assertEqual(t, 8, nodes[0].lines)

	s = `/empinfo/employee[2]/name`
	nodes = selectNodes(employee_example, s)
	assertEqual(t, 1, len(nodes))
	assertEqual(t, 9, nodes[0].lines)

	s = `//employee[position()=2]`
	nodes = selectNodes(employee_example, s)
	assertEqual(t, 1, len(nodes))
	assertEqual(t, 8, nodes[0].lines)

	s = `//employee[position()>1]`
	nodes = selectNodes(employee_example, s)
	assertEqual(t, 2, len(nodes))

	expecteds := []int{8, 13}
	for i := 0; i < 2; i++ {
		n := nodes[i]
		assertEqual(t, expecteds[i], n.lines)
	}

	s = `//employee[position()<=3]`
	nodes = selectNodes(employee_example, s)
	expecteds = []int{3, 8, 13}
	for i := 0; i < 2; i++ {
		n := nodes[i]
		assertEqual(t, expecteds[i], n.lines)
	}
}

func TestXPathPredicate_Nodes(t *testing.T) {
	s := `//employee[name]`
	nodes := selectNodes(employee_example, s)
	assertEqual(t, 3, len(nodes))
	expecteds := []int{3, 8, 13}
	for i := 0; i < 2; i++ {
		n := nodes[i]
		assertEqual(t, expecteds[i], n.lines)
	}
}

func TestXPathPredicate_Operators(t *testing.T) {
	s := `/empinfo/employee[@id]`
	nodes := selectNodes(employee_example, s)
	assertEqual(t, 3, len(nodes))

	s = `/empinfo/employee[@id = 2]`
	nodes = selectNodes(employee_example, s)
	assertEqual(t, 1, len(nodes))
	assertEqual(t, 8, nodes[0].lines)

	s = `/empinfo/employee[1][@id=1]`
	nodes = selectNodes(employee_example, s)
	assertEqual(t, 1, len(nodes))
	assertEqual(t, 3, nodes[0].lines)

	s = `/empinfo/employee[@id][2]`
	nodes = selectNodes(employee_example, s)
	assertEqual(t, 1, len(nodes))
	assertEqual(t, 8, nodes[0].lines)

	s = `//designation[@discipline and @experience]`
	nodes = selectNodes(employee_example, s)
	assertEqual(t, 2, len(nodes))
	expecteds := []int{5, 10}
	for i := 0; i < 2; i++ {
		n := nodes[i]
		assertEqual(t, expecteds[i], n.lines)
	}

	s = `//designation[@discipline or @experience]`
	nodes = selectNodes(employee_example, s)
	assertEqual(t, 3, len(nodes))
	expecteds = []int{5, 10, 15}
	for i := 0; i < 3; i++ {
		n := nodes[i]
		assertEqual(t, expecteds[i], n.lines)
	}

	s = `//designation[@discipline | @experience]`
	nodes = selectNodes(employee_example, s)
	assertEqual(t, 3, len(nodes))
	expecteds = []int{5, 10, 15}
	for i := 0; i < 3; i++ {
		n := nodes[i]
		assertEqual(t, expecteds[i], n.lines)
	}

	s = `/empinfo/employee[@id != "2" ]`
	nodes = selectNodes(employee_example, s)
	assertEqual(t, 2, len(nodes))
	expecteds = []int{3, 13}
	for i := 0; i < 2; i++ {
		n := nodes[i]
		assertEqual(t, expecteds[i], n.lines)
	}

	s = `/empinfo/employee[@id > 1 ]`
	nodes = selectNodes(employee_example, s)
	assertEqual(t, 2, len(nodes))
	expecteds = []int{8, 13}
	for i := 0; i < 2; i++ {
		n := nodes[i]
		assertEqual(t, expecteds[i], n.lines)
	}

	s = `/empinfo/employee[@id and @id = "2"]`
	nodes = selectNodes(employee_example, s)
	assertEqual(t, 1, len(nodes))
	assertEqual(t, 8, nodes[0].lines)

	s = `/empinfo/employee[@id = "1" or @id = "2"]`
	nodes = selectNodes(employee_example, s)
	expecteds = []int{3, 8}
	for i := 0; i < 2; i++ {
		n := nodes[i]
		assertEqual(t, expecteds[i], n.lines)
	}
}

func TestXPathPredicate_NestedPredicate(t *testing.T) {
	s := `//employee[./name[@from]]`
	nodes := selectNodes(employee_example, s)
	assertEqual(t, 1, len(nodes))
	assertEqual(t, 8, nodes[0].lines)

	s = `//employee[.//name[@from = "CA"]]`
	nodes = selectNodes(employee_example, s)
	assertEqual(t, 1, len(nodes))
	assertEqual(t, 8, nodes[0].lines)
}
