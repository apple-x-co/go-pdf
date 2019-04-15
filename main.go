package main

import (
	"apple-x-co/go-pdf/types"
	"encoding/json"
	"fmt"
	"github.com/signintech/gopdf"
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
		outputPath  = flag.StringP("out", "o", "output.pdf", "file path of output pdf.")
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

	f, err := os.Open(*inputPath)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	var decoded types.LinerLayout
	bytes := []byte(string(b))
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		fmt.Println("error:", err)
		return
	}
	//fmt.Printf("%v\n", decoded)

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	if err := pdf.AddTTFFont("default", *ttfPath); err != nil {
		log.Print(err.Error())
		return
	}
	if err := pdf.SetFont("default", "", 14); err != nil {
		log.Print(err.Error())
		return
	}

	drawPdf(&pdf, decoded)

	if err := pdf.WritePdf(*outputPath); err != nil {
		log.Print(err.Error())
		return
	}

	_ = pdf.Close()
}

func drawPdf(pdf *gopdf.GoPdf, linerLayout types.LinerLayout) {

	fmt.Printf("orientation: %v\n", linerLayout.Orientation)

	pdf.AddPage() // todo: 出力ごとに今の高さを調べて、必要なときに自動改行を行うようにする

	for _, element := range linerLayout.Elements {
		switch element.Type {
		case "text":
			var decoded types.ElementText
			_ = json.Unmarshal(element.Attributes, &decoded)
			_ = pdf.Cell(nil, decoded.Text)
		case "image":
			var decoded types.ElementImage
			_ = json.Unmarshal(element.Attributes, &decoded)
			_ = pdf.Cell(nil, decoded.Path)
		}
	}

	for _, linerLayout := range linerLayout.LinearLayouts {
		drawPdf(pdf, linerLayout)
	}
}
