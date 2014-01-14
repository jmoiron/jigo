package jigo

import (
	"fmt"
	"regexp"
)

// the grammar for jinja is fairly complex because it not only includes python
// literals like strings, lists, tuples, and dicts, some of which are dynamically
// typed *and* mutable, but it also has support for lots of python expressions
// like basic as well as python expressions like arithmetic exprs, list comps,
// exponentiation, unpacking, etc.
//
// jinja2 can get literals safely and for free with python's `ast.literal_eval`
// but we have no such luck, and it could fall back on 'eval' for expression
// evaluation but we'd have to implement a significant portion of python's
// builtin data and evaluation model to get this.
//
// instead, we'll start with support for literals like ints, floats, strings,
// and lists, mapping to int64, float64, string, and []interface{}, but eschew
// implementing non-boolean expressions.
//
// this makes a lot of simple templates like `{{ 1 + 2 }}` invalid, but this
// can be added later and starting out it's more important to get jinja's rich
// module, looping, and macro systems implemented first.

// TODO: does jinja support binary/octal/hex literals?
// NOTE: container literals with lookupExprs in them have to be evaluated
//       at rendering time, and may be much slower than container literals
//       which contain only constants.

// grammar:
//		number = [0-9]+ | [0-9]+ . [0-9]+
//		word = [a-zA-Z] [a-zA-Z0-9]+
//		binop = <|<=|>|>=|!=|==
//		boolbinop = or|and
// 		boolunary = not
//		filter = |
//		string = " .* "
//      list = "[" [expr, expr...] "]"
//      literal = string | number
//		atom = word | literal
//		groupexpr = ( expr )
//		paramexpr = ( expr[, expr...] )
//		funcexpr = word paramexpr
//		filterexpr = "|" word [paramexpr]
//		varexpr = variable [filterexpr]
//		boolexpr = expr [boolbinop expr...] | boolunary expr
//		expr = varexpr | boolexpr | funcexpr [binop expr...]

type context map[string]interface{}

func (c context) lookup(name string) (interface{}, error) {
	val, ok := c[name]
	if !ok {
		return nil, fmt.Errorf("Name %s not found.", name)
	}
	return val, nil
}

type parser struct {
	lexer *scanner
	name  string
}

type parseError struct {
	line    int
	message string
}

func (p parseError) Error() string {
	return fmt.Sprintf("%3d: %s", p.line, p.message)
}

type Template struct {
	root node
	name string
}

func (t Template) Render(ctx context) (string, error) {
	return t.root.Render(ctx)
}

func ParseFile(path string) (Template, error) {
	return Template{listNode{}, path}, nil
}

func Parse(name string, body []byte) (Template, error) {
	var err error
	t := Template{name: name}
	s := newScanner(name, body)
	p := parser{s, name}
	t.root, err = p.parse()
	return t, err
}

func (p parser) parse() (node, error) {
	root := listNode{}
	for tok := p.lexer.next(); tok.token != tokenEOF; tok = p.lexer.next() {
		switch tok.token {
		case tokenText:
			root.nodes = append(root.nodes, textNode{tok})
		case tokenOpenTag:
			node, err := p.parseTagNode()
			if err != nil {
				return root, err
			}
			root.nodes = append(root.nodes, node)
		case tokenOpenVar:
			node, err := p.parseVarNode()
			if err != nil {
				return root, err
			}
			root.nodes = append(root.nodes, node)
		default:
			fmt.Printf("Unidentified token type %s\n", tok.token.Name())
		}
	}
	return root, nil
}

func (p parser) parseTagNode() (node, error) {
	return listNode{}, nil
}

var nameRegex = regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9]*`)
var intRegex = regexp.MustCompile(`[0-9]+`)
var floatRegex = regexp.MustCompile(`[0-9]+\.[0-9]+`)

// var node contains an expression
func (p parser) parseVarNode() (node, error) {
	root := listNode{}
	for tok := p.lexer.next(); tok.token != tokenCloseVar; tok = p.lexer.next() {
		switch tok.token {
		case tokenName:
			switch string(tok.value) {
			case "true", "false":
				// boolexpr
			case "nil", "none":
				// nilexpr
			default:
				if nameRegex.Match(tok.value) {
					node := lookupNode{tok, string(tok.value)}
					root.nodes = append(root.nodes, node)
				} else if intRegex.Match(tok.value) {
					// intexpr
				} else if floatRegex.Match(tok.value) {
					// floatexpr
				}
			}

		case tokenString:
			node := stringNode{tok, string(tok.value)}
			root.nodes = append(root.nodes, node)
		default:
			return root, parseError{
				p.lexer.lineno,
				fmt.Sprintf("Unexpected token type %s.", tok.token.Name()),
			}
		}
	}
	return root, nil
}
