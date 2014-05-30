package jigo

import (
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func (n NodeType) String() string {
	switch n {
	case NodeList:
		return "NodeList"
	case NodeText:
		return "NodeText"
	case NodeVar:
		return "NodeVar"
	default:
		return "Unknown Type"
	}
}

// Recurisvely spew a whole tree of nodes.
func spewTree(n Node, indent string) {
	switch n.(type) {
	case *ListNode:
		n := n.(*ListNode)
		fmt.Printf("%s(ListNode) {\n", indent)
		for _, n := range n.Nodes {
			spewTree(n, indent+"  ")
		}
		fmt.Printf("%s}\n", indent)
	case *VarNode:
		n := n.(*VarNode)
		fmt.Printf("%s(VarNode) {\n", indent)
		spewTree(n.Node, indent+"  ")
		fmt.Printf("%s}\n", indent)
	case *AddExpr:
		n := n.(*AddExpr)
		fmt.Printf("%s(AddExpr) {\n", indent)
		spewTree(n.lhs, indent+"  ")
		fmt.Printf("%s    +\n", indent)
		spewTree(n.rhs, indent+"  ")
		fmt.Printf("%s}\n", indent)
	case *MulExpr:
		n := n.(*MulExpr)
		fmt.Printf("%s(MulExpr) {\n", indent)
		spewTree(n.lhs, indent+"  ")
		fmt.Printf("%s    *\n", indent)
		spewTree(n.rhs, indent+"  ")
		fmt.Printf("%s}\n", indent)
	default:
		fmt.Printf(indent)
		spew.Dump(n)
	}
}

type parseTest struct {
	isError   bool
	nodeTypes []NodeType
}

type parsetest struct {
	*testing.T
}

func (p *parsetest) Test(input string, test parseTest) {
	t := p.T
	e := NewEnvironment()
	t.Log(input)
	tree, err := e.parse(input, "test", "test.jigo")

	/*
		if len(nodes) != len(tests) {
			t.Errorf("Expected %d nodes, got %d\n", len(tests), len(nodes))
		}
		for i, tok := range nodes {
			if i >= len(tests) {
				return
			}
			test := tests[i]
			if test.typ != tok.typ {
				fmt.Printf("nodes: %v\ntests:  %v\n", nodes, tests)
				t.Errorf("Expecting %d token type to be %d, got %d\n", i, test.typ, tok.typ)
			}
			if tok.typ != tokenWhitespace && test.val != tok.val {
				fmt.Printf("nodes: %v\ntests:  %v\n", nodes, tests)
				t.Errorf("Expecting %d token val to be `%s`, got `%s`\n", i, test.val, tok.val)
			}
		}
	*/
	if err != nil && !test.isError {
		t.Errorf("Unexpected error: %s\n", err)
	}

	if len(test.nodeTypes) != len(tree.Root.Nodes) {
		t.Errorf("Wrong number of nodes in %s\n", tree.Root)
		t.Errorf("Was expecting %d top level nodes, found %d", len(tree.Root.Nodes), len(test.nodeTypes))
		return
	}

	for i, nt := range test.nodeTypes {
		rnt := tree.Root.Nodes[i].Type()
		if nt != rnt {
			t.Errorf("Type mismatch: expecting %dth to be %s, but was %s", i, nt, rnt)
		}
	}
	spewTree(tree.Root, "")
	t.Log(tree.Root)
}

func TestParser(t *testing.T) {
	tester := parsetest{t}

	tester.Test(
		`Hello, World`,
		parseTest{nodeTypes: []NodeType{NodeText}},
	)

	tester.Test(
		`Hello, {# comment #}World`,
		parseTest{nodeTypes: []NodeType{NodeText, NodeText}},
	)

	tester.Test(
		`Hello {{ name }}`,
		parseTest{nodeTypes: []NodeType{NodeText, NodeVar}},
	)

	tester.Test(
		`{{ 1 + 2 }}`,
		parseTest{nodeTypes: []NodeType{NodeVar}},
	)

	tester.Test(
		`{{ "foo" + "bar" }}`,
		parseTest{nodeTypes: []NodeType{NodeVar}},
	)

	tester.Test(
		`{{ 1 + 2 * 3 + 4}}`,
		parseTest{nodeTypes: []NodeType{NodeVar}},
	)

	tester.Test(
		`{{ {"hello": "world", 1: "one"} }}`,
		parseTest{nodeTypes: []NodeType{NodeVar}},
	)

}
