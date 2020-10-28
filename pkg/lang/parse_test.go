package lang

import (
	"testing"
)

var parser Parser = NewDefaultParser()

func TestVariableAssignment(t *testing.T) {
	tokens := []Token{
		{
			Type: Identifier,
			data: "hello",
		},
		{
			Type: EqualSign,
		},
		{
			Type: String,
			data: "world",
		},
		{
			Type: Semicolon,
		},
	}

	statements, err := parser.Parse(tokens)

	if err != nil {
		t.Error(err)
		return
	}

	if len(statements) != 1 {
		t.Errorf("Expected exactly one statement-- got %d", len(statements))
		return
	}

	statement := statements[0]
	if statement.Type != VariableAssignment {
		t.Errorf("Expected VariableAssignment-- got %s", statement.Type.String())
		return
	}

	data := statement.data.(VariableStatement)
	if data.name != "hello" {
		t.Errorf("Expected \"hello\"-- got %s", data.name)
		return
	}

	if data.value != "world" {
		t.Errorf("Expected \"world\"-- got %v", data.value)
		return
	}
}

func TestColorLiteral(t *testing.T) {
	tokens := []Token{
		{
			Type: Identifier,
			data: "color",
		},
		{
			Type: EqualSign,
		},
		{
			Type: LeftParen,
		},
		{
			Type: Integer,
			data: uint8(12),
		},
		{
			Type: Comma,
		},
		{
			Type: Integer,
			data: uint8(10),
		},
		{
			Type: Comma,
		},
		{
			Type: Integer,
			data: uint8(93),
		},
		{
			Type: Comma,
		},
		{
			Type: RightParen,
		},
		{
			Type: Semicolon,
		},
	}

	statements, err := parser.Parse(tokens)

	if err != nil {
		t.Error(err)
		return
	}

	if len(statements) != 1 {
		t.Errorf("Expected exactly one statement-- got %d", len(statements))
		return
	}

	statement := statements[0]
	if statement.Type != VariableAssignment {
		t.Errorf("Expected VariableAssignment-- got %s", statement.Type.String())
		return
	}

	data := statement.data.(VariableStatement)
	if data.name != "color" {
		t.Errorf("Expected \"color\"-- got %s", data.name)
		return
	}

	value := data.value.(ColorLiteral)
	if value.r != 12 || value.g != 10 || value.b != 93 || value.a != 255 {
		t.Errorf("Expected (12, 10, 93, 255) -- got (%d, %d, %d, %d)", value.r, value.g, value.b, value.a)
		return
	}
}
