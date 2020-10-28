package lang

import (
	"image/color"
	"testing"
)

var compiler Compiler = NewDefaultCompiler()

func TestAttributeAssignment(t *testing.T) {
	statements := []Statement{
		{
			Type: SlideDecl,
		},
		{
			Type: AttributeAssignment,
			data: AttributeStatement{
				name:  "backgroundColor",
				value: "black",
			},
		},
	}

	show, err := compiler.Compile(statements)

	if err != nil {
		t.Error(err)
		return
	}

	if len(show.Slides) != 2 {
		t.Errorf("Expected exactly two slides-- got %d", len(show.Slides))
		return
	}

	// Skip title slide
	slide := show.Slides[1]
	if slide.Background != color.Black {
		t.Errorf("Expected black background-- got %s", slide.Background)
	}
}
