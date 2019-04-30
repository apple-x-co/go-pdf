package main

import (
	"apple-x-co/go-pdf/pdf"
	"apple-x-co/go-pdf/types"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	flag "github.com/spf13/pflag"
)

var version string
var revision string

func main() {
	var (
		inputPath   = flag.StringP("in", "i", "layout.json", "file path of input json.")
		outputPath  = flag.StringP("out", "o", "output.configure", "file path of output configure.")
		ttfPath     = flag.StringP("ttf", "t", "fonts/TakaoPGothic.ttf", "file path of ttf.")
		showHelp    = flag.BoolP("help", "h", false, "show help message")
		showVersion = flag.BoolP("version", "v", false, "show version")
	)
	flag.Parse()

	if *showHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}
	if *showVersion {
		fmt.Println("version:", version+"."+revision)
		os.Exit(0)
	}

	execuute(*inputPath, *outputPath, *ttfPath)

	os.Exit(0)
}

func execuute(inputPath string, outputPath string, ttfPath string) {
	f, err := os.Open(inputPath)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	var documentConfigure = types.DocumentConfigure{
		TextSize:      pdf.DefaultTextSize,
		TextColor:     types.Color{R: pdf.DefaultColorR, G: pdf.DefaultColorG, B: pdf.DefaultColorB},
		AutoPageBreak: true,
		CompressLevel: pdf.DefaultCompressLevel,
		TTFPath:       ttfPath,
		CommonHeader:  types.Header{Size: types.Size{Width: pdf.UnsetWidth, Height: pdf.UnsetHeight}},
		CommonFooter:  types.Footer{Size: types.Size{Width: pdf.UnsetWidth, Height: pdf.UnsetHeight}},
	}
	bytes := []byte(string(b))
	if err := json.Unmarshal(bytes, &documentConfigure); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	//fmt.Printf("%v\n", configure)

	document := pdf.PDF{}
	document.Draw(documentConfigure)
	if err := document.Save(outputPath); err != nil {
		document.Destroy()
		log.Print(err.Error())
		os.Exit(1)
	}
	document.Destroy()
}
