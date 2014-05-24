package jigo

import (
	"fmt"
	"testing"
)

func TestStack(t *testing.T) {
	var p Pos
	var n Node
	s := newStack(p)
	s.push(n)
	if s.len() != 1 {
		t.Errorf("Expected s.len() to be 1, got %d\n", s.len())
	}
	s.pop()
	if s.len() != 0 {
		t.Errorf("Expected s.len() to be 0, got %d\n", s.len())
	}
	n = s.pop()
	if n != nil {
		t.Errorf("Expected n to be nil, but was %v\n", n)
	}
	s.push(newList(p))
	s.push(newLookup(p, "foo"))
	s.push(newText(p, "Hello!"))

	if s.len() != 3 {
		t.Errorf("Expected s.len() to be 3, got %d\n", s.len())
	}
	n = s.pop()
	if n.Type() != NodeText {
		t.Errorf("Expected n.Type to be NodeText, got %s\n", n.Type())
	}
	if n.String() != "Hello!" {
		t.Errorf("Expected n.String() to be \"Hello!\", got %s\n", n.String())
	}

	n = s.pop()

	if n.Type() != NodeLookup {
		t.Errorf("Expected n.Type() to be NodeLookup, got %s\n", n.Type())
	}
	if n.String() != "foo" {
		t.Errorf("Expected n.String() to be \"foo\", got %s\n", n.String())
	}
	if s.len() != 1 {
		t.Errorf("Expected len of 1, got %d\n", s.len())
	}
	fmt.Println(s)
}
