package lang

import (
	"image/color"
	"strings"
	"testing"

	"github.com/mbStavola/slydes/pkg/types"
)

var sly = NewSly()

func TestSimplePresentation(t *testing.T) {
	source := `
	# This is a very simple slideshow
	# Hopefully everything works as intended
	let coolGray = (26, 83, 92);
	let paleGreen = (247, 255, 247,);
	let tealBlue = (78, 205, 196, 255);

	slide intro {
		block ex { ---Welcome!--- }
	}

	slide first {
		self.backgroundColor = coolGray;
		block title {
			---Default
Text
---
		}

		block scope {
			self.font = "Fira Code";
			self.fontColor = paleGreen;
			---Amazing---
		}

		block secondScope : scope {
			self.fontColor = "red";
			self.justify = "center";
			---GoodBye---
		}
	}`

	show, err := sly.ReadSlideShowString(source)
	if err != nil {
		t.Error(err)
		return
	}

	if len(show.Slides) != 2 {
		t.Errorf("Expected exactly two slides-- got %d", len(show.Slides))
		return
	}

	titleSlide := show.Slides[0]
	if len(titleSlide.Blocks) != 1 {
		t.Errorf("Expected exactly one block-- got %d", len(titleSlide.Blocks))
		return
	} else if titleSlide.Blocks[0].Words != "Welcome!" {
		t.Errorf("Expected words \"Welcome!\"-- got %s", strings.TrimSpace(titleSlide.Blocks[0].Words))
		return
	}

	firstSlide := show.Slides[1]
	background := firstSlide.Background.(color.RGBA)
	if background.R != 26 || background.G != 83 || background.B != 92 || background.A != 255 {
		t.Errorf("Expected (26, 83, 92, 255) background-- got (%d, %d, %d, %d)", background.R, background.G, background.B, background.A)
		return
	} else if len(firstSlide.Blocks) != 3 {
		t.Errorf("Expected exactly three blocks-- got %d", len(firstSlide.Blocks))
		return
	}

	firstBlock := firstSlide.Blocks[0]
	if strings.TrimSpace(firstBlock.Words) != "Default\nText" {
		t.Errorf("Expected words \"Default\nText\"-- got %s", strings.TrimSpace(firstBlock.Words))
		return
	} else if firstBlock.Style.Color != color.Black {
		r, g, b, a := firstBlock.Style.Color.RGBA()
		t.Errorf("Expected black font color-- got (%d, %d, %d, %d)", r, g, b, a)
		return
	} else if firstBlock.Style.Font != "Times New Roman" {
		t.Errorf("Expected Times New Roman font-- got %s", firstBlock.Style.Font)
		return
	} else if firstBlock.Style.Justification != types.Left {
		t.Errorf("Expected left justification-- got %s", firstBlock.Style.Justification)
		return
	}

	secondBlock := firstSlide.Blocks[1]
	fontColor := secondBlock.Style.Color.(color.RGBA)
	if strings.TrimSpace(secondBlock.Words) != "Amazing" {
		t.Errorf("Expected words \"Amazing\"-- got %s", strings.TrimSpace(secondBlock.Words))
		return
	} else if fontColor.R != 247 || fontColor.G != 255 || fontColor.B != 247 || fontColor.A != 255 {
		t.Errorf("Expected (247, 255, 247, 255) font color-- got (%d, %d, %d, %d)", fontColor.R, fontColor.G, fontColor.B, fontColor.A)
		return
	} else if secondBlock.Style.Font != "Fira Code" {
		t.Errorf("Expected Fira Code font-- got %s", secondBlock.Style.Font)
		return
	} else if secondBlock.Style.Justification != types.Left {
		t.Errorf("Expected left justification-- got %s", secondBlock.Style.Justification)
		return
	}

	thirdBlock := firstSlide.Blocks[2]
	fontColor = thirdBlock.Style.Color.(color.RGBA)
	if strings.TrimSpace(thirdBlock.Words) != "GoodBye" {
		t.Errorf("Expected words \"GoodBye\"-- got %s", strings.TrimSpace(thirdBlock.Words))
	} else if fontColor.R != 255 || fontColor.G != 0 || fontColor.B != 0 || fontColor.A != 255 {
		t.Errorf("Expected (255, 0, 0, 255) font color-- got (%d, %d, %d, %d)", fontColor.R, fontColor.G, fontColor.B, fontColor.A)
		return
	} else if thirdBlock.Style.Font != "Fira Code" {
		t.Errorf("Expected Fira Code font-- got %s", thirdBlock.Style.Font)
		return
	} else if thirdBlock.Style.Justification != types.Center {
		t.Errorf("Expected center justification-- got %s", thirdBlock.Style.Justification)
		return
	}
}

func TestMacro(t *testing.T) {
	source := `
	let red = "red";
	macro slMacro() {
		self.backgroundColor = red;
	}

	macro blMacro() {
		self.fontColor = "blue";
	}

	slide first {
		$slMacro();

		block ex {
			$blMacro();
			---Example---
		}
	}`

	show, err := sly.ReadSlideShowString(source)
	if err != nil {
		t.Error(err)
		return
	}

	if len(show.Slides) != 1 {
		t.Errorf("Expected exactly one slides-- got %d", len(show.Slides))
		return
	}

	slide := show.Slides[0]
	block := slide.Blocks[0]
	bgColor := slide.Background.(color.RGBA)
	fontColor := block.Style.Color.(color.RGBA)
	if bgColor.R != 255 && bgColor.G != 0 && bgColor.B != 0 && bgColor.A != 255 {
		t.Errorf("Expected red background color-- got (%d, %d, %d, %d)", bgColor.R, bgColor.G, bgColor.B, bgColor.A)
		return
	} else if fontColor.R != 0 && fontColor.G != 0 && fontColor.B != 255 && fontColor.A != 255 {
		t.Errorf("Expected blue font color-- got (%d, %d, %d, %d)", fontColor.R, fontColor.G, fontColor.B, fontColor.A)
	}
}
