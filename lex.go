package jigo

import (
	"bytes"
	"fmt"
	"strings"
)

// Jigo's tokenizer
//
// Jigo's tokenizer is a stateful tokenizer that lexes differently depending
// on what kind of block it is in.  If it is in a var/tag block, it tokenizes
// Jinja template keywords, arithmetic operators, et al.  Lots of this code is
// based on code in text/template & html/template

var (
	JigoOpenTag  = "{%"
	JigoCloseTag = "%}"
	JigoOpenVar  = "{{"
	JigoCloseVar = "}}"
)

type tokenType int

const (
	tokenText         tokenType = iota
	tokenOpenTag                // {%
	tokenCloseTag               // %}
	tokenOpenVar                // {{
	tokenCloseVar               // }}
	tokenOpenComment            // {#
	tokenCloseComment           // #}
	tokenError                  // the value is an error message
	tokenIf
	tokenElse
	tokenElif
	tokenEndif
	tokenFor
	tokenEndfor
	tokenIn
	tokenOr
	tokenAnd
	tokenIs
	tokenNot
	tokenConstant // 'nil', 'none', 'false', 'true'
	tokenString   // " or ' delimited string, w/ backslash escapes
	tokenName     // a name, like a variable or filter or global
	tokenNumber
	tokenPipe
	tokenDot
	tokenChar // for , +, -, etc.
	tokenInclude
	tokenImport
	tokenFrom
	tokenEOF
)

func (t tokenType) Name() string {
	switch t {
	case tokenText:
		return "Text"
	case tokenOpenTag:
		return "OpenTag"
	case tokenCloseTag:
		return "CloseTag"
	case tokenOpenVar:
		return "OpenVar"
	case tokenCloseVar:
		return "CloseVar"
	case tokenOpenComment:
		return "OpenComment"
	case tokenCloseComment:
		return "CloseComment"
	case tokenError:
		return "Error"
	case tokenIf:
		return "If"
	case tokenElse:
		return "Else"
	case tokenElif:
		return "Elif"
	case tokenFor:
		return "For"
	case tokenIn:
		return "On"
	case tokenOr:
		return "Or"
	case tokenAnd:
		return "And"
	case tokenIs:
		return "Is"
	case tokenNot:
		return "Not"
	case tokenConstant:
		return "Constant"
	case tokenString:
		return "String"
	case tokenName:
		return "Name"
	case tokenNumber:
		return "Number"
	case tokenPipe:
		return "Pipe"
	case tokenDot:
		return "Dot"
	case tokenChar:
		return "Char"
	case tokenInclude:
		return "Include"
	case tokenImport:
		return "Import"
	case tokenFrom:
		return "From"
	case tokenEOF:
		return "EOF"
	default:
		return "UnknownTokenType"
	}
}

// represents a single token
type token struct {
	token tokenType
	value []byte
	start int
}

func (t token) String() string {
	return fmt.Sprintf("%d: `%s` (%d)",
		t.token,
		strings.Replace(string(t.value), "\n", "\\n", -1),
		t.start,
	)
}

type tagConfig struct {
	ot  []byte // open/cose tag/var block settings
	ct  []byte
	ov  []byte
	cv  []byte
	otl int
	ctl int
	ovl int
	cvl int
}

type stateFn func(*scanner) stateFn

// keep track of the scanner state
type scanner struct {
	name      string // for error reporting
	input     []byte // input
	tokens    chan *token
	length    int // len of input
	p         int // pointer
	w         int // width to the currently scanning token
	lineno    int
	nextToken *token
	state     stateFn
	tagConfig
}

func newTagConfig() *tagConfig {
	t := &tagConfig{
		ot: []byte(JigoOpenTag),
		ct: []byte(JigoCloseTag),
		ov: []byte(JigoOpenVar),
		cv: []byte(JigoCloseVar),
	}
	t.otl = len(t.ot)
	t.ctl = len(t.ct)
	t.ovl = len(t.ov)
	t.cvl = len(t.cv)
	return t
}

// initialize a new scanner
func newScanner(name string, text []byte) *scanner {
	s := &scanner{
		name:   name,
		input:  text,
		tokens: make(chan *token),
		length: len(text),
		p:      0,
		w:      0,
	}
	s.tagConfig = *newTagConfig()
	go s.scan()
	return s
}

var cmp = bytes.Compare

func (s *scanner) emit(t tokenType) {
	s.tokens <- &token{t, s.input[s.w:s.p], s.w}
	s.w = s.p
}

func (s *scanner) ignore() {
	s.w = s.p
}

func (s *scanner) emitError(message string) {
	s.tokens <- &token{
		tokenError,
		[]byte(fmt.Sprintf("%s: %s at %d", s.name, message, s.p)),
		s.p,
	}
}

func scanComment(s *scanner) stateFn {
	for ; s.p < len(s.input); s.p++ {
		switch s.input[s.p] {
		case '}':
			if s.input[s.p-1] == '#' {
				s.ignore()
				s.w = s.p + 1
				return scanText
			}
		}
	}
	s.emitError("Expected endcomment '#}', found EOF")
	return nil
}

func textToken(s *scanner) {
	if s.p == s.w {
		return
	} else {
		s.emit(tokenText)
	}
}

func tagToken(s *scanner) {
	if s.p == s.w {
		return
	} else if s.p-s.w == 1 {
		s.emit(tokenChar)
	} else {
		tok := string(s.input[s.w:s.p])
		switch tok {
		case "if":
			s.emit(tokenIf)
		case "else":
			s.emit(tokenElse)
		case "elif":
			s.emit(tokenElif)
		case "endif":
			s.emit(tokenEndif)
		case "for":
			s.emit(tokenFor)
		case "endfor":
			s.emit(tokenEndfor)
		case "in":
			s.emit(tokenIn)
		case "or":
			s.emit(tokenOr)
		case "not":
			s.emit(tokenNot)
		case "is":
			s.emit(tokenIs)
		case "include":
			s.emit(tokenInclude)
		case "import":
			s.emit(tokenImport)
		case "from":
			s.emit(tokenFrom)
		default:
			s.emit(tokenName)
		}
	}
}

func scanTag(s *scanner) stateFn {
	for ; s.p < len(s.input); s.p++ {
		switch s.input[s.p] {
		case s.ct[0]:
			if cmp(s.input[s.p:s.p+s.ctl], s.ct) == 0 {
				tagToken(s)
				s.p += s.ctl
				s.emit(tokenCloseTag)
				return scanText
			}
			fallthrough
		case s.cv[0]:
			if cmp(s.input[s.p:s.p+s.cvl], s.cv) == 0 {
				tagToken(s)
				s.p += s.cvl
				s.emit(tokenCloseVar)
				return scanText
			}
			fallthrough
		case ' ', '\t', '\n':
			tagToken(s)
			s.w = s.p + 1
		}
	}
	// check that the last w->p is a closetag of some kind
	s.emit(tokenEOF)
	return nil
}

func scanText(s *scanner) stateFn {
	for ; s.p < len(s.input); s.p++ {
		switch s.input[s.p] {
		case s.ot[0]:
			if cmp(s.input[s.p:s.p+s.otl], s.ot) == 0 {
				textToken(s)
				s.p += s.otl
				s.emit(tokenOpenTag)
				return scanTag
			}
			fallthrough
		case s.ov[0]:
			if cmp(s.input[s.p:s.p+s.ovl], s.ov) == 0 {
				textToken(s)
				s.p += s.ovl
				s.emit(tokenOpenVar)
				return scanTag
			}
			fallthrough
		case '{':
			if s.input[s.p+1] == '#' {
				textToken(s)
				s.p += 2
				s.ignore()
				return scanComment
			}
		}
		continue
	}
	textToken(s)
	s.emit(tokenEOF)
	return nil
}

func (s *scanner) scan() {
	for s.state = scanText; s.state != nil; {
		s.state = s.state(s)
	}
}

// Peek at the next token but do not move forward
func (s *scanner) peek() *token {
	if s.nextToken == nil {
		s.nextToken = s.next()
	}
	return s.nextToken
}

func (s *scanner) next() *token {
	// if nextToken is not nil, we've peeked, so use that value and continue
	if s.nextToken != nil {
		t := s.nextToken
		s.nextToken = nil
		return t
	}
	return <-s.tokens
}
