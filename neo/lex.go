// Based on code from the Go authors from text/template/parse/lex.go.
// Also based on code from jinja2's lexer.py.
// Both distributed under a BSD license.

package jigo

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Pos int

func (p Pos) Position() Pos {
	return p
}

type itemType int

// item represents a token or text string returned from the scanner.
type item struct {
	typ itemType // The type of this item.
	pos Pos      // The starting position, in bytes, of this item in the input string.
	val string   // The value of this item.
}

func (i item) String() string {
	switch {
	case i.typ == tokenEOF:
		return "EOF"
	case i.typ == tokenError:
		return fmt.Sprintf("<Err: %s>", i.val)
	case i.typ == tokenName:
		return fmt.Sprintf("<%s>", i.val)
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	default:
		return fmt.Sprintf("%q", i.val)
	}
}

// Token definitions from jinja/lexer.py
const (
	tokenAdd itemType = iota
	tokenAssign
	tokenColon
	tokenComma
	tokenDiv
	tokenDot
	tokenEq
	tokenEqEq
	tokenFloordiv
	tokenGt
	tokenGteq
	tokenLbrace
	tokenLbracket
	tokenLparen
	tokenLt
	tokenLteq
	tokenMod
	tokenMul
	tokenNe
	tokenPipe
	tokenPow
	tokenRbrace
	tokenRbracket
	tokenRparen
	tokenSemicolon
	tokenSub
	tokenTilde
	tokenWhitespace
	tokenFloat
	tokenInteger
	tokenName
	tokenString
	tokenOperator
	tokenBlockBegin
	tokenBlockEnd
	tokenVariableBegin
	tokenVariableEnd
	tokenRawBegin
	tokenRawEnd
	tokenCommentBegin
	tokenCommentEnd
	tokenComment
	tokenLinestatementBegin
	tokenLinestatementEnd
	tokenLinecommentBegin
	tokenLinecommentEnd
	tokenLinecomment
	tokenText // tokenData in jinja2
	tokenInitial
	tokenEOF
	tokenError
	// add a distinct token for bool constants
	tokenBool
)

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

type lexerCfg struct {
	BlockStartString    string
	BlockEndString      string
	VariableStartString string
	VariableEndString   string
	CommentStartString  string
	CommentEndString    string
}

// lexer holds the state of the scanner.
type lexer struct {
	lexerCfg
	name     string // the name of the input; used only for error reports
	filename string // the filename of the input; used only for error reports
	input    string // the string being scanned
	// these are supposed to represent the delims we're looking for, but jigo
	// has a list of possible delims.
	leftDelim  string    // start of action
	rightDelim string    // end of action
	state      stateFn   // the next lexing function to enter
	pos        Pos       // current position in the input
	start      Pos       // start position of this item
	width      Pos       // width of last rune read from input
	lastPos    Pos       // position of most recent item returned by nextItem
	items      chan item // channel of scanned items
	// we will need a more sophisticated delim stack to parse jigo
	//parenDepth int       // nesting depth of ( ) exprs
}

const eof = -1

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// lineNumber reports which line we're on, based on the position of
// the previous item returned by nextItem. Doing it this way
// means we don't have to worry about peek double counting.
func (l *lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.lastPos], "\n")
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{tokenError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next item from the input.
func (l *lexer) nextItem() item {
	item := <-l.items
	l.lastPos = item.pos
	return item
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for l.state = lexText; l.state != nil; {
		l.state = l.state(l)
	}
	close(l.items)
}

// conditionally emit the current text token
func (l *lexer) emitText() {
	if l.pos > l.start {
		l.emit(tokenText)
	}
}

// emit the left delimiter
func (l *lexer) emitLeft() {
	switch l.leftDelim {
	case l.BlockStartString:
		l.emit(tokenBlockBegin)
	case l.VariableStartString:
		l.emit(tokenVariableBegin)
	}
}

// emit the right delimiter
func (l *lexer) emitRight() {
	switch l.rightDelim {
	case l.BlockEndString:
		l.emit(tokenBlockEnd)
	case l.VariableEndString:
		l.emit(tokenVariableEnd)
	}
}

// atTerminator reports whether the input is at valid termination character to
// appear after an identifier. Breaks .X.Y into two pieces.
func (l *lexer) atTerminator() bool {
	r := l.peek()
	if isSpace(r) || isEndOfLine(r) {
		return true
	}
	// if r is an operator...
	switch r {
	case eof, '.', ',', '|', ':', ')', '(', '+', '/', '~', '{', '}', '-', '%', '*', '=':
		return true
	}

	// is the delimiter next to us?
	if l.pos+1 < Pos(len(l.input)) && strings.HasPrefix(l.input[l.pos+1:], l.rightDelim) {
		return true
	}
	return false
}

// lexText starts off the lexing, and is used as a passthrough for all non-jigo
// syntax areas of the template.
func lexText(l *lexer) stateFn {
	for {
		if l.pos == Pos(len(l.input)) {
			break
		}
		switch l.input[l.pos] {
		case l.BlockStartString[0]:
			if strings.HasPrefix(l.input[l.pos:], l.BlockStartString) {
				l.emitText()
				l.leftDelim = l.BlockStartString
				l.rightDelim = l.BlockEndString
				return lexBlock
			}
			fallthrough
		case l.VariableStartString[0]:
			if strings.HasPrefix(l.input[l.pos:], l.VariableStartString) {
				l.emitText()
				l.leftDelim = l.VariableStartString
				l.rightDelim = l.VariableEndString
				return lexBlock
			}
			fallthrough
		case l.CommentStartString[0]:
			if strings.HasPrefix(l.input[l.pos:], l.CommentStartString) {
				l.emitText()
				return lexComment
			}
			fallthrough
		default:
			if l.next() == eof {
				break
			}
		}
	}
	// Correctly reached EOF.
	l.emitText()
	l.emit(tokenEOF)
	return nil
}

func lexBlock(l *lexer) stateFn {
	l.pos += Pos(len(l.leftDelim))
	l.emitLeft()
	return lexInsideBlock

	return lexText
}

func lexInsideBlock(l *lexer) stateFn {
	for {
		if l.pos == Pos(len(l.input)) {
			return nil
		}
		if strings.HasPrefix(l.input[l.pos:], l.rightDelim) {
			l.pos += Pos(len(l.rightDelim))
			l.emitRight()
			return lexText
		}
		// take the next rune and see what it is
		r := l.next()

		switch {
		case isSpace(r):
			return lexSpace
		case isAlphaNumeric(r):
			return lexIdentifier
		}

		switch r {
		case '+':
			l.emit(tokenAdd)
		case '~':
			l.emit(tokenTilde)
		case '/':
			if l.accept("/") {
				l.emit(tokenFloordiv)
			} else {
				l.emit(tokenDiv)
			}
		case '<':
			if l.accept("=") {
				l.emit(tokenLteq)
			} else {
				l.emit(tokenLt)
			}
		case '>':
			if l.accept("=") {
				l.emit(tokenGteq)
			} else {
				l.emit(tokenGt)
			}
		case '*':
			if l.accept("*") {
				l.emit(tokenPow)
			} else {
				l.emit(tokenMul)
			}
		// TODO: ballancing
		case '(':
			l.emit(tokenLparen)
		case '{':
			l.emit(tokenLbrace)
		case '[':
			l.emit(tokenLbracket)
		case ')':
			l.emit(tokenRparen)
		case '}':
			l.emit(tokenRbrace)
		case ']':
			l.emit(tokenRbracket)
		case '-':
			l.emit(tokenSub)
		}
	}
}

func lexSpace(l *lexer) stateFn {
	for isSpace(l.peek()) {
		l.next()
	}
	l.emit(tokenWhitespace)
	return lexInsideBlock
}

func lexIdentifier(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
			// absorb.
		default:
			l.backup()
			word := l.input[l.start:l.pos]
			if !l.atTerminator() {
				return l.errorf("bad character %#U", r)
			}
			switch word {
			case "true", "false":
				l.emit(tokenBool)
			default:
				l.emit(tokenName)
			}
			return lexInsideBlock
		}
	}
}

func lexComment(l *lexer) stateFn {
	l.pos += Pos(len(l.CommentStartString))
	l.emit(tokenCommentBegin)
	i := strings.Index(l.input[l.pos:], l.CommentEndString)
	if i < 0 {
		return l.errorf("unclosed comment")
	}
	l.pos += Pos(i)
	l.emitText()
	l.pos += Pos(len(l.CommentEndString))
	l.emit(tokenCommentEnd)
	return lexText
}

// -- utils --

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
