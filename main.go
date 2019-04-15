package main

import (
	"fmt"
	"github.com/signintech/gopdf"
	"log"

	flag "github.com/spf13/pflag"
)

var version string
var revision string

func main() {
	var (
		inputPath   = flag.StringP("in", "i", "go-pdf.json", "file path of input json.")
		outputPath  = flag.StringP("out", "o", "go-pdf.pdf", "file path of output pdf.")
		ttfPath     = flag.StringP("ttf", "t", "fonts/TakaoPGothic.ttf", "file path of ttf.")
		showHelp    = flag.BoolP("help", "h", false, "show help message")
		showVersion = flag.BoolP("version", "v", false, "show version")
	)
	flag.Parse()

	if *showHelp {
		flag.PrintDefaults()
		return
	}
	if *showVersion {
		fmt.Println("version:", version+"."+revision)
		return
	}

	fmt.Println(*inputPath)

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()
	err := pdf.AddTTFFont("default", *ttfPath)
	if err != nil {
		log.Print(err.Error())
		return
	}
	err = pdf.SetFont("default", "", 14)
	if err != nil {
		log.Print(err.Error())
		return
	}
	_ = pdf.Cell(nil, "あいうえお")
	_ = pdf.WritePdf(*outputPath)
}
