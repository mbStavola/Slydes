package html

import (
	"fmt"
	"github.com/mbStavola/slydes/pkg/types"
	"html/template"
	"image/color"
	"os"
)

func Render(show types.Show) error {
	helpers := template.FuncMap{
		"count": func(slides []types.Slide) int {
			return len(slides) - 1
		},
		"style": func(style types.Style) template.CSS {
			fontColor := fontColorStyle(style.Color)
			styleText := fmt.Sprintf(
				"font-size: %dpx; font-family: %s; text-align: %s; color: %s;",
				style.Size,
				style.Font,
				style.Justification,
				fontColor,
			)
			return template.CSS(styleText)
		},
		"color": fontColorStyle,
	}
	slideshow, err := template.New("slideshow").Funcs(helpers).Parse(source)
	if err != nil {
		return err
	}

	return slideshow.Execute(os.Stdout, show)
}

const source = `
<div class="content">
    {{range $i, $slide := .Slides}}
		<div class="slide hide" id="slide-{{ $i }}" style="background-color: {{ color $slide.Background }};">
			{{range $j, $block := $slide.Blocks}}
				<div class="block" id="slide-{{ $i }}-block-{{ $j }}" style="{{ style $block.Style }}">
					<span>{{ $block.Words }}</span>
				</div>
			{{end}}
        </div>
    {{end}}
</div>

<style>
	body, html {
		margin: 0;
		padding: 0;
		overflow: hidden;
	}

	.slide {
		width: 100%;
		height: 100%;
		padding: 2em;
	}

	.block > * {
		white-space: pre-line;
	}

	.hide {
		display: none;
	}
</style>

<script>
	// Show title slide by default
	var titleSlide = document.getElementById('slide-0');
	titleSlide.className = titleSlide.className.replace(/hide/g, '');

	function hide(i) {
		var slide = document.getElementById('slide-' + i);
		slide.className += ' hide';
	}

	function show(i) {
		var slide = document.getElementById('slide-' + i);
		slide.className = slide.className.replace(/hide/g, '');
	}

	var currentSlide = 0;

	// Handle keypress left
	document.addEventListener("keydown", function(event) {
		if (currentSlide === 0 || event.keyCode !== 37) {
			return;
		}

		hide(currentSlide);
		currentSlide -= 1;
		show(currentSlide);
	});

	// Handle keypress right
	document.addEventListener("keydown", function(event) {
		if (currentSlide === {{ count .Slides }} || event.keyCode !== 39) {
			return;
		}

		hide(currentSlide);
		currentSlide += 1;
		show(currentSlide);
	});
</script>
`

func fontColorStyle(color color.Color) template.CSS {
	r, g, b, a := color.RGBA()
	convert := func(x uint32) uint8 {
		return uint8((float32(x) / 65535.0) * 255.0)
	}

	// Place rgba values in a range of 0 to 255
	r2, g2, b2, a2 := convert(r), convert(g), convert(b), convert(a)
	colorStyle := fmt.Sprintf("rgba(%d, %d, %d, %d)", r2, g2, b2, a2)

	return template.CSS(colorStyle)
}
