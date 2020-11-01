package lang

import (
	"fmt"
	"strings"
)

type ErrorInfoBundle struct {
	errors []ErrorInfo
}

func newErrorInfoBundle() ErrorInfoBundle {
	return ErrorInfoBundle{errors: make([]ErrorInfo, 0)}
}

func (b *ErrorInfoBundle) Add(err ErrorInfo) {
	b.errors = append(b.errors, err)
}

func (b ErrorInfoBundle) HasErrors() bool {
	return len(b.errors) > 0
}

func (b ErrorInfoBundle) Error() string {
	builder := strings.Builder{}
	for _, err := range b.errors {
		builder.WriteString(err.Error())
		builder.WriteByte('\n')
	}

	return builder.String()
}

type ErrorInfo struct {
	line     uint
	location string
	stage    stage
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
		stage:    lexing,
		message:  message,
	}
}

func tokenErrorInfo(token Token, stage stage, message string) ErrorInfo {
	return ErrorInfo{
		line:     token.line,
		location: fmt.Sprintf(" at '%c'", token.lexeme),
		stage:    stage,
		message:  message,
	}
}

func (err ErrorInfo) Error() string {
	stage := ""
	if err.stage != unspecified {
		stage = fmt.Sprintf(" %s ", err.stage)
	}

	return fmt.Sprintf(
		"[line=%d]%sError%s: %s",
		err.line,
		stage,
		err.location,
		err.message,
	)
}
