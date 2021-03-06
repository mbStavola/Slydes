package lang

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type TokenType int

const (
	InvalidToken TokenType = iota
	EOF
	Skip

	// Single Character
	LeftParen
	RightParen
	LeftBrace
	RightBrace
	Semicolon
	Colon
	AtSign
	DollarSign
	EqualSign
	Comma
	Dot

	// Keywords
	Let
	Mut
	Macro
	Slide
	Block
	Self

	// Literals
	Identifier

	Text
	String
	Integer
)

func (t TokenType) String() string {
	return []string{
		"InvalidToken",
		"EOF",
		"Skip",

		"LeftParen",
		"RightParen",
		"LeftBrace",
		"RightBrace",
		"Semicolon",
		"Colon",
		"AtSign",
		"DollarSign",
		"EqualSign",
		"Comma",
		"Dot",

		"Let",
		"Mut",
		"Macro",
		"Slide",
		"Block",
		"Self",

		"Identifier",

		"Text",
		"String",
		"Integer",
	}[t]
}

type Token struct {
	Type   TokenType
	lexeme rune
	line   uint
	data   interface{}
}

type Lexer interface {
	Lex(reader io.Reader) ([]Token, error)
}

type DefaultLexer struct{}

func NewDefaultLexer() DefaultLexer {
	return DefaultLexer{}
}

func (lex DefaultLexer) Lex(reader io.Reader) ([]Token, error) {
	muncher := newRuneMuncher(reader)
	errBundle := newErrorInfoBundle()

	tokens := make([]Token, 0, 1024)

	for !muncher.atEnd() {
		token, err := processRune(muncher)
		if err != nil && errors.As(err, &ErrorInfo{}) {
			errBundle.Add(err.(ErrorInfo))
		} else if err != nil {
			return tokens, err
		}

		if token.Type == Skip {
			continue
		}

		tokens = append(tokens, token)
	}

	if errBundle.HasErrors() {
		return tokens, errBundle
	}

	return tokens, nil
}

func processRune(muncher *runeMuncher) (Token, error) {
	char, _, err := muncher.ReadRune()
	if err != nil {
		return Token{}, err
	}

	switch char {
	case '#':
		_, _, _ = muncher.ReadLine()
		muncher.newLine()
		return Token{Type: Skip}, nil

	case '(':
		return Token{
			Type:   LeftParen,
			line:   muncher.line,
			lexeme: char,
		}, nil

	case ')':
		return Token{
			Type:   RightParen,
			line:   muncher.line,
			lexeme: char,
		}, nil

	case '{':
		return Token{
			Type:   LeftBrace,
			line:   muncher.line,
			lexeme: char,
		}, nil

	case '}':
		return Token{
			Type:   RightBrace,
			line:   muncher.line,
			lexeme: char,
		}, nil

	case '@':
		return Token{
			Type:   AtSign,
			line:   muncher.line,
			lexeme: char,
		}, nil

	case '$':
		return Token{
			Type:   DollarSign,
			line:   muncher.line,
			lexeme: char,
		}, nil

	case '=':
		return Token{
			Type:   EqualSign,
			line:   muncher.line,
			lexeme: char,
		}, nil

	case ';':
		return Token{
			Type:   Semicolon,
			line:   muncher.line,
			lexeme: char,
		}, nil

	case ':':
		return Token{
			Type:   Colon,
			line:   muncher.line,
			lexeme: char,
		}, nil

	case ',':
		return Token{
			Type:   Comma,
			line:   muncher.line,
			lexeme: char,
		}, nil

	case '.':
		return Token{
			Type:   Dot,
			line:   muncher.line,
			lexeme: char,
		}, nil

	case 'l':
		if ok, err := muncher.eatKeyword("et"); err == io.EOF {
			return Token{}, lexemeErrorInfo(muncher.line, char, "Unexpected end of file")
		} else if err != nil {
			return Token{}, err
		} else if ok {
			return Token{
				Type:   Let,
				line:   muncher.line,
				lexeme: char,
			}, nil
		}

	case 'm':
		if ok, err := muncher.eatKeyword("ut"); err == io.EOF {
			return Token{}, lexemeErrorInfo(muncher.line, char, "Unexpected end of file")
		} else if err != nil {
			return Token{}, err
		} else if ok {
			return Token{
				Type:   Mut,
				line:   muncher.line,
				lexeme: char,
			}, nil
		} else if ok, err := muncher.eatKeyword("acro"); err == io.EOF {
			return Token{}, lexemeErrorInfo(muncher.line, char, "Unexpected end of file")
		} else if err != nil {
			return Token{}, err
		} else if ok {
			return Token{
				Type:   Macro,
				line:   muncher.line,
				lexeme: char,
			}, nil
		}

	case 's':
		if ok, err := muncher.eatKeyword("lide"); err == io.EOF {
			return Token{}, lexemeErrorInfo(muncher.line, char, "Unexpected end of file")
		} else if err != nil {
			return Token{}, err
		} else if ok {
			return Token{
				Type:   Slide,
				line:   muncher.line,
				lexeme: char,
			}, nil
		} else if ok, err := muncher.eatKeyword("elf"); err == io.EOF {
			return Token{}, lexemeErrorInfo(muncher.line, char, "Unexpected end of file")
		} else if err != nil {
			return Token{}, err
		} else if ok {
			return Token{
				Type:   Self,
				line:   muncher.line,
				lexeme: char,
			}, nil
		}

	case 'b':
		if ok, err := muncher.eatKeyword("lock"); err == io.EOF {
			return Token{}, lexemeErrorInfo(muncher.line, char, "Unexpected end of file")
		} else if err != nil {
			return Token{}, err
		} else if ok {
			return Token{
				Type:   Block,
				line:   muncher.line,
				lexeme: char,
			}, nil
		}

	case '-':
		if chars, err := muncher.Peek(2); err == io.EOF {
			return Token{}, lexemeErrorInfo(muncher.line, char, "Unexpected end of file")
		} else if err != nil {
			return Token{}, err
		} else if string(chars[:]) != "--" {
			return Token{}, lexemeErrorInfo(muncher.line, char, "Malformed text block")
		}

		// Eat the starting dashes
		if err = muncher.eatN(2); err != nil {
			return Token{}, err
		}

		text := strings.Builder{}
		dashCounter := 0

		// Read runes until we encounter three dashes in a row
		err := muncher.eatWhile(func(char rune) bool {
			if char == '-' && dashCounter == 2 {
				return false
			} else if char == '-' {
				dashCounter++
				return true
			} else if dashCounter > 0 {
				// If we've seen some number of dashes that is less than
				// three, write them to the text buffer and reset the count
				dashes := bytes.Repeat([]byte("-"), dashCounter)
				text.Write(dashes)
				dashCounter = 0
			}

			text.WriteRune(char)

			return true
		})

		if err != nil {
			return Token{}, err
		}

		return Token{
			Type:   Text,
			line:   muncher.line,
			lexeme: char,
			data:   text.String(),
		}, nil

	case '"':
		str, err := muncher.ReadString('"')
		if err == io.EOF {
			return Token{}, simpleErrorInfo(muncher.line, "Unterminated String")
		} else if err != nil {
			return Token{}, err
		}

		return Token{
			Type:   String,
			line:   muncher.line,
			lexeme: char,
			// Cut off the dangling " in the string
			data: str[:len(str)-1],
		}, nil

	case '0':
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		num := strings.Builder{}
		num.WriteRune(char)

		err := muncher.eatWhile(func(char rune) bool {
			if !unicode.IsNumber(char) {
				muncher.UnreadRune()
				return false
			}

			num.WriteRune(char)

			return true
		})

		if err != nil {
			return Token{}, err
		}

		data, err := strconv.ParseUint(num.String(), 10, 8)
		if err != nil {
			return Token{}, err
		}

		return Token{
			Type:   Integer,
			line:   muncher.line,
			lexeme: char,
			data:   uint8(data),
		}, nil

	case ' ', '\t', '\r':
		return Token{Type: Skip}, nil

	case '\n':
		muncher.newLine()
		return Token{Type: Skip}, nil
	}

	if unicode.IsLetter(char) {
		ident := strings.Builder{}
		ident.WriteRune(char)

		err := muncher.eatWhile(func(char rune) bool {
			if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
				muncher.UnreadRune()
				return false
			}

			ident.WriteRune(char)

			return true
		})

		if err != nil {
			return Token{}, err
		}

		return Token{
			Type:   Identifier,
			line:   muncher.line,
			lexeme: char,
			data:   ident.String(),
		}, nil
	}

	return Token{}, lexemeErrorInfo(muncher.line, char, "Unexpected character")
}

type runeMuncher struct {
	line uint
	*bufio.Reader
}

func newRuneMuncher(reader io.Reader) *runeMuncher {
	r := new(runeMuncher)

	r.line = 1
	r.Reader = bufio.NewReader(reader)

	return r
}

func (r *runeMuncher) atEnd() bool {
	_, err := r.Peek(1)
	return err == io.EOF
}

func (r *runeMuncher) newLine() {
	r.line++
}

// Helper function to conditionally eat a lexeme if it matches
// the expected rune
func (r *runeMuncher) eatIf(expected rune) (bool, error) {
	if actual, _, err := r.ReadRune(); err != nil {
		// TODO(Matt): Would it be more correct to unread the rune here?
		return false, err
	} else if actual == expected {
		return true, nil
	}

	return false, r.UnreadRune()
}

// Helper function to discard a specified number of runes
func (r *runeMuncher) eatN(n int) error {
	for i := 0; i < n; i++ {
		if _, _, err := r.ReadRune(); err != nil {
			return err
		}
	}

	return nil
}

func (r *runeMuncher) eatWhile(callback func(rune) bool) error {
	for char, _, err := r.ReadRune(); ; char, _, err = r.ReadRune() {
		if err != nil {
			return err
		}

		if shouldContinue := callback(char); !shouldContinue {
			break
		}
	}

	return nil
}

func (r *runeMuncher) eatKeyword(rest string) (bool, error) {
	restLen := len(rest)

	if chars, err := r.Peek(restLen); err != nil {
		return false, err
	} else if string(chars[:]) == rest {
		r.eatN(restLen)
		return true, nil
	}

	return false, nil
}
