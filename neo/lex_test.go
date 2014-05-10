package jigo

import "testing"

type tokenTest struct {
	typ itemType
	val string
}

var (
	ttCommentBegin = tokenTest{tokenCommentBegin, "{#"}
	ttCommentEnd   = tokenTest{tokenCommentEnd, "#}"}
	ttBlockBegin   = tokenTest{tokenBlockBegin, "{%"}
	ttBlockEnd     = tokenTest{tokenBlockEnd, "%}"}
	ttEOF          = tokenTest{tokenEOF, ""}
)

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
		t.Fatalf("Expected %d tokens, got %d\n", len(tests), len(tokens))
	}
	for i, tok := range tokens {
		test := tests[i]
		if test.typ != tok.typ {
			t.Errorf("Expecting %d token type to be %d, got %d\n", i, test.typ, tok.typ)
		}
		if test.val != tok.val {
			t.Errorf("Expecting %d token val to be `%s`, got `%s`\n", i, test.val, tok.val)
		}
	}
}

func TestLexer(t *testing.T) {
	tester := lextest{t}

	// Testing simple text with no jigo syntax
	tester.Test(
		`Hello, world`,
		[]tokenTest{{tokenText, `Hello, world`}, ttEOF},
	)

	// Testing simple text with single jigo comment
	tester.Test(
		`{# comment #}`,
		[]tokenTest{
			ttCommentBegin,
			{tokenText, " comment "},
			ttCommentEnd,
			ttEOF,
		},
	)

	tester.Test(
		`Hello, {# comment #}World`,
		[]tokenTest{
			{tokenText, "Hello, "},
			ttCommentBegin,
			{tokenText, " comment "},
			ttCommentEnd,
			{tokenText, "World"},
			ttEOF,
		},
	)
}
