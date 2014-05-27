package jigo

import (
	"fmt"
	"runtime"
	"strings"
)

// Important to jigo, as to most languages, is the idea of an expression.
// Here is the ebnf grammar for jigo expressions, which are actually largely
// based off of the ebnf grammar for Go.  The way that these manifest
// themselves in a template is quite different, since they must be evaluated
// in potentially different contexts at runtime.

/*
NOTE: Jigo's only knowledge of custom types is whether or not the semantic
operators provide can operate on them.  These decisions obviously must be
made at runtime.  Its internally defined types are any type which for which
you can define a literal:

Type                Go Type
string              string
int                 int64
float               float64
list                []interface{}
map                 map[interface{}]interface{}

Obviously, wherever there is interface{}, you can use expressions which
evaluate to Go types.

NOTE: Operands form the basis for all combinatorial expressions.  They can
be parenthetical expressions themselves.  Lower-cased things like int_lit
and identifier are actually defined at the lexer.

Operand    = Literal | identifier | MethodExpr | "(" Expression ")" .
Literal    = BasicLit | MapLiteral | ListLiteral .
BasicLit   = int_lit | float_lit  string_lit .


MapLiteral     = "{" [ MapElementList [ "," ] ] "}"
ListLiteral    = "[" [ ElementList [ "," ] ] "]"
MapElementList = MapElement { "," MapElement }
MapElement     = Key ":" Element
ElementList    = Element { "," Element }
Element        = Expression | Operand
Key            = Literal | Operand

PrimaryExpr =
    Operand |
    PrimaryExpr Selector |
    PrimaryExpr Index |
    PrimaryExpr Slice |
    PrimaryExpr TypeAssertion |
    PrimaryExpr Call .


// NOTE: Variadic functions, type assertions, and 3-part slices are removed from the
// Go grammar to reduce initial complexity.  We are not quite implementing a statically
// typed templating language, just one that has "sane" literals and sensible evaluation
// of expected programming language expressions.

Selector       = "." identifier .
Index          = "[" Expression "]" .
Slice          = "[" ( [ Expression ] ":" [ Expression ] ) "]" .
Call           = "(" [ ArgumentList [ "," ] ] ")" .
ArgumentList   = ExpressionList
ExpressionList = Expression { "," Expression }


Expression = UnaryExpr | Expression binary_op UnaryExpr .
UnaryExpr  = PrimaryExpr | unary_op UnaryExpr .

binary_op  = "||" | "&&" | rel_op | add_op | mul_op .
rel_op     = "==" | "!=" | "<" | "<=" | ">" | ">=" .
add_op     = "+" | "-" | "^" .
mul_op     = "*" | "/" | "%" | "//" .
unary_op   = "+" | "-" | "!" | "*" .

NOTE: Because bitwise or operator | is required as the filter operator,
the other bitwise &, &^, ^, << and >> and unary ^ operators have been
removed as their usefulness in templates seems dubious.

In addition, divmod/floordiv (//) has been added, and the channel op `<-`
has been removed.

Precedence    Operator
    5             *  /  //  %
    4             +  -
    3             ==  !=  <  <=  >  >=
    2             &&
    1             ||

+    sum                    integers, floats, strings
-    difference             integers, floats
*    product                integers, floats
/    quotient               integers, floats
//   divmod                 integers, floats
%    remainder              integers

*/

// Tree is the representation of a single parsed template.
type Tree struct {
	Name      string    // name of the template represented by the tree.
	ParseName string    // name of the top-level template during parsing, for error messages.
	Root      *ListNode // top-level root of the tree.
	text      string    // text parsed to create the template (or its parent)
	// Parsing only; cleared after parse.
	// funcs []map[string]interface{}
	lex *lexer
	// FIXME: the peek max is based on the way that the grammar works,
	// but I don't know enough about the expression grammar i've loosely
	// described to know if 3 is sufficient.
	token     [3]item // three-token lookahead for parser.
	peekCount int
	stack     nodeStack
	// vars      []string // variables defined at the moment.
}

// Copy returns a copy of the Tree. Any parsing state is discarded.
func (t *Tree) Copy() *Tree {
	if t == nil {
		return nil
	}
	return &Tree{
		Name:      t.Name,
		ParseName: t.ParseName,
		Root:      t.Root.CopyList(),
		text:      t.text,
	}
}

// next returns the next token.
func (t *Tree) next() item {
	if t.peekCount > 0 {
		t.peekCount--
	} else {
		t.token[0] = t.lex.nextItem()
	}
	return t.token[t.peekCount]
}

// backup backs the input stream up one token.
func (t *Tree) backup() {
	t.peekCount++
}

// backup2 backs the input stream up two tokens.
// The zeroth token is already there.
func (t *Tree) backup2(t1 item) {
	t.token[1] = t1
	t.peekCount = 2
}

// backup3 backs the input stream up three tokens
// The zeroth token is already there.
func (t *Tree) backup3(t2, t1 item) { // Reverse order: we're pushing back.
	t.token[1] = t1
	t.token[2] = t2
	t.peekCount = 3
}

// peek returns but does not consume the next token.
func (t *Tree) peek() item {
	if t.peekCount > 0 {
		return t.token[t.peekCount-1]
	}
	t.peekCount = 1
	t.token[0] = t.lex.nextItem()
	return t.token[0]
}

// nextNonSpace returns the next non-space token.
func (t *Tree) nextNonSpace() (token item) {
	for {
		token = t.next()
		if token.typ != tokenWhitespace {
			break
		}
	}
	return token
}

// peekNonSpace returns but does not consume the next non-space token.
func (t *Tree) peekNonSpace() (token item) {
	for {
		token = t.next()
		if token.typ != tokenWhitespace {
			break
		}
	}
	t.backup()
	return token
}

// expect peeks at the next non-space token, and if it is not itemType
// fails with an error.  If it is, that item is returned and consumed.
func (t *Tree) expect(i itemType) (token item) {
	token = t.peekNonSpace()
	if token.typ != i {
		t.unexpected(token, fmt.Sprint(i))
	}
	return t.nextNonSpace()
}

// Parsing.

// New allocates a new parse tree with the given name.
func newTree(name string) *Tree {
	return &Tree{
		Name: name,
	}
}

// ErrorContext returns a textual representation of the location of the node in the input text.
func (t *Tree) ErrorContext(n Node) (location, context string) {
	pos := int(n.Position())
	text := t.text[:pos]
	byteNum := strings.LastIndex(text, "\n")
	if byteNum == -1 {
		byteNum = pos // On first line.
	} else {
		byteNum++ // After the newline.
		byteNum = pos - byteNum
	}
	lineNum := 1 + strings.Count(text, "\n")
	context = n.String()
	if len(context) > 20 {
		context = fmt.Sprintf("%.20s...", context)
	}
	return fmt.Sprintf("%s:%d:%d", t.ParseName, lineNum, byteNum), context
}

// errorf formats the error and terminates processing.
func (t *Tree) errorf(format string, args ...interface{}) {
	t.Root = nil
	format = fmt.Sprintf("template: %s:%d: %s", t.ParseName, t.lex.lineNumber(), format)
	panic(fmt.Errorf(format, args...))
}

// recover is the handler that turns panics into returns from the top level of Parse.
func (t *Tree) recover(errp *error) {
	e := recover()
	if e != nil {
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
		if t != nil {
			t.stopParse()
		}
		*errp = e.(error)
	}
	return
}

// unexpected complains about the token and terminates processing.
func (t *Tree) unexpected(token item, context string) {
	t.errorf("unexpected %s in %s", token, context)
}

// startParse initializes the parser, using the lexer.
func (t *Tree) startParse(lex *lexer) {
	t.Root = nil
	t.lex = lex
}

// stopParse terminates parsing.
func (t *Tree) stopParse() {
	t.lex = nil
}

// Parse parses the template given the lexer.
func (t *Tree) Parse(lex *lexer) (tree *Tree, err error) {
	defer t.recover(&err)
	t.ParseName = t.Name
	t.startParse(lex)
	t.text = lex.input
	t.parse()
	t.stopParse()
	return t, nil
}

// -- parsing --

// Here is where the code will depart quite a bit from text/template.
// Starting at parse(), we must take care of all sorts of different
// block types and tags, and we also have no concept of embedding
// multiple named templates within one file/[]byte/string.

// parse is the top-level parser for a template, essentially the same
// as itemList except it also parses {{define}} actions.
// It runs to EOF.
func (t *Tree) parse() (next Node) {
	t.Root = newList(t.peek().pos)
	for t.peek().typ != tokenEOF {
		var n Node
		switch t.peek().typ {
		case tokenBlockBegin:
			// the start of a {% .. %} tag block.
			fmt.Println("Got BlockBegin")
			continue

		case tokenVariableBegin:
			// the start of a {{ .. }} variable print block.
			n = t.parseVar()

		case tokenCommentBegin:
			t.skipComment()
			continue

		case tokenText:
			// this token is text, lets save it in a text node and continue
			n = t.parseText()
		}
		t.Root.append(n)

		/*
			delim := t.next()
			if t.nextNonSpace().typ == itemDefine {
				newT := New("definition") // name will be updated once we know it.
				newT.text = t.text
				newT.ParseName = t.ParseName
				newT.startParse(t.funcs, t.lex)
				newT.parseDefinition(treeSet)
				continue

			}
			t.backup2(delim)


			n := t.textOrAction()
			if n.Type() == nodeEnd {
				t.errorf("unexpected %s", n)
			}
			t.Root.append(n)
		*/
	}
	return nil
}

func (t *Tree) parseText() Node {
	switch token := t.next(); token.typ {
	case tokenText:
		return newText(token.pos, token.val)
	default:
		t.unexpected(token, "input")
	}
	return nil
}

// Skips over a comment;  comments are not represented in the final AST.
func (t *Tree) skipComment() {
	t.expect(tokenCommentBegin)
	for {
		token := t.nextNonSpace()
		switch token.typ {
		case tokenText:
			continue
		case tokenCommentEnd:
		default:
			t.unexpected(token, "end commend")
		}
		break
	}
}

// Parse a variable print expression, from tokenVariableBegin to tokenVariableEnd
// Contains a single expression.
func (t *Tree) parseVar() Node {
	token := t.expect(tokenVariableBegin)
	expr := newVar(token.pos)
	expr.Node = t.parseExpr(nil, tokenVariableEnd)
	t.expect(tokenVariableEnd)
	return expr
	/*
			case tokenAdd, tokenSub:
				expr.append(newAritmeticOp(token))
			case tokenMul, tokenDiv, tokenFloordiv:
				// how do we do this again ?
			case tokenGt, tokenGteq, tokenLt, tokenLteq, tokenEqEq:
				t.unexpected(token, "unexpected boolean operator in var block")
			case tokenVariableEnd:
				t.nextNonSpace()
			default:
				t.unexpected(token, "end variable")
			}
			break
		}
		return n
	*/
}

// Parses an expression until it hits a terminator.  An expression one of
// a few types of expressions, some of which can contain Expressions
// themselves.  A stack is passed with each callframe.
func (t *Tree) parseExpr(stack *nodeStack, terminator itemType) Node {
	token := t.peekNonSpace()
	if stack == nil {
		stack = newStack(token.pos)
	}
	for {
		token = t.peekNonSpace()
		switch token.typ {
		case terminator:
			if stack.len() != 1 {
				fmt.Printf("Stack: %#v\n", stack)
				t.unexpected(token, "zero length expression")
			}
			return stack.pop()
		case tokenName:
			stack.push(t.lookupExpr())
		case tokenLparen:
			t.expect(tokenLparen)
			stack.push(t.parseExpr(tokenRparen))
		case tokenLbrace:
			stack.push(t.mapExpr())
		case tokenLbracket:
			if stack.len() == 0 {
				stack.push(t.listExpr())
			} else {
				stack.push(t.indexExpr())
			}
		case tokenFloat, tokenInteger, tokenString:
			stack.push(t.literalExpr())
		case tokenAdd, tokenSub:
			t.nextNonSpace()
			if stack.len() > 0 {
				lhs := stack.pop()
				rhs := t.parseExpr(terminator)
				// TODO: we must peek to see if the next oper is a mul oper
				// in order to conserve order of operations
				stack.push(newAddExpr(lhs, rhs, token))
			} else {
				t.unexpected(token, "binary op")
			}
			// FIXME: unary + is a noop, but unary - isn't..
		case tokenMul, tokenMod, tokenDiv, tokenFloordiv:
			t.nextNonSpace()
			if stack.len() > 0 {
				lhs := stack.pop()
				rhs := t.parseExpr(terminator)
				// we know this strongly binds ltr so no need to peek
				stack.push(newMulExpr(lhs, rhs, token))
				panic("exit")
			} else {
				t.unexpected(token, "binary op")
			}
		default:
			t.unexpected(token, "expression")
		}
	}
	panic("required expr but got none")
}

// in this sense, a literal is a simple lexer-level literal
func (t *Tree) literalExpr() Node {
	token := t.nextNonSpace()
	switch token.typ {
	case tokenFloat, tokenInteger, tokenString:
		return newLiteral(token.pos, token.typ, token.val)
	default:
		t.unexpected(token, "literal")
	}
	return nil
}

func (t *Tree) lookupExpr() Node {
	name := t.nextNonSpace()
	return newLookup(name.pos, name.val)
}

func (t *Tree) parenExpr() Node {
	t.next()
	return nil
}

func (t *Tree) mapExpr() Node {
	for {
		token := t.nextNonSpace()
		switch token.typ {
		case tokenRbracket:
			return nil
		}
	}
}

func (t *Tree) listExpr() Node {
	t.next()
	return nil
}

func (t *Tree) indexExpr() Node {
	for {
		token := t.nextNonSpace()
		switch token.typ {
		case tokenRbrace:
			return nil
		}
	}
}
