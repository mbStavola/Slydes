package html

import (
	"github.com/mbStavola/slydes/pkg/types"
	"html/template"
	"os"
)

func Render(show types.Show) error {
	helpers := template.FuncMap{
		"count": func(slides []types.Slide) int {
			return len(slides) - 1
		},
	}
	slideshow, err := template.New("slideshow").Funcs(helpers).Parse(`
<h1>Slideshow</h1>
<div class="content">
    {{range $i, $slide := .Slides}}
		<div class="slide hide" id="slide-{{ $i }}">
			{{range $j, $block := $slide.Blocks}}
				<div class="block" id="slide-{{ $i }}-block-{{ $j }}">
					{{ $block.Words }}
				</div>
			{{end}}
        </div>
    {{end}}
</div>

<style>
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
`)
	if err != nil {
		return err
	}

	return slideshow.Execute(os.Stdout, show)
}
