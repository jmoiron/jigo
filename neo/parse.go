package jigo

import (
	"fmt"
	"runtime"
	"strings"
)

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
func (t *Tree) parseVar() Node {
	token := t.expect(tokenVariableBegin)
	n := newVar(token.pos)
	exprList := newList(token.pos)
	for {
		token := t.peekNonSpace()
		switch token.typ {
		case tokenName:
			exprList.append(t.varExpr())
			continue
		case tokenLparen:
			exprList.append(t.parenExpr())
			continue
		case tokenLbrace:
			exprList.append(t.mapExpr())
			continue
		case tokenLbracket:
			exprList.append(t.listExpr())
			continue
		case tokenGt, tokenGteq, tokenLt, tokenLteq, tokenEqEq:
			t.unexpected(token, "unexpected boolean operator in var block")
		case tokenVariableEnd:
			t.nextNonSpace()
		default:
			t.unexpected(token, "end variable")
		}
		break
	}
	n.Expr = exprList
	return n
}

func (t *Tree) varExpr() Node {
	t.next()
	return nil
}

func (t *Tree) parenExpr() Node {
	t.next()
	return nil
}

func (t *Tree) mapExpr() Node {
	t.next()
	return nil
}

func (t *Tree) listExpr() Node {
	t.next()
	return nil
}
