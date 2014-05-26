package jigo

import (
	"fmt"
	"testing"
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

	fmt.Println("WHEE")
	tester.Test(
		`{{ 1 + 2 * 3 + 4}}`,
		parseTest{nodeTypes: []NodeType{NodeVar}},
	)

	tester.Test(
		`{{ {"hello": "world"}[choice] }}`,
		parseTest{nodeTypes: []NodeType{}},
	)

}
