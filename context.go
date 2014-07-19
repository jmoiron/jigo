package jigo

import (
	"fmt"
	"reflect"
)

// A context represents an environment passed in by a user to a template.  Certain
// tags can create temporary contexts (for, macro, etc), which get created at eval
// time.
type Context struct {
	ctx   interface{}
	kind  reflect.Kind
	value reflect.Value
}

// Contexts can be structs or maps, or pointers to these types, but no other type.
func NewContext(i interface{}) (*Context, error) {
	// save the original value, though we likely won't use it
	var v reflect.Value
	c := &Context{ctx: i}
	// indirect v
	for v = reflect.ValueOf(c); v.Kind() == reflect.Ptr; v = reflect.ValueOf(c) {
		v = reflect.Indirect(v)
	}
	c.kind = v.Kind()
	c.value = v
	if c.kind != reflect.Map && c.kind != reflect.Struct {
		return c, fmt.Errorf("Context must be a struct or map, not %s")
	}
	return c, nil
}

type contextStack []Context

// lookup finds a name in the context stack.  If no name is found, then an undefined
// sentinel is returned.
func (c contextStack) lookup(name string) (v reflect.Value, ok bool) {
	// TODO: go through stack
	return c[0].lookup(name)
}

// lookup finds a single name in a single context.  If no name is found, then
// an empty Value is returned and ok is False.
func (c Context) lookup(name string) (v reflect.Value, ok bool) {
	switch c.kind {
	case reflect.Map:
		v := v.MapIndex(reflect.ValueOf(name))
		return v, v.IsValid()
	case reflect.Struct:
	}
	return v, false
}
