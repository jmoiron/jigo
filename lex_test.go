package jigo

import "testing"

type lextest struct {
	input  string
	types  []tokenType
	values []string
	t      *testing.T
	tok    *token
}

func testLexer(input string, types []tokenType, values []string, t *testing.T) {
	lt := lextest{input, types, values, t, nil}
	s := newScanner("test", []byte(input))
	tokens := tokenize(s.tokens)
	for i, tok := range tokens {
		lt.tok = tok
		if i < len(types) && i < len(values) {
			lt.assertType(types[i])
			lt.assertValue(values[i])
		} else {
			t.Errorf("Received too many tokens: %v\n", tok)
		}
	}
}

func (l *lextest) assertType(typ tokenType) {
	if l.tok.token != typ {
		l.t.Errorf("Expecting %s, got %s\n  %s\n", typ.Name(), l.tok.token.Name(), l.input)
	}
}

func (l *lextest) assertValue(val string) {
	if string(l.tok.value) != val {
		l.t.Errorf("Expecting %s, got %s\n  %s\n", val, string(l.tok.value), l.input)
	}
}

func tokenize(ct chan *token) []*token {
	var s *token
	tokens := make([]*token, 0, 50)
	for {
		s = <-ct
		tokens = append(tokens, s)
		if s.token == tokenEOF {
			return tokens
		}
	}
	return tokens
}

func TestLexerBasic(t *testing.T) {
	testLexer(
		`{% if foo %}bar{%else%}baz{% endif %}`,
		[]tokenType{
			tokenOpenTag, tokenIf, tokenName, tokenCloseTag, tokenText,
			tokenOpenTag, tokenElse, tokenCloseTag, tokenText, tokenOpenTag,
			tokenEndif, tokenCloseTag, tokenEOF,
		},
		[]string{"{%", "if", "foo", "%}", "bar", "{%", "else", "%}", "baz", "{%", "endif", "%}", ""},
		t,
	)
	testLexer(
		`{# comment #}{% if foo -%} bar {%- elif baz %} bing{%endif    %}`,
		[]tokenType{
			tokenOpenTag, tokenIf, tokenName, tokenChar, tokenCloseTag, tokenText,
			tokenOpenTag, tokenChar, tokenElif, tokenName, tokenCloseTag, tokenText,
			tokenOpenTag, tokenEndif, tokenCloseTag, tokenEOF,
		},
		[]string{"{%", "if", "foo", "-", "%}", " bar ", "{%", "-", "elif", "baz", "%}",
			" bing", "{%", "endif", "%}", "",
		},
		t,
	)
	testLexer(
		`{% macro foo() %}{% if (foo-1) > 3 * x %}hi{%endif%} {% endmacro %}`,
		[]tokenType{
			tokenOpenTag, tokenMacro, tokenName, tokenChar, tokenChar, tokenCloseTag,
			tokenOpenTag, tokenIf, tokenChar, tokenName, tokenChar, tokenName, tokenChar,
			tokenChar, tokenName, tokenChar, tokenName, tokenCloseTag, tokenText,
			tokenOpenTag, tokenEndif, tokenCloseTag, tokenText, tokenOpenTag,
			tokenEndmacro, tokenCloseTag, tokenEOF,
		},
		[]string{"{%", "macro", "foo", "(", ")", "%}",
			"{%", "if", "(", "foo", "-", "1", ")",
			">", "3", "*", "x", "%}", "hi",
			"{%", "endif", "%}", " ", "{%",
			"endmacro", "%}", "",
		},
		t,
	)
	testLexer(
		`<html>{# ignore {% tags %} in comments ##}</html>`,
		[]tokenType{tokenText, tokenText, tokenEOF},
		[]string{"<html>", "</html>", ""},
		t,
	)
}
