package jigo

type expr interface {
	Eval(ctx context) (interface{}, error)
}
