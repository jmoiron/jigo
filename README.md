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
interpreted and has compilation tools (not to mention [eval](https://docs.python.org/2/library/functions.html#eval) and
[literal_eval](https://docs.python.org/2/library/ast.html#ast.literal_eval)), but these expressions are not consistent with what is
possible or expedient in Go.

[Twig](http://twig.sensiolabs.org/), a Jinja2-like template system implemented in PHP, allows for a
limited expression syntax which more closely mimics Python's than PHP's.
Jigo allows for a similarly Python-inspired expression syntax, which will
have a clean, repeatable two way mapping to Go types and are strongly type-checked
at runtime.

There are no plans and currently no support for context-aware escaping.
Because of its inclusion in html/template, this is a big deal for many in the Go
community, and while I'm not against adding it to Jigo, I am more interested in 
getting the language in a usable state *first*.

* If you want logic-less templates, try [moustache](https://github.com/hoisie/mustache)
* If you want execution safety, try [liquid](https://github.com/hoisie/mustache) or [mandira](http://jmoiron.github.io/mandira/)
* If you want contextually aware escaping, try [html/template](http://golang.org/pkg/html/template/)
* If you want something aiming to be compatible with django templates, try [pongo2](https://github.com/flosch/pongo2)

## Differences

There are many differences.  Some infrequently used features of Jinja2 have
been dropped, and the expression syntax is far less sophisticated.

### Features

* Line statements and line comments have been dropped

**TODO** More info here as I get through more of the implementation.

## Type Safety

Template lookups are obviously not inspectable for types at compile time, but literals
are, and although there are no type declarations, all types are inferred according to
rules defined below.  This means that the following code raises a type error at compile
time:

```jinja
{{ 1 + "foo" }}
```

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
  and map to `map[string]interface{}`.  A stringExpr is an expression that is
  coerced to a string automatically.  This means that the "1" and 1 represent the
  same key.
* No advanced python expressions/literals (comprehensions, sets, generators, etc).

### Arithmetic

* The basic arithmetic operators `+,-,/,*` work as expected on numerics of the
  same type.  `%` is only defined on integers, as in Go, and has the same
  behavior as Go's `%` operator, which *differs* from Python's. The extended operators 
  `**,//` do power and floor-div, repsepctfully.
* If floats and ints are mixed, all values are coerced to float64 for the
  computation.
* String concatenation is allowed via the `~` operator, which coerces all
  surrounding arguments to strings.  `+` will work if all types are strictly
  `string`, but will fail a type check otherwise.

## Other Operators

* `is` performs a test, similar to Jinja2.
* `in` is also implemented, but only for slice and map types.
* There is no `[]`/`.` duality:
  * `[]`, the selection operator, is only valid on slice and map types
  * `.`, attribute selection operator, is only valid on struct types
* Function calling semantics allow for keyword arguments if the
  function's final argument is of type `jigo.Kwargs`.  Varargs are allowed
  if a funciton is variadic or the final argument is of type `jigo.Varargs`.

