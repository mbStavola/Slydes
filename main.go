package main

import (
	"flag"
	"fmt"
	"github.com/mbStavola/slydes/pkg/lang"
	"github.com/mbStavola/slydes/render/html"
	"github.com/mbStavola/slydes/render/native"
	"os"
	"strings"
)

func main() {
	filename := flag.String("file", "", "slide to open")
	output := flag.String("out", "html", "method of display (html, native)")
	debug := flag.Bool("debug", false, "print debug info")

	flag.Parse()

	if *filename == "" {
		fmt.Print("Filename must be provided")
		return
	} else if !strings.HasSuffix(*filename, ".sly") {
		fmt.Print("Only .sly files are supported")
		return
	} else if *output != "native" && *output != "html" {
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
	if *debug {
		sly = debugSly(sly)
	}

	show, err := sly.ReadSlideShow(file)
	if err != nil {
		fmt.Print(err)
		return
	}

	switch *output {
	case "native":
		if err := native.Render(show); err != nil {
			fmt.Print(err)
		}

		break
	case "html":
		if err := html.Render(show); err != nil {
			fmt.Print(err)
		}

		break
	}
}
