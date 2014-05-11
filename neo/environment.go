package jigo

import "errors"

type Environment struct {
	// The string marking the start of a block.  Defaults to `{%`.
	BlockStartString string
	// The string marking the end of a block.  Defaults to `%}`.
	BlockEndString string
	// The string marking the start of a var statement. Defaults to `{{`.
	VariableStartString string
	// The string marking the end of a var statement. Defaults to `}}`.
	VariableEndString string
	// The string marking the start of a comment.  Defaults to `{#`.
	CommentStartString string
	// The string marking the end of a comment.  Defaults to `#}`.
	CommentEndString string
	// If true, first newline after a block is removed.  Default false.
	TrimBlocks bool
	// If true, leading whitespace is stripped from the start of a line to a block.  Default false.
	LstripBlocks bool
	// If true, html auto-escaping is enabled by default for all var output.
	AutoEscape bool
	// Should the loader attempt to auto reload.
	AutoReload bool

	// -- Will not support --
	// I've decided not to support line statements and line comments, they're unnecessary.
	// LineStatementPrefix string
	// LineCommentPrefix   string
	// For simplicity, trailing newlines will always be kept.
	// KeepTrailingNewline bool
	// The sequence that starts a newline.  Only allow `\n`.
	// NewlineSequence string

	// -- TBI --

	// finalize ~ A callable that can be used to process the result of a var expr
	// as it is being output.  For example can convert `nil` to "".  I think since
	// Go is statically typed it's unlikely we'll have use for this

	// filters ~ a mapping of names to functions for use in | filters.  Mandira
	// already supports these, so these should not be too difficult.

	// tests ~ a mapping of functions for use with the is operator;  will have to define
	// a TestFunc interface.

	// Global variables to pass to every template.  Shadowed by actual local contexts.
	Globals map[string]interface{}
	// extensions ~ not sure these are easily doable with Go.

	// loader ~ loaders can customize where templates come from, so you can
	// refer to template paths by say memcached key for an MC loader or fs path
	// anchored at some root for an FS loader.  These are cool, but will start
	// with just a built-in FS loader.

	// cache ~ cache of recently parsed templates.  []Ast?

	// cache_size ~ LRU of recently parsed templates, defaults to 50..
	// will start by keeping all parsed templates, keyed by path & env key

	// bytecode_cache ~ we're going to do an AST cache which will basically
	// just be a Gobbed AST.
}

// sanityCheck checks an environment for possible improper configurations.
func (e Environment) sanityCheck() error {
	if e.CommentStartString == e.BlockStartString || e.CommentStartString == e.VariableStartString || e.BlockStartString == e.VariableStartString {
		return errors.New("BlockStartString, VariableBlockString, and CommentStartString must be distinct.")
	}
	return nil
}

func NewEnvironment() *Environment {
	return &Environment{
		BlockStartString:    "{%",
		BlockEndString:      "%}",
		VariableStartString: "{{",
		VariableEndString:   "}}",
		CommentStartString:  "{#",
		CommentEndString:    "#}",
		Globals:             make(map[string]interface{}),
	}
}

// lex returns a new lexer for some source.
func (e *Environment) lex(source, name, filename string) *lexer {
	cfg := lexerCfg{
		BlockStartString:    e.BlockStartString,
		BlockEndString:      e.BlockEndString,
		VariableStartString: e.VariableStartString,
		VariableEndString:   e.VariableEndString,
		CommentStartString:  e.CommentStartString,
		CommentEndString:    e.CommentEndString,
	}
	l := &lexer{
		lexerCfg:   cfg,
		name:       name,
		filename:   filename,
		input:      source,
		leftDelim:  cfg.BlockStartString,
		rightDelim: cfg.BlockEndString,
		items:      make(chan item),
		delimStack: make([]rune, 0, 10),
	}
	go l.run()
	return l
}
