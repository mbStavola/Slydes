package lang

import (
	"strings"
	"testing"
)

var lexer Lexer = NewDefaultLexer()

func TestEqualSign(t *testing.T) {
	source := `=`
	reader := strings.NewReader(source)
	tokens, err := lexer.Lex(reader)

	if err != nil {
		t.Error(err)
		return
	}

	if len(tokens) != 1 {
		t.Errorf("Expected exactly one token-- got %d", len(tokens))
		return
	}

	if tokens[0].Type != EqualSign {
		t.Errorf("Expected EqualSign-- got %s", tokens[0].Type.String())
	}
}

func TestComments(t *testing.T) {
	source := `
	=
	# Ignore all of this
	@`

	reader := strings.NewReader(source)
	tokens, err := lexer.Lex(reader)

	if err != nil {
		t.Error(err)
		return
	}

	if len(tokens) != 2 {
		t.Errorf("Expected exactly two tokens-- got %d", len(tokens))
		return
	}

	if tokens[0].Type != EqualSign {
		t.Errorf("Expected EqualSign in position 1-- got %s", tokens[0].Type.String())
		return
	}

	if tokens[1].Type != AtSign {
		t.Errorf("Expected AtSign in position 2-- got %s", tokens[1].Type.String())
	}
}

func TestTextBlock(t *testing.T) {
	source := `---This is one block of text---`

	reader := strings.NewReader(source)
	tokens, err := lexer.Lex(reader)

	if err != nil {
		t.Error(err)
		return
	}

	if len(tokens) != 1 {
		t.Errorf("Expected exactly one token-- got %d", len(tokens))
		return
	}

	if tokens[0].Type != Text {
		t.Errorf("Expected Text in position 1-- got %s", tokens[0].Type.String())
		return
	}

	if tokens[0].data != "This is one block of text" {
		t.Errorf("Expected \"This is one block of text\"-- got \"%s\"", tokens[0].data)
	}
}
