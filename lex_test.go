package jigo

import (
	"fmt"
	"testing"
)

var lt1 = []byte(`
{% import foo %}
{% if bar %}
	Bye
{% else -%}
	Hi
{%- endif %}

{# comment #}
<html>`)

func TestLexer(t *testing.T) {
	s := newScanner("test", lt1)
	for tok := s.next(); tok.token != tokenEOF; tok = s.next() {
		fmt.Printf("%s\n", tok)
	}
}
