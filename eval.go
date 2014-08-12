package jigo

import (
	"bytes"
	"errors"
	"fmt"
	"math"
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

func (r *renderer) renderNode(n Node) error {
	switch t := n.(type) {
	case *TextNode:
		_, err := r.b.Write(t.Text)
		return err
	case *VarNode:
		return r.renderVar(t)
	case *IfBlockNode:
		return r.renderCond(t)
	case *ListNode:
		return r.renderList(t)
	default:
		return fmt.Errorf("Unknown node type %v", t.Type())
	}

}

func (r *renderer) renderList(n *ListNode) error {
	for _, node := range n.Nodes {
		err := r.renderNode(node)
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
		i, err := eval(t, r.c)
		if err != nil {
			return err
		}
		// evaluated expressions are coerced to string with Sprint before rendering
		r.b.WriteString(fmt.Sprint(i))
		return nil
	default:
		return fmt.Errorf("Unknown node type %v", t.Type())
	}
	return nil
}

// renderCond renders evaluates and renders conditional block tags
func (r *renderer) renderCond(n *IfBlockNode) error {
	for _, cond := range n.Conditionals {
		c := cond.(*ConditionalNode)
		g, err := eval(c.Guard, r.c)
		if err != nil {
			return err
		}
		val, err := asBool(g)
		if err != nil {
			return fmt.Errorf(`Non-boolean "%s" used in boolean context.`, g)
		}
		if val {
			return r.renderNode(c.Body)
		}
	}
	// if there's an else, render it
	if n.Else != nil {
		return r.renderNode(n.Else)
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

// main ltr eval
func eval(n Node, c contextStack) (interface{}, error) {
	switch t := n.(type) {
	case *LookupNode:
		// we ignore lookup errors here and return nil
		val, ok := c.lookup(t.Name)
		if !ok {
			return nil, nil
		}
		return val.Interface(), nil
	case *FloatNode:
		return t.Value, nil
	case *IntegerNode:
		return t.Value, nil
	case *StringNode:
		return t.Value, nil
	case *BoolNode:
		return t.Value, nil
	case *AddExpr:
		lhs, err := eval(t.lhs, c)
		if err != nil {
			return nil, err
		}
		rhs, err := eval(t.rhs, c)
		if err != nil {
			return nil, err
		}
		return evalAdd(lhs, rhs, t.operator)
	}
	return nil, nil
}

// evalAdd evaluatse arithmetic expressions between an lhs and an rhs, which
// have already been evaluated themselves and turned to interface{} values.
// The type of the lhs determines the expected type on the rhs.  If the types
// are not compatible, then an error is returned.  Mixed numeric types are
// coerced to float64.
func evalAdd(lhs, rhs interface{}, oper item) (interface{}, error) {
	lt, rt := typeOf(lhs), typeOf(rhs)
	if lt != rt {
		// if both types are numeric, perform operation as float64
		if isNumericVar(lt) && isNumericVar(rt) {
			lt = floatType
		} else {
			return nil, fmt.Errorf("type error: %s and %s not compatible with %s", lt, rt, oper.val)
		}
	}

	switch lt {
	case stringType:
		return arithmeticString(asString(lhs), asString(rhs), oper)
	case intType:
		l, _ := asInteger(lhs)
		r, _ := asInteger(rhs)
		return arithmeticInt(l, r, oper)
	case floatType:
		l, _ := asFloat(lt)
		r, _ := asFloat(rt)
		return arithmeticFloat(l, r, oper)
	}
	return "?add", nil
}

func arithmeticFloat(lhs, rhs float64, oper item) (float64, error) {
	switch oper.val {
	case "+":
		return lhs + rhs, nil
	case "-":
		return lhs - rhs, nil
	case "*":
		return lhs * rhs, nil
	case "/":
		return lhs / rhs, nil
	case "//":
		return math.Floor(lhs / rhs), nil
	case "%":
		return 0.0, errors.New("% not defined on float")
	}
	return 0.0, errors.New("Unknown operator " + oper.val)
}

func arithmeticInt(lhs, rhs int64, oper item) (int64, error) {
	switch oper.val {
	case "+":
		return lhs + rhs, nil
	case "-":
		return lhs - rhs, nil
	case "*":
		return lhs * rhs, nil
	case "/":
		return lhs / rhs, nil
	case "//":
		return lhs / rhs, nil
	case "%":
		return lhs % rhs, nil
	}
	return 0.0, errors.New("Unknown operator " + oper.val)
}

func arithmeticString(lhs, rhs string, oper item) (string, error) {
	if oper.val != "+" {
		return "", errors.New(oper.val + " not defined on string")
	}
	return lhs + rhs, nil

}
