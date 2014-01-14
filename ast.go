package jigo

import (
	"bytes"
	"fmt"
)

type node interface {
	Render(context) (string, error)
}

// default render buffer for the root node set to 128K
const defaultBufferSize = 1024 * 128

// root AST node
type listNode struct {
	nodes []node
}

// a textNode which corresponds to text outside the template lang
type textNode struct {
	tok *token
}

type conditionalNode struct {
	guard expr
	body  node
}

type stringNode struct {
	tok   *token
	value string
}

type lookupNode struct {
	tok  *token
	name string
}

func (l listNode) Render(ctx context) (string, error) {
	b := make([]byte, 0, defaultBufferSize)
	buf := bytes.NewBuffer(b)
	for _, node := range l.nodes {
		s, err := node.Render(ctx)
		if err != nil {
			return "", err
		}
		buf.WriteString(s)
	}
	return buf.String(), nil
}

func (t textNode) Render(ctx context) (string, error) {
	return string(t.tok.value), nil
}

type exprNode struct {
	toks  []*token
	exprs []expr
}

func (s stringNode) Render(ctx context) (string, error) {
	return s.value, nil
}

func (s stringNode) Eval(ctx context) (interface{}, error) {
	return s.value, nil
}

func (l lookupNode) Render(ctx context) (string, error) {
	val, err := ctx.lookup(l.name)
	// FIXME: add a loud errors mode which doesn't eat lookup errors
	if err != nil {
		return "", nil
	}
	return fmt.Sprint(val), nil
}

func (s lookupNode) Eval(ctx context) (interface{}, error) {
	return ctx.lookup(s.name)
}
