# jigo

[![Build Status](https://drone.io/github.com/jmoiron/jigo/status.png)](https://drone.io/github.com/jmoiron/jigo/latest) [![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/jmoiron/jigo) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/jmoiron/jigo/master/LICENSE) ![version](http://img.shields.io/badge/version-pre--α-4ECDC4.svg?style=flat)


Jigo is a Jinja2-like template language for Go.  It is in *pre-alpha* stages and **not
yet ready for use much less production**.  Any documentation you see here is subject
to change.  Jigo's name is also subject to change.

## Goals

The goal of jigo is to create a template system which is powerful, flexible, and
familiar.  The [Django template][django] syntax has inspired [Jinja2][jinja2], 
[Twig][twig] in php, [Liquid][liquid] in Ruby, [Jinja-JS][jinja-js] in JavaScript,
and indeed [Pongo2][pongo2] in Go.  

Although jigo is an outright attempt to implement a very functional subset of
Jinja2's semantics, the fact that it is written in Go means that much of Jinja2's
support for rich Python expressions and semantics are dropped.

[Twig][twig], a Jinja2-like template system implemented in PHP, allows for a limited
expression syntax which more closely mimics Python's than PHP's. Jigo allows for a
similarly Python-inspired expression syntax, with a clean, explicit two way mapping
to Go types and stronger type matching requirements.

Unlike Go's [html/template][htmltemplate], There are no plans and currently no support for 
context-aware escaping. Because of `html/template`, this is a big deal for many 
in the Go community, and while I'm not against adding it to Jigo, I am more interested
in  getting the language in a usable state *first*.

* If you want logic-less templates, try [moustache][moustache-go] (*note* buggy and unmaintained)
* If you want execution safety, try [liquid][liquid-go] or [mandira][mandira]
* If you want contextually aware escaping, try [html/template][htmltemplate]
* If you want something aiming to be compatible with django templates, try [pongo2][pongo2]

[django]: https://docs.djangoproject.com/en/dev/topics/templates/ "Django Templates"
[jinja2]: http://jinja.pocoo.org/docs/ "Jinja2 Templates"
[twig]: http://twig.sensiolabs.org/ "Twig"
[liquid]: http://liquidmarkup.org/ "Liquid Markup"
[jinja-js]: https://github.com/sstur/jinja-js "Jinja-js"
[pongo2]: https://github.com/flosch/pongo2 "Pongo2 Templates"
[moustache-go]: https://github.com/hoisie/mustache "Mustache"
[liquid-go]: https://github.com/karlseguin/liquid "Liquid"
[mandira]: https://jmoiron.github.io/mandira/ "Mandira"
[htmltemplate]: https://golang.org/pkg/html/template "html/template"


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
expression syntax is strongly typed.  Unlike Go, integer and floating point
arithmetic can be mixed, but any introduction of floats coerces all values
in the expression to float.  

Function calling semantics allow for keyword arguments if the function's final
argument is of type `jigo.Kwargs`.  Varargs are allowed if a funciton is variadic
or the final argument is of type `jigo.Args`.  For a function to be both
variadic *and* keyword, it must accept (`jigo.Args`, `jigo.Kwargs`) in that order.

Jigo follows [Go's operator precedence](http://golang.org/ref/spec#Operator_precedence)
*and* Go's definition of `%`, which is *remainder*, like C, and unlike
Python.  `%` is only defined in integers, and will be a compile time error on
float literals and a runtime error on float variables.

### Other Operators:

* `**` is power, eg `2**4 = 16`
* `//` is floor-div, eg. `14//3 = 4`
* `~` is a string concatenation object, which explicitly coerces both sides to
  the string type via `fmt.Sprint`
* `is` will perform [tests]() similar to Jinja2.
* `in` is only valid for array, slice and map types.  It is linear on arrays and slices.
* `[]` is the selection operator, only valid on array, slice, and map types.
* `.` is the attribute operator, only valid on struct types.

### Literals

* All integer numeric literals map to `int64` (beware overflow)
* All numeric literals with a "." in it become `float64`
* Strings are standard " delimited, with \\ escapes.  No multi-line or \`\` 
  string syntax support.
* Lists are defined as `'[' expr [, expr]... ']'`, and map to the Go type `[]interface{}`
* Hashes are defined as `'{' stringExpr ':' expr [, stringExpr ':' expr]... '}'`,
  and map to `map[string]interface{}`.  A `stringExpr` is an expression that is
  coerced to a string automatically.  This means that the "1" and 1 represent the
  same key.
* No advanced python expressions/literals (comprehensions, sets, generators, etc).

