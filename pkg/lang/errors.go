package lang

import "fmt"

type ErrorInfo struct {
	line     uint
	location string
	message  string
}

func simpleErrorInfo(line uint, message string) ErrorInfo {
	return ErrorInfo{
		line:     line,
		location: "",
		message:  message,
	}
}

func lexemeErrorInfo(line uint, lexeme rune, message string) ErrorInfo {
	return ErrorInfo{
		line:     line,
		location: fmt.Sprintf(" at '%c'", lexeme),
		message:  message,
	}
}

func tokenErrorInfo(token Token, message string) ErrorInfo {
	return ErrorInfo{
		line:     token.line,
		location: fmt.Sprintf(" at '%c'", token.lexeme),
		message:  message,
	}
}

func (err ErrorInfo) Error() string {
	return fmt.Sprintf("[line=%d] Error%s: %s", err.line, err.location, err.message)
}
