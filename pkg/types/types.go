// Common types for the Slydes package
package types

import "image/color"

type Justification int

const (
	Left Justification = iota
	Right
	Center
)

func (j Justification) String() string {
	return []string{
		"Left",
		"Right",
		"Center",
	}[j]
}

type Show struct {
	Slides []Slide
}

func NewShow() Show {
	return Show{
		Slides: make([]Slide, 0, 32),
	}
}

type Slide struct {
	Background color.Color
	Blocks     []Block
}

func NewSlide() Slide {
	return Slide{
		Background: color.White,
		Blocks:     make([]Block, 0),
	}
}

// A Block represents a styled grouping of text
type Block struct {
	Words string
	Style Style
}

func NewBlock() Block {
	return Block{
		Style: NewStyle(),
	}
}

type Style struct {
	Color         color.Color
	Font          string
	Size          uint8
	Justification Justification
}

func NewStyle() Style {
	return Style{
		Color: color.Black,
		Font:  "Times New Roman",
		Size:  12,
	}
}
