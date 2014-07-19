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

type Template struct {
	Name string
	base *Tree
	env  *Environment
}

func (t *Template) Eval(context interface{}) {

}

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

// Collapse N nodes into one
func (t *Tree) collapse(stack *nodeStack) Node {
	if stack.len() < 2 {
		return stack.pop()
	}
	rhs := stack.pop()
	lhs := stack.pop()
	switch lhs.(type) {
	case *AddExpr:
		lhs := lhs.(*AddExpr)
		lhs.rhs = rhs
		return lhs
	case *MulExpr:
		lhs := lhs.(*MulExpr)
		lhs.rhs = rhs
		return lhs
	default:
		stack.push(lhs)
		stack.push(rhs)
		return nil
	}
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
func (t *Tree) parse() {
	t.Root = newList(t.peek().pos)
	for n := t.parseNextNode(); n != nil; n = t.parseNextNode() {
		t.Root.append(n)
	}
}

// parseNextNode parses the next outer node and returns it.  If EOF is encountered,
// parseNextNode returns nil.  Comments are discarded.
func (t *Tree) parseNextNode() Node {
	for t.peek().typ != tokenEOF {
		switch t.peek().typ {
		case tokenCommentBegin:
			t.skipComment()
			continue
		case tokenBlockBegin:
			return t.parseBlock()
		case tokenVariableBegin:
			return t.parseVar()
		case tokenText:
			return t.parseText()
		}
	}
	return nil
}

func (t *Tree) nextBlockName() string {
	if t.peekNonSpace().typ != tokenBlockBegin {
		return ""
	}
	eat := t.nextNonSpace()
	name := t.peekNonSpace()
	t.backup2(eat)
	return name.val
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
			t.unexpected(token, "end comment")
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
}

func (t *Tree) parseBlock() Node {
	start := t.expect(tokenBlockBegin)
	blockType := t.peekNonSpace()
	switch blockType.val {
	case "for":
	case "if":
		t.backup2(start)
		return t.parseIf()
	case "block":
	case "extends":
	case "print":
	case "macro":
	case "include":
	case "from":
	case "import":
	case "call":
	case "set":
		t.backup2(start)
		return t.parseSet()
	default:
		t.unexpected(blockType, "invalid block type")
	}
	return nil
}

func (t *Tree) parseSet() Node {
	start := t.expect(tokenBlockBegin)
	set := t.nextNonSpace()
	if set.val != "set" {
		t.unexpected(set, "set")
	}
	name := t.lookupExpr()
	t.expect(tokenEq)
	val := t.parseSingleExpr(nil, tokenBlockEnd)
	t.expect(tokenBlockEnd)
	return newSet(start.pos, name, val)
}

func (t *Tree) parseIf() Node {
	begin := t.expect(tokenBlockBegin)
	iftok := t.nextNonSpace()
	if iftok.val != "if" {
		t.unexpected(iftok, "if")
	}
	node := newIf(begin.pos)

	cond := newIfCond(begin.pos)
	cond.Guard = t.parseSingleExpr(nil, tokenBlockEnd)
	t.expect(tokenBlockEnd)
	body := newList(t.peek().pos)
	// we need some kind of parseBody here

	inElse := false
	for {
		block := t.nextBlockName()
		switch block {
		case "elif":
			if inElse {
				panic("Elif encountered after previous else")
			}
			// set the body for the previous conditional and append it
			cond.Body = body
			node.Conditionals = append(node.Conditionals, cond)
			// create a new elif conditional
			cond := newElifCond(t.next().pos)
			t.nextNonSpace()
			cond.Guard = t.parseSingleExpr(nil, tokenBlockEnd)
			t.expect(tokenBlockEnd)
			body = newList(t.peek().pos)
		case "else":
			if inElse {
				panic("Else encountered after previous else")
			}
			cond.Body = body
			node.Conditionals = append(node.Conditionals, cond)
			t.expect(tokenBlockBegin)
			t.nextNonSpace()
			t.expect(tokenBlockEnd)
			body = newList(t.peek().pos)
			inElse = true
		case "endif":
			// eat the endif and return successfully
			t.expect(tokenBlockBegin)
			t.nextNonSpace()
			t.expect(tokenBlockEnd)
			if inElse {
				node.Else = body
			} else {
				node.Conditionals = append(node.Conditionals, cond)
			}
			return node
		default:
			n := t.parseNextNode()
			if n == nil {
				panic("EOF inside an If")
				// EOF inside an If
				return nil
			}
			body.append(n)
		}
	}
	fmt.Println(t.peekNonSpace())
	return nil
}

// parse a single expression simple expression.  This is a lookup, literal, or
// index expression.
func (t *Tree) parseSingleExpr(stack *nodeStack, terminator itemType) Node {
	token := t.peekNonSpace()
	switch token.typ {
	case terminator:
		t.unexpected(token, "expected expression")
	case tokenName:
		return t.lookupExpr()
	case tokenLparen:
		t.expect(tokenLparen)
		return t.parseExpr(nil, tokenRparen)
	case tokenLbrace:
		return t.mapExpr()
	case tokenLbracket:
		return t.listExpr()
	case tokenFloat, tokenInteger, tokenString, tokenBool:
		return t.literalExpr()
	case tokenAdd, tokenSub:
		unary := t.nextNonSpace()
		value := t.parseSingleExpr(nil, terminator)
		switch value.Type() {
		case NodeUnary:
			t.unexpected(unary, "expression")
		case NodeFloat:
			// FIXME: apply unary oper to value
			return value
		case NodeInteger:
			// FIXME: apply unary oper to value
			return value
		default:
			return newUnaryNode(value, unary)
		}
	default:
		t.unexpected(token, "expression")
	}
	panic("unexpected")
}

// Parses an expression until it hits a terminator.  An expression one of
// a few types of expressions, some of which can contain Expressions
// themselves.  A stack is passed with each callframe.
func (t *Tree) parseExpr(stack *nodeStack, terminator itemType) Node {
	token := t.peekNonSpace()
	if stack == nil {
		stack = newStack(token.pos)
	}
	stack.push(t.parseSingleExpr(stack, terminator))
	for {
		token = t.peekNonSpace()
		switch token.typ {
		case terminator:
			if stack.len() == 0 {
				fmt.Printf("Stack: %#v\n", stack)
				t.unexpected(token, "zero length expression")
			}
			return stack.pop()
		case tokenAdd, tokenSub:
			// consume the plus token
			t.nextNonSpace()
			// if the stack isn't empty, the previous expression is the lhs for the
			// upcoming expression:
			if stack.len() > 0 {
				rhs := t.parseExpr(stack, terminator)
				// if the token after the rhs expression is an operator with greater precedence,
				// push rhs to the stack and continue going from there
				if tok := t.peekNonSpace(); tok.precedence() > token.precedence() {
					stack.push(rhs)
					return t.parseExpr(stack, terminator)
				} else {
					// otherwise, take lhs off the stack and create a new AddExpr
					lhs := stack.pop()
					stack.push(newAddExpr(lhs, rhs, token))
				}
			}
		case tokenMul, tokenMod, tokenDiv, tokenFloordiv:
			t.nextNonSpace()
			if stack.len() > 0 {
				lhs := stack.pop()
				rhs := t.parseSingleExpr(stack, terminator)
				// we know this strongly binds ltr so no need to peek
				stack.push(newMulExpr(lhs, rhs, token))
			} else {
				t.unexpected(token, "binary op")
			}
		case tokenComma:
			// if we are terminating a map, param list, or list, return the expression
			if terminator == tokenRbracket || terminator == tokenRparen || terminator == tokenRbrace {
				if stack.len() != 1 {
					t.unexpected(token, "expression")
				}
				return stack.pop()
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
	case tokenFloat, tokenInteger, tokenString, tokenBool:
		return newLiteral(token.pos, token.typ, token.val)
	default:
		t.unexpected(token, "literal")
	}
	return nil
}

func (t *Tree) lookupExpr() Node {
	name := t.nextNonSpace()
	return t.maybeIndexExpr(newLookup(name.pos, name.val))
}

// determine if there is one or more index expressions on the end
// of the expression passed in.  If there is, return a lookup expr,
// otherwise, return the original node
func (t *Tree) maybeIndexExpr(n Node) Node {
	for {
		tok := t.peekNonSpace()
		if tok.typ == tokenLbrace {
			t.nextNonSpace()
			index := t.parseExpr(nil, tokenRbrace)
			n = newIndexExpr(n, index)
		} else {
			return n
		}
	}
}

func (t *Tree) parenExpr() Node {
	t.next()
	return nil
}

func (t *Tree) mapExpr() Node {
	tok := t.expect(tokenLbrace)
	map_ := newMapExpr(tok.pos)
	for {
		token := t.peekNonSpace()
		switch token.typ {
		case tokenComma:
			if map_.len() == 0 {
				t.unexpected(token, "map expression")
			}
			t.expect(tokenComma)
		case tokenRbrace:
			t.next()
			return t.maybeIndexExpr(map_)
		default:
			elem := t.mapElem()
			map_.append(elem.(*MapElem))
		}
	}
}

// parse a single map element;  assume that the next token is not '}'
func (t *Tree) mapElem() Node {
	key := t.parseExpr(nil, tokenColon)
	colon := t.nextNonSpace()
	if colon.typ != tokenColon {
		t.unexpected(colon, "map key expr")
	}
	val := t.parseExpr(nil, tokenRbrace)
	return newMapElem(key, val)

}

func (t *Tree) listExpr() Node {
	tok := t.expect(tokenLbracket)
	list := newList(tok.pos)
	for {
		token := t.peekNonSpace()
		switch token.typ {
		case tokenComma:
			if list.len() == 0 {
				t.unexpected(token, "list expression")
			}
			t.expect(tokenComma)
		case tokenRbracket:
			t.next()
			return t.maybeIndexExpr(list)
		default:
			elem := t.parseExpr(nil, tokenRbrace)
			list.append(elem)
		}
	}
}
