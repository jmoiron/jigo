package jigo

import "testing"

type m map[string]interface{}

func TestSimpleEval(t *testing.T) {
	fixtures := []struct {
		name, body string
		context    m
		result     string
	}{
		{"Hello, World", "Hello, World", m{}, "Hello, World"},
		{"Comment", "Hello, {# comment #}World", m{}, "Hello, World"},
		{"Variable", "Hello {{ name }}", m{"name": "Jason"}, "Hello Jason"},
		{
			"Variable Unicode",
			"{{ greeting }}, {{name}}",
			m{"greeting": "おはようございます", "name": "山田くん"},
			"おはようございます, 山田くん",
		},
		{"Math", "{{ 1 + 2 }}", m{}, "3"},
		{"Cat", `{{ "foo" + "bar" }}`, m{}, "foobar"},
		{"Cat Var", `{{ foo + "bar" }}`, m{"foo": "baz"}, "bazbar"},
		//{"CoerceConcat", `{{ 1 ~ "1" }}`, m{}, "11"},
		{
			"Conditional",
			`{% if true %}true{% else %}false{% endif %}`,
			m{},
			"true",
		},
		{"Conditional Var",
			`{% if var %}true{% else %}false{% endif %}`,
			m{"var": false}, "false"},
	}

	// use defaults
	e := NewEnvironment()

	for _, fixture := range fixtures {
		template, err := e.ParseString(fixture.body, fixture.name, "temp")
		if err != nil {
			t.Error(err)
			continue
		}
		result, err := template.Render(fixture.context)
		if err != nil {
			t.Errorf("Test %s: unexpected error %s\n", fixture.name, err)
			continue
		}
		if result != fixture.result {
			t.Errorf("Test %s: Expected:\n`%s`\nGot:\n`%s`\n", fixture.name, fixture.result, result)
		}
	}

	/*
		tester.Test(
			`{{ 1 + 2 }}`,
			parseTest{nodeTypes: []NodeType{NodeVar}},
		)

		tester.Test(
			`{{ "foo" + "bar" }}`,
			parseTest{nodeTypes: []NodeType{NodeVar}},
		)

		tester.Test(
			`{{ 1 + 2 * 3 + 4}}`,
			parseTest{nodeTypes: []NodeType{NodeVar}},
		)

		tester.Test(
			`{{ {"hello": "world", 1: "one"} }}`,
			parseTest{nodeTypes: []NodeType{NodeVar}},
		)

		tester.Test(
			`{% set foo = 1 %}`,
			parseTest{nodeTypes: []NodeType{NodeSet}},
		)

		tester.Test(
			`{% if true %}something{% else %}something else{% endif %}`,
			parseTest{nodeTypes: []NodeType{NodeIf}},
		)
	*/
}
