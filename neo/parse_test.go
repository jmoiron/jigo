package jigo

import (
	"fmt"
	"testing"
)

type parseTest struct {
	isError bool
	output  string
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

	fmt.Printf("Tree: %#v\n", tree.Root)
	fmt.Printf("Tree: %s\n", tree.Root)
}

func TestParser(t *testing.T) {
	tester := parsetest{t}
	fmt.Println(tester)

	tester.Test(
		`Hello, {# comment #}World`,
		parseTest{isError: false},
	)
}
