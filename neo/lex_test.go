package jigo

import (
	"fmt"
	"testing"
)

type tokenTest struct {
	typ itemType
	val string
}

var (
	ttCommentBegin  = tokenTest{tokenCommentBegin, "{#"}
	ttCommentEnd    = tokenTest{tokenCommentEnd, "#}"}
	ttBlockBegin    = tokenTest{tokenBlockBegin, "{%"}
	ttBlockEnd      = tokenTest{tokenBlockEnd, "%}"}
	ttVariableBegin = tokenTest{tokenVariableBegin, "{{"}
	ttVariableEnd   = tokenTest{tokenVariableEnd, "}}"}
	ttEOF           = tokenTest{tokenEOF, ""}
	ttSub           = tokenTest{tokenSub, "-"}
	ttAdd           = tokenTest{tokenAdd, "+"}
	ttDiv           = tokenTest{tokenDiv, "/"}
	ttComma         = tokenTest{tokenComma, ","}
	ttPipe          = tokenTest{tokenPipe, "|"}
	ttLparen        = tokenTest{tokenLparen, "("}
	ttRparen        = tokenTest{tokenRparen, ")"}
	ttLbrace        = tokenTest{tokenLbrace, "{"}
	ttRbrace        = tokenTest{tokenRbrace, "}"}
	ttLbracket      = tokenTest{tokenLbracket, "["}
	ttRbracket      = tokenTest{tokenRbracket, "]"}
	ttColon         = tokenTest{tokenColon, ":"}
	ttMul           = tokenTest{tokenMul, "*"}
	ttPow           = tokenTest{tokenPow, "**"}
	ttFloordiv      = tokenTest{tokenFloordiv, "//"}
	ttGt            = tokenTest{tokenGt, ">"}
	ttLt            = tokenTest{tokenLt, "<"}
	ttGteq          = tokenTest{tokenGteq, ">="}
	ttLteq          = tokenTest{tokenLteq, "<="}
	ttEq            = tokenTest{tokenEq, "="}
	ttEqEq          = tokenTest{tokenEqEq, "=="}
	sp              = tokenTest{tokenWhitespace, " "}
)

func (t tokenTest) String() string {
	return `"` + t.val + `"`
}

func tn(name string) tokenTest {
	return tokenTest{tokenName, name}
}

func tt(name string) tokenTest {
	return tokenTest{tokenText, name}
}

func ts(value string) tokenTest {
	return tokenTest{tokenString, value}
}

func tokenize(l *lexer) []item {
	items := make([]item, 0, 50)
	for t := range l.items {
		items = append(items, t)
		// fmt.Printf("%#v\n", t)
	}
	return items
}

type lextest struct{ *testing.T }

func (lt *lextest) Test(input string, tests []tokenTest) {
	t := lt.T
	e := NewEnvironment()
	l := e.lex(input, "test", "test.jigo")
	tokens := tokenize(l)
	if len(tokens) != len(tests) {
		t.Errorf("Expected %d tokens, got %d\n", len(tests), len(tokens))
	}
	for i, tok := range tokens {
		if i >= len(tests) {
			return
		}
		test := tests[i]
		if test.typ != tok.typ {
			fmt.Printf("tokens: %v\ntests:  %v\n", tokens, tests)
			t.Errorf("Expecting %d token type to be %d, got %d\n", i, test.typ, tok.typ)
		}
		if tok.typ != tokenWhitespace && test.val != tok.val {
			fmt.Printf("tokens: %v\ntests:  %v\n", tokens, tests)
			t.Errorf("Expecting %d token val to be `%s`, got `%s`\n", i, test.val, tok.val)
		}
	}
}

func TestLexer(t *testing.T) {
	tester := lextest{t}

	// Testing simple text with no jigo syntax
	tester.Test(
		`Hello, world`,
		[]tokenTest{tt(`Hello, world`), ttEOF},
	)

	// Testing simple text with single jigo comment
	tester.Test(
		`{# comment #}`,
		[]tokenTest{ttCommentBegin, tt(" comment "), ttCommentEnd, ttEOF},
	)

	tester.Test(
		`Hello, {# comment #}World`,
		[]tokenTest{tt("Hello, "), ttCommentBegin, tt(" comment "), ttCommentEnd, tt("World"), ttEOF},
	)

	tester.Test(
		`{{ foo }}`,
		[]tokenTest{ttVariableBegin, sp, tn("foo"), sp, ttVariableEnd, ttEOF},
	)

	tester.Test(
		`{{ (a - b) + c }}`,
		[]tokenTest{
			ttVariableBegin, sp, ttLparen, tn("a"), sp, ttSub, sp, tn("b"), ttRparen, sp,
			ttAdd, sp, tn("c"), sp, ttVariableEnd, ttEOF,
		},
	)

	tester.Test(
		`Hello.  {% if true %}World{% else %}Nobody{% endif %}`,
		[]tokenTest{
			tt("Hello.  "), ttBlockBegin, sp, tn("if"), sp,
			{tokenBool, "true"}, sp, ttBlockEnd, tt("World"), ttBlockBegin, sp, tn("else"), sp,
			ttBlockEnd, tt("Nobody"), ttBlockBegin, sp, tn("endif"), sp, ttBlockEnd, ttEOF,
		},
	)

	tester.Test(
		`<html>{# ignore {% tags %} in comments ##}</html>`,
		[]tokenTest{
			tt("<html>"), ttCommentBegin, tt(" ignore {% tags %} in comments #"),
			ttCommentEnd, tt("</html>"), ttEOF,
		},
	)

	tester.Test(
		`{# comment #}{% if foo -%} bar {%- elif baz %} bing{%endif    %}`,
		[]tokenTest{
			ttCommentBegin, tt(" comment "), ttCommentEnd, ttBlockBegin, sp, tn("if"), sp,
			tn("foo"), sp, ttSub, ttBlockEnd, tt(" bar "), ttBlockBegin, ttSub, sp, tn("elif"),
			sp, tn("baz"), sp, ttBlockEnd, tt(" bing"), ttBlockBegin, tn("endif"), sp,
			ttBlockEnd, ttEOF,
		},
	)

	// test a big mess of tokens including single and double character tokens
	tester.Test(
		`{{ +--+ /+//,|*/**=>>=<=< == }}`,
		[]tokenTest{
			ttVariableBegin, sp, ttAdd, ttSub, ttSub, ttAdd, sp, ttDiv, ttAdd, ttFloordiv,
			ttComma, ttPipe, ttMul, ttDiv, ttPow, ttEq, ttGt, ttGteq, ttLteq, ttLt, sp, ttEqEq,
			sp, ttVariableEnd, ttEOF,
		},
	)

	tester.Test(
		`{{ ([{}]()) }}`,
		[]tokenTest{
			ttVariableBegin, sp,
			ttLparen, ttLbracket, ttLbrace, ttRbrace, ttRbracket, ttLparen, ttRparen, ttRparen, sp,
			ttVariableEnd, ttEOF,
		},
	)

	tester.Test(
		`{{ ([{]) }}`,
		[]tokenTest{
			ttVariableBegin, sp, ttLparen, ttLbracket, ttLbrace,
			{tokenError, "Imbalanced delimiters, expected }, got ]"},
		},
	)

	// Test that unballancing delimiters takes precedence over closing the block, ie.
	// that the `}}` closing of the map doesn't close the var tag.
	tester.Test(
		`{{ ({a:b, {a:b}}) }}`,
		[]tokenTest{
			ttVariableBegin, sp, ttLparen, ttLbrace, tn("a"), ttColon, tn("b"), ttComma, sp,
			ttLbrace, tn("a"), ttColon, tn("b"), ttRbrace, ttRbrace, ttRparen, sp, ttVariableEnd, ttEOF,
		},
	)

	tester.Test(
		`{{ "Hello, " + "World" }}`,
		[]tokenTest{
			ttVariableBegin, sp, ts("Hello, "), sp, ttAdd, sp, ts("World"), sp, ttVariableEnd, ttEOF,
		},
	)

	st := []tokenTest{ttVariableBegin, sp, ts(`Hello, "World"`), sp, ttVariableEnd, ttEOF}
	tester.Test("{{ `Hello, \"World\"` }}", st)
	tester.Test(`{{ "Hello, \"World\"" }}`, st)
}
