# jigo

Jigo is a Jinja2-like template language for Golang.

## Goals

The goal was to provide a templating language for Go that would suit both
backend developers and frontend designers.  As the Django template system
has shown itself to be capable at this, and Jinja2 is refinement on that
system, it was more or less copied to create Jigo.

Django's template system is nearly language agnostic, but Jinja2's allows
for some limited programming within the templates, including support for
sophisticated python expressions.  It can do this easily as Python is
interpreted and has compilation tools (not to mention [eval]() and
[literal_eval]()), but these expressions are not consistent with what is
possible or expedient in Go.

[Twig](), a Jinja2-like template system implemented in PHP, allows for a
limited expression syntax which more closely mimics Python's than PHP's.
Jigo allows for a similarly Python-inspired expression syntax, which will
have literals which map cleanly to Go types and are strongly type-checked
at runtime.

## Differences

There are many differences.  Some infrequently used features of Jinja2 have
been dropped, and the expression syntax is far less sophisticated.

### Features

* Line statements and line comments have been dropped

**TODO** More info here as I get through more of the implementation.

## Expressions

Expressions are coerced to strings at render time with `fmt.Sprint`, but
expression syntax is strongly typed.

### Literals

* All integer numeric literals map to `int64` (beware overflow)
* All numeric literals with a "." in it become `float64`
* Strings are standard " delimited, with \\ escapes.  No multi-line or \`\` 
  string syntax support.
* Lists are defined as `'[' expr [, expr]... ']'`, and map to `[]interface{}`
* Hashes are defined as `'{' stringExpr ':' expr [, stringExpr ':' expr]... '}'`,
  and map to `map[string]interface{}`.  stringExpr is an expression that is
  coerced to a string immediately.  This means that the "1" and 1 represent the
  same key.
* No advanced python literals (comprehensions, sets, generators, etc).

### Arithmetic

* The basic arithmetic operators `+,-,/,*` work as expected on numerics of the
  same type.  `%` is only defined on integers. The extended operators `**,//`
  do power and floor-div.
* If floats and ints are mixed, all values are coerced to float64 for the
  computation.
* String concatenation is allowed via the `~` operator, which coerces all
  surrounding arguments to strings.  `+` will work if all types are strictly
  `string`, but will fail otherwise.

## Other Operators

* `is` performs a test, similar to Jinja2.
* `in` is also implemented, but only for slice and map types.
* There is no `[]`/`.` duality:
  * `[]`, the selection operator, is only valid on slice and map types
  * `.`, attribute selection operator, is only valid on struct types
* Function calling semantics allow for keyword arguments if the
  function's final argument is of type `jigo.Kwargs`.  Varargs are allowed
  if a funciton is variadic or the final argument is of type `jigo.Varargs`.

