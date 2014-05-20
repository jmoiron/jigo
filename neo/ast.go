package jigo

import (
	"bytes"
	"fmt"
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

func (l *ListNode) append(n Node) {
	l.Nodes = append(l.Nodes, n)
}

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

func (l *ListNode) Copy() Node {
	return l.CopyList()
}

// TextNode holds plain text.
type TextNode struct {
	NodeType
	Pos
	Text []byte // The text; may span newlines.
}

func newText(pos Pos, text string) *TextNode {
	return &TextNode{NodeType: NodeText, Pos: pos, Text: []byte(text)}
}

func (t *TextNode) String() string {
	return fmt.Sprintf(textFormat, t.Text)
}

func (t *TextNode) Copy() Node {
	return &TextNode{NodeType: NodeText, Text: append([]byte{}, t.Text...)}
}

// VarNode represents a var print expr, ie {{ ... }}.
type VarNode struct {
	NodeType
	Pos
	Expr Node // The expression node within this one
}

func newVar(pos Pos) *VarNode {
	return &VarNode{NodeType: NodeVar, Pos: pos}
}

func (v *VarNode) String() string {
	return fmt.Sprintf("{{ %s }}", v.Expr)
}

func (v *VarNode) Copy() Node {
	return &VarNode{NodeType: NodeVar, Expr: v.Expr.Copy()}
}
