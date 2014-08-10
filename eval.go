package jigo

import (
	"bytes"
	"fmt"
)

// This file contains ast evaluation.
//
// Eval was kept separate from the ast so that the ast itself was not
// burdened with runtime evaluation, and also so that the ast was free
// to also be used for other purposes such as prettifying or codegen.

type renderer struct {
	t *Template
	c contextStack
	b bytes.Buffer
}

func newRenderer(t *Template) *renderer {
	return &renderer{t: t}
}

func (r *renderer) render(c contextStack) (string, error) {
	r.c = c
	err := r.renderList(r.t.base.Root)
	return r.b.String(), err
}

func (r *renderer) renderList(n *ListNode) error {
	var err error
	for _, node := range n.Nodes {
		switch t := node.(type) {
		case *TextNode:
			r.b.Write(t.Text)
		case *VarNode:
			err = r.renderVar(t)
		default:
			return fmt.Errorf("Unknown node type %v", t.Type())
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *renderer) renderVar(n *VarNode) error {
	switch t := n.Node.(type) {
	case *LookupNode:
		return r.renderLookup(t)
	case *AddExpr:
		r.b.WriteString("?")
		return nil
	default:
		return fmt.Errorf("Unknown node type %v", t.Type())
	}
	return nil
}

func (r *renderer) renderLookup(n *LookupNode) error {
	// FIXME: strict mode where lookup failures are runtime errors?
	v, ok := r.c.lookup(n.Name)
	if ok {
		r.b.WriteString(fmt.Sprint(v.Interface()))
	}
	return nil
}
