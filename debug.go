package main

import (
	"fmt"
	"github.com/mbStavola/slydes/pkg/lang"
	"github.com/mbStavola/slydes/pkg/types"
	"io"
)

// Wrap the provided stages with verbose loggers
func debugSly(sly lang.Sly) lang.Sly {
	return lang.Sly{
		Lexer:    debugLexer{Lexer: sly.Lexer},
		Parser:   debugParser{Parser: sly.Parser},
		Compiler: debugCompiler{Compiler: sly.Compiler},
	}
}

type debugLexer struct {
	Lexer lang.Lexer
}

func (l debugLexer) Lex(reader io.Reader) ([]lang.Token, error) {
	tokens, err := l.Lexer.Lex(reader)
	if err != nil {
		return tokens, err
	}

	fmt.Printf("Lexing Stage: %v\n", tokens)
	fmt.Println("==================================")

	return tokens, nil
}

type debugParser struct {
	Parser lang.Parser
}

func (p debugParser) Parse(tokens []lang.Token) ([]lang.Statement, error) {
	statements, err := p.Parser.Parse(tokens)
	if err != nil {
		return statements, err
	}

	fmt.Printf("Parsing Stage: %v\n", statements)
	fmt.Println("==================================")

	return statements, nil
}

type debugCompiler struct {
	Compiler lang.Compiler
}

func (c debugCompiler) Compile(statements []lang.Statement) (types.Show, error) {
	show, err := c.Compiler.Compile(statements)
	if err != nil {
		return show, err
	}

	fmt.Printf("Compilation Stage: %v\n", show)
	fmt.Println("==================================")

	return show, nil
}
