package jigo

import (
	"bytes"
	"fmt"
	"strconv"
)

var textFormat = "%s" // Changed to "%q" in tests for better error messages.

type NodeType int

func (t NodeType) Type() NodeType {
	return t
}

type Node interface {
	Type() NodeType
	String() string
	// Copy does a deep copy of the Node and all its components.
	Copy() Node
	Position() Pos // byte position of start of node in full original input string
}

const (
	NodeList NodeType = iota
	NodeText
	NodeVar
	NodeLookup
	NodeFloat
	NodeInteger
	NodeString
)

// ListNode holds a sequence of nodes.
type ListNode struct {
	NodeType
	Pos
	Nodes []Node // The element nodes in lexical order.
}

func newList(pos Pos) *ListNode {
	return &ListNode{NodeType: NodeList, Pos: pos}
}

func (l *ListNode) append(n Node) { l.Nodes = append(l.Nodes, n) }

func (l *ListNode) String() string {
	b := new(bytes.Buffer)
	for _, n := range l.Nodes {
		fmt.Fprint(b, n)
	}
	return b.String()
}

func (l *ListNode) CopyList() *ListNode {
	if l == nil {
		return l
	}
	n := newList(l.Pos)
	for _, elem := range l.Nodes {
		n.append(elem.Copy())
	}
	return n
}

func (l *ListNode) Copy() Node { return l.CopyList() }

// TextNode holds plain text.
type TextNode struct {
	NodeType
	Pos
	Text []byte // The text; may span newlines.
}

func newText(pos Pos, text string) *TextNode {
	return &TextNode{NodeType: NodeText, Pos: pos, Text: []byte(text)}
}

func (t *TextNode) String() string { return fmt.Sprintf(textFormat, t.Text) }
func (t *TextNode) Copy() Node     { return &TextNode{NodeText, t.Pos, append([]byte{}, t.Text...)} }

// VarNode represents a var print expr, ie {{ ... }}.
// It is represented as a sequence of expressions.
type VarNode struct {
	NodeType
	Pos
	Node Node
}

func newVar(pos Pos) *VarNode {
	return &VarNode{NodeType: NodeVar, Pos: pos}
}

func (v *VarNode) String() string { return "{{ " + v.Node.String() + " }}" }
func (v *VarNode) Copy() Node     { return &VarNode{v.NodeType, v.Pos, v.Node} }

// A LookupNode is a variable lookup.
type LookupNode struct {
	NodeType
	Pos
	Name string
}

func newLookup(pos Pos, name string) *LookupNode {
	return &LookupNode{NodeType: NodeLookup, Pos: pos, Name: name}
}
func (l *LookupNode) String() string { return l.Name }
func (l *LookupNode) Copy() Node     { return newLookup(l.Pos, l.Name) }

type StringNode struct {
	NodeType
	Pos
	Value string
}

func (s *StringNode) Copy() Node     { return &StringNode{s.NodeType, s.Pos, s.Value} }
func (s *StringNode) String() string { return fmt.Sprintf(`"%s"`, s.Value) }

type IntegerNode struct {
	NodeType
	Pos
	Value int64
}

func (i *IntegerNode) Copy() Node     { return &IntegerNode{i.NodeType, i.Pos, i.Value} }
func (i *IntegerNode) String() string { return strconv.FormatInt(i.Value, 10) }

type FloatNode struct {
	NodeType
	Pos
	Value float64
}

func (f *FloatNode) Copy() Node     { return &FloatNode{f.NodeType, f.Pos, f.Value} }
func (f *FloatNode) String() string { return fmt.Sprint(f.Value) }

// newLiteral creates a new string, integer, or float node depending on itemType
func newLiteral(pos Pos, typ itemType, val string) Node {
	switch typ {
	case tokenFloat:
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			panic(err)
		}
		return &FloatNode{NodeFloat, pos, v}
	case tokenInteger:
		// FIXME: complex integer types?  hex, octal, etc?
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			panic(err)
		}
		return &IntegerNode{NodeInteger, pos, v}
	case tokenString:
		return &StringNode{NodeString, pos, val}
	}
	panic(fmt.Sprint("unexpected literal type ", typ))
}
