package lang

import (
	"io"
	"strings"

	"github.com/mbStavola/slydes/pkg/types"
)

// This type represents a three phase "compiler" for
// the Sly language which can produce an instance of Show
type Sly struct {
	Lexer    Lexer
	Parser   Parser
	Compiler Compiler
}

// Construct a new Sly instance with working defaults
//
// Since the default Sly is both small and devoid of
// mutable state, we return by value over a reference
func NewSly() Sly {
	lexer := NewDefaultLexer()
	parser := NewDefaultParser()
	compiler := NewDefaultCompiler()
	return Sly{Lexer: lexer, Parser: parser, Compiler: compiler}
}

func (sly Sly) ReadSlideShowString(source string) (types.Show, error) {
	reader := strings.NewReader(source)
	return sly.ReadSlideShow(reader)
}

func (sly Sly) ReadSlideShow(reader io.Reader) (types.Show, error) {
	tokens, err := sly.Lexer.Lex(reader)
	if err != nil {
		return types.Show{}, err
	}

	statements, err := sly.Parser.Parse(tokens)
	if err != nil {
		return types.Show{}, err
	}

	return sly.Compiler.Compile(statements)
}
