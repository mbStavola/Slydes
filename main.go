package main

import (
	"flag"
	"fmt"
	"github.com/mbStavola/slydes/pkg/lang"
	"github.com/mbStavola/slydes/pkg/render/html"
	"github.com/mbStavola/slydes/pkg/render/native"
	"os"
	"strings"
)

func main() {
	filename := flag.String("file", "", "slide to open")
	output := flag.String("out", "native", "method of display (native, html, debug)")

	flag.Parse()

	if *filename == "" {
		fmt.Print("Filename must be provided")
		return
	} else if !strings.HasSuffix(*filename, ".sly") {
		fmt.Print("Only .sly files are supported")
		return
	} else if *output != "native" && *output != "html" && *output != "debug" {
		fmt.Print("Output must be either native or html")
		return
	}

	file, err := os.Open(*filename)
	if err != nil {
		fmt.Print(err)
		return
	}
	defer file.Close()

	sly := lang.NewSly()
	if *output == "debug" {
		sly = debugSly(sly)
	}

	show, err := sly.ReadSlideShow(file)
	if err != nil {
		fmt.Print(err)
		return
	}

	switch *output {
	case "native":
		native.Render(show)

		break
	case "html":
		html.Render(show)

		break
	}
}
