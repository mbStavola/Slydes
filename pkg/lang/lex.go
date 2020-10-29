package lang

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type TokenType int

const (
	InvalidToken TokenType = iota
	EOF

	// Single Character
	LeftParen
	RightParen
	LeftBrace
	RightBrace
	Semicolon
	AtSign
	DollarSign
	EqualSign
	Comma

	// Special
	SlideScope
	SubScope

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

		"LeftParen",
		"RightParen",
		"LeftBrace",
		"RightBrace",
		"Semicolon",
		"AtSign",
		"DollarSign",
		"EqualSign",
		"Comma",

		"SlideScope",
		"SubScope",

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
	bufReader := bufio.NewReader(reader)
	tokens := make([]Token, 0, 1024)
	var line uint = 1

	for char, _, err := bufReader.ReadRune(); err != io.EOF; char, _, err = bufReader.ReadRune() {
		if err != nil {
			return tokens, err
		}

		switch char {
		case '#':
			_, _, _ = bufReader.ReadLine()
			line += 1

			break
		case '[':
			ty := SlideScope
			if shouldEat, err := eatIf(bufReader, '['); err == io.EOF {
				return tokens, lexemeErrorInfo(line, char, "Unexpected end of file")
			} else if err != nil {
				return tokens, err
			} else if shouldEat {
				ty = SubScope
			}

			title, err := bufReader.ReadString(']')
			if err == io.EOF {
				return tokens, lexemeErrorInfo(line, char, "Unterminated scope")
			} else if err != nil {
				return tokens, err
			}

			if shouldEat, err := eatIf(bufReader, ']'); err == io.EOF {
				return tokens, lexemeErrorInfo(line, char, "Unexpected end of file")
			} else if err != nil {
				return tokens, err
			} else if ty == SubScope && !shouldEat {
				return tokens, lexemeErrorInfo(line, ']', "Subscope expected closing ']'")
			} else if ty == SlideScope && shouldEat {
				return tokens, lexemeErrorInfo(line, ']', "Dangling scope end")
			}

			tokens = append(tokens, Token{
				Type:   ty,
				line:   line,
				lexeme: char,
				// Cut off the dangling ] in the scope title
				data: title[:len(title)-1],
			})

			break
		case '(':
			tokens = append(tokens, Token{
				Type:   LeftParen,
				line:   line,
				lexeme: char,
			})

			break
		case ')':
			tokens = append(tokens, Token{
				Type:   RightParen,
				line:   line,
				lexeme: char,
			})

			break
		case '{':
			tokens = append(tokens, Token{
				Type:   LeftBrace,
				line:   line,
				lexeme: char,
			})

			break
		case '}':
			tokens = append(tokens, Token{
				Type:   RightBrace,
				line:   line,
				lexeme: char,
			})

			break
		case '@':
			tokens = append(tokens, Token{
				Type:   AtSign,
				line:   line,
				lexeme: char,
			})

			break
		case '$':
			tokens = append(tokens, Token{
				Type:   DollarSign,
				line:   line,
				lexeme: char,
			})

			break
		case '=':
			tokens = append(tokens, Token{
				Type:   EqualSign,
				line:   line,
				lexeme: char,
			})

			break
		case ';':
			tokens = append(tokens, Token{
				Type:   Semicolon,
				line:   line,
				lexeme: char,
			})

		case ',':
			tokens = append(tokens, Token{
				Type:   Comma,
				line:   line,
				lexeme: char,
			})

			break
		case '-':
			if chars, err := bufReader.Peek(2); err == io.EOF {
				return tokens, lexemeErrorInfo(line, char, "Unexpected end of file")
			} else if err != nil {
				return tokens, err
			} else if string(chars[:]) != "--" {
				return tokens, lexemeErrorInfo(line, char, "Malformed text block")
			}

			// Eat the starting dashes
			if err = eatN(bufReader, 2); err != nil {
				return tokens, err
			}

			text := strings.Builder{}
			dashCounter := 0

			// Read runes until we encounter three dashes in a row
			for char, _, err := bufReader.ReadRune(); ; char, _, err = bufReader.ReadRune() {
				if char == '-' && dashCounter == 2 {
					break
				} else if char == '-' {
					dashCounter += 1
					continue
				} else if dashCounter > 0 {
					// If we've seen some number of dashes that is less than
					// three, write them to the text buffer and reset the count
					dashes := bytes.Repeat([]byte("-"), dashCounter)
					text.Write(dashes)
					dashCounter = 0
				}

				if err != nil {
					return tokens, err
				}

				text.WriteRune(char)
			}

			tokens = append(tokens, Token{
				Type:   Text,
				line:   line,
				lexeme: char,
				data:   text.String(),
			})

			break
		case '"':
			str, err := bufReader.ReadString('"')
			if err == io.EOF {
				return tokens, simpleErrorInfo(line, "Unterminated String")
			} else if err != nil {
				return tokens, err
			}

			tokens = append(tokens, Token{
				Type:   String,
				line:   line,
				lexeme: char,
				// Cut off the dangling " in the string
				data: str[:len(str)-1],
			})

			break
		case '0':
		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			num := strings.Builder{}
			num.WriteRune(char)

			for char, _, err := bufReader.ReadRune(); ; char, _, err = bufReader.ReadRune() {
				if err != nil {
					return tokens, err
				}

				if !unicode.IsNumber(char) {
					bufReader.UnreadRune()
					break
				}

				num.WriteRune(char)
			}

			data, err := strconv.ParseUint(num.String(), 10, 8)
			if err != nil {
				return tokens, err
			}

			tokens = append(tokens, Token{
				Type:   Integer,
				line:   line,
				lexeme: char,
				data:   uint8(data),
			})

			break
		case ' ', '\t', '\r':

			break
		case '\n':
			line += 1

			break
		default:
			if unicode.IsLetter(char) {
				ident := strings.Builder{}
				ident.WriteRune(char)

				for char, _, err := bufReader.ReadRune(); ; char, _, err = bufReader.ReadRune() {
					if err != nil {
						return tokens, err
					}

					if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
						bufReader.UnreadRune()
						break
					}

					ident.WriteRune(char)
				}

				tokens = append(tokens, Token{
					Type:   Identifier,
					line:   line,
					lexeme: char,
					data:   ident.String(),
				})

				break
			}

			return tokens, lexemeErrorInfo(line, char, "Unexpected character")
		}
	}

	return tokens, nil
}

// Helper function to conditionally eat a lexeme if it matches
// the expected rune
func eatIf(reader *bufio.Reader, expected rune) (bool, error) {
	if actual, _, err := reader.ReadRune(); err != nil {
		// TODO(Matt): Would it be more correct to unread the rune here?
		return false, err
	} else if actual == expected {
		return true, nil
	}

	return false, reader.UnreadRune()
}

// Helper function to discard a specified number of runes
func eatN(reader *bufio.Reader, n int) error {
	for i := 0; i < n; i++ {
		if _, _, err := reader.ReadRune(); err != nil {
			return err
		}
	}

	return nil
}
