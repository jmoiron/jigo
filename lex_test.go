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
		lt.assertType(types[i])
		lt.assertValue(values[i])
	}
}

func (l *lextest) assertType(typ tokenType) {
	if l.tok.token != typ {
		l.t.Errorf("Expecting %s, got %s\n", typ.Name(), l.tok.token.Name())
	}
}

func (l *lextest) assertValue(val string) {
	if string(l.tok.value) != val {
		l.t.Errorf("Expecting %s, got %s\n", val, string(l.tok.value))
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
		`{% if foo %}bar{% else %}baz{% endif %}`,
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
}
