package lang

import (
	"errors"
	"fmt"
)

type StatementType int

const (
	InvalidStatement StatementType = iota

	SlideDecl
	ScopeDecl

	VariableDeclaration

	VariableAssignment
	AttributeAssignment
	MacroAssignment

	WordBlock

	MacroCall
)

func (s StatementType) String() string {
	return []string{
		"InvalidStatement",

		"SlideDecl",
		"ScopeDecl",

		"VariableDeclaration",

		"VariableAssignment",
		"AttributeAssignment",
		"MacroAssignment",

		"WordBlock",
		"MacroCall",
	}[s]
}

type Statement struct {
	Type  StatementType
	token Token
	data  interface{}
}

type VariableDeclStatement struct {
	name      string
	isMutable bool
	value     interface{}
}

type VariableStatement struct {
	name  string
	value interface{}
}

type AttributeStatement struct {
	name  string
	value interface{}
}

type MacroStatement struct {
	name       string
	statements []Statement
}

type VariableReference struct {
	reference string
}

type MacroInvocation struct {
	reference string
}

type ColorLiteral struct {
	r uint8
	g uint8
	b uint8
	a uint8
}

type Parser interface {
	Parse(tokens []Token) ([]Statement, error)
}

type DefaultParser struct{}

func NewDefaultParser() DefaultParser {
	return DefaultParser{}
}

func (pars DefaultParser) Parse(tokens []Token) ([]Statement, error) {
	statements := make([]Statement, 0, 1024)
	muncher := tokenMuncher{tokens: tokens}
	errBundle := newErrorInfoBundle()

	for !muncher.atEnd() {
		statement, err := declaration(&muncher)
		if err != nil && errors.As(err, &ErrorInfo{}) {
			errBundle.Add(err.(ErrorInfo))
			synchronizeFromErrorState(&muncher)
		} else if err != nil {
			return statements, err
		}

		statements = append(statements, statement)
	}

	if errBundle.HasErrors() {
		return statements, errBundle
	}

	return statements, nil
}

func declaration(muncher *tokenMuncher) (Statement, error) {
	if muncher.eatIf(SlideScope) {
		token := muncher.previous()
		return Statement{
			Type:  SlideDecl,
			token: token,
		}, nil
	}

	if muncher.eatIf(SubScope) {
		token := muncher.previous()
		return Statement{
			Type:  ScopeDecl,
			token: token,
		}, nil
	}

	if muncher.eatIf(Text) {
		token := muncher.previous()
		return Statement{
			Type:  WordBlock,
			token: token,
			data:  token.data,
		}, nil
	}

	return assignment(muncher)
}

func assignment(muncher *tokenMuncher) (Statement, error) {
	ty := VariableAssignment
	token := muncher.peek()

	switch token.Type {
	case Let:
		ty = VariableDeclaration
		muncher.eat()

		break
	case Mut:
		ty = VariableDeclaration
		muncher.eat()

		break
	case AtSign:
		ty = AttributeAssignment
		muncher.eat()

		break
	case DollarSign:
		ty = MacroAssignment
		muncher.eat()
	}

	if muncher.eatIf(Identifier) {
		identToken := muncher.previous()

		// Macro call
		if ty == MacroAssignment && muncher.eatIf(LeftParen) {
			if _, err := muncher.tryEat(RightParen); err != nil {
				return Statement{}, err
			}
			if _, err := muncher.tryEat(Semicolon); err != nil {
				return Statement{}, err
			}

			return Statement{
				Type:  MacroCall,
				token: token,
				data: MacroInvocation{
					reference: identToken.data.(string),
				},
			}, nil
		}

		if _, err := muncher.tryEat(EqualSign); err != nil {
			return Statement{}, err
		}

		var data interface{}
		if ty == MacroAssignment {
			if _, err := muncher.tryEat(LeftBrace); err != nil {
				return Statement{}, err
			}

			statements := make([]Statement, 0)
			for !muncher.check(RightBrace) {
				statement, err := declaration(muncher)
				if err != nil {
					return Statement{}, err
				}
				statements = append(statements, statement)
			}

			// Eat closing brace
			muncher.eat()

			data = MacroStatement{
				name:       identToken.data.(string),
				statements: statements,
			}
		} else {
			value, err := colorLiteral(muncher)
			if err != nil {
				return Statement{}, err
			}

			if ty == VariableDeclaration {
				data = VariableDeclStatement{
					name:      identToken.data.(string),
					isMutable: token.Type == Mut,
					value:     value,
				}
			} else if ty == VariableAssignment {
				data = VariableStatement{
					name:  identToken.data.(string),
					value: value,
				}
			} else {
				data = AttributeStatement{
					name:  identToken.data.(string),
					value: value,
				}
			}
		}

		if _, err := muncher.tryEat(Semicolon); err != nil {
			return Statement{}, err
		}

		return Statement{
			Type:  ty,
			token: token,
			data:  data,
		}, nil
	}

	message := fmt.Sprintf("Unexpected token %s", token.Type.String())
	return Statement{}, tokenErrorInfo(token, parsing, message)
}

func colorLiteral(muncher *tokenMuncher) (interface{}, error) {
	if muncher.eatIf(LeftParen) {
		values := []uint8{0, 0, 0, 255}

		if value, err := muncher.tryEat(Integer); err == nil {
			values[0] = value.data.(uint8)
		}

		// We need to eat at least two more comma + ident pairs
		for i := 1; i < 3; i++ {
			if _, err := muncher.tryEat(Comma); err != nil {
				return nil, err
			}

			// TODO(Matt): We should be able to use variables here
			value, err := muncher.tryEat(Integer)
			if err != nil {
				return nil, err
			}

			values[i] = value.data.(uint8)
		}

		// Allow trailing comma
		if muncher.eatIf(Comma) {
			// ... and fourth param if supplied
			if muncher.eatIf(Integer) {
				value := muncher.previous()
				values[3] = value.data.(uint8)
				// Get rid of any trailing comma
				muncher.eatIf(Comma)
			}
		}

		if _, err := muncher.tryEat(RightParen); err != nil {
			return nil, err
		}

		return ColorLiteral{
			r: values[0],
			g: values[1],
			b: values[2],
			a: values[3],
		}, nil
	}

	return value(muncher)
}

func value(muncher *tokenMuncher) (interface{}, error) {
	token := muncher.peek()

	if token.Type == String {
		muncher.eat()
		return token.data.(string), nil
	} else if token.Type == Integer {
		muncher.eat()
		return token.data.(uint8), nil
	} else if token.Type == Identifier {
		muncher.eat()
		return VariableReference{reference: token.data.(string)}, nil
	}

	return nil, tokenErrorInfo(token, parsing, "Expected value")
}

func synchronizeFromErrorState(muncher *tokenMuncher) {
	muncher.eat()

	for !muncher.atEnd() {
		if muncher.previous().Type == Semicolon {
			return
		}

		muncher.eat()
	}
}

type tokenMuncher struct {
	tokens  []Token
	current int
}

func (tm *tokenMuncher) atEnd() bool {
	return tm.current >= len(tm.tokens)
}

func (tm *tokenMuncher) eatIf(expected TokenType) bool {
	if tm.check(expected) {
		tm.eat()
		return true
	}

	return false
}

func (tm *tokenMuncher) tryEat(expected TokenType) (Token, error) {
	token := tm.peek()
	if token.Type == expected {
		tm.eat()
		return token, nil
	}

	message := fmt.Sprintf("Expected %s, but was %s", expected.String(), token.Type.String())
	return Token{}, tokenErrorInfo(token, parsing, message)
}

func (tm *tokenMuncher) previous() Token {
	return tm.peekN(-1)
}

func (tm *tokenMuncher) peek() Token {
	return tm.peekN(0)
}

func (tm *tokenMuncher) peekNext() Token {
	return tm.peekN(1)
}

func (tm *tokenMuncher) peekN(n int) Token {
	if tm.current+n >= len(tm.tokens) {
		return Token{Type: EOF}
	}

	return tm.tokens[tm.current+n]
}

func (tm *tokenMuncher) check(expected TokenType) bool {
	if tm.atEnd() {
		return false
	}

	actual := tm.peek()
	return expected == actual.Type
}

func (tm *tokenMuncher) eat() Token {
	if !tm.atEnd() {
		tm.current++
	}

	return tm.previous()
}
