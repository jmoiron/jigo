package jigo

import "testing"

// test basic parsing by rendering with empty contexts
func TestParserBasic(t *testing.T) {
	tests := []struct{ name, body, result string }{
		{"test1", `<html></html>`, `<html></html>`},
		{"test2", `foo {# comment #}bar`, `foo bar`},
	}

	for _, test := range tests {
		tm, err := Parse(test.name, []byte(test.body))
		if err != nil {
			t.Errorf("%s: %s\n", test.name, err)
		}
		res, err := tm.root.Render(context{})
		if err != nil {
			t.Errorf("%s: %s\n", test.name, err)
		}
		if res != test.result {
			t.Errorf("%s: expected:\n`%s`\ngot:\n`%s`\n", test.name, test.result, res)
		}
	}
}

func TestParserSimple(t *testing.T) {
	tests := []struct {
		name, body, result string
		ctx                context
	}{
		{"test1", `{{ foo }}`, "Hello", context{"foo": "Hello"}},
		{"test2", `foo {# comment #} {{ x }}`, "foo  bar", context{"x": "bar"}},
	}

	for _, test := range tests {
		tm, err := Parse(test.name, []byte(test.body))
		if err != nil {
			t.Errorf("%s: %s\n", test.name, err)
		}
		res, err := tm.Render(test.ctx)
		if err != nil {
			t.Errorf("%s: %s\n", test.name, err)
		}
		if res != test.result {
			t.Errorf("%s: expected:\n`%s`\ngot:\n`%s`\n", test.name, test.result, res)
		}
	}
}
