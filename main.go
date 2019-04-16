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

	var page types.Page
	bytes := []byte(string(b))
	if err := json.Unmarshal(bytes, &page); err != nil {
		fmt.Println("error:", err)
		return
	}
	//fmt.Printf("%v\n", page)

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: page.Width, H: page.Height}, Unit: gopdf.Unit_PT})
	if err := pdf.AddTTFFont("default", *ttfPath); err != nil {
		log.Print(err.Error())
		return
	}
	if err := pdf.SetFont("default", "", 14); err != nil {
		log.Print(err.Error())
		return
	}

	left, top, right, bottom := pdf.Margins()
	fmt.Printf("left: %v\n", left)
	fmt.Printf("top: %v\n", top)
	fmt.Printf("right: %v\n", right)
	fmt.Printf("bottom: %v\n", bottom)

	pdf.AddPage()
	drawPdf(&pdf, page, page.LinerLayout)

	if err := pdf.WritePdf(*outputPath); err != nil {
		log.Print(err.Error())
		return
	}

	_ = pdf.Close()
}

func drawPdf(pdf *gopdf.GoPdf, page types.Page, linerLayout types.LinerLayout) {

	fmt.Printf("orientation: %v\n", linerLayout.Orientation)

	width := page.Width - pdf.MarginLeft() - pdf.MarginRight()
	//height := page.Height - pdf.MarginTop() - pdf.MarginBottom()

	for _, element := range linerLayout.Elements {
		x := pdf.GetX()
		//y := pdf.GetY()

		switch element.Type {
		case "line_break":
			var decoded types.ElementLineBreak
			_ = json.Unmarshal(element.Attributes, &decoded)
			pdf.Br(decoded.Height)

		case "text":
			var decoded types.ElementText
			_ = json.Unmarshal(element.Attributes, &decoded)

			measureWidth, _ := pdf.MeasureTextWidth(decoded.Text)
			measureRect := gopdf.Rect{W: measureWidth, H: 0}

			if linerLayout.IsHorizontal() {
				if x+measureWidth > width {
					fmt.Printf("break!!\n")
					if lineHeight := linerLayout.LineHeight; lineHeight != 0 {
						pdf.Br(lineHeight)
					} else if lineHeight := page.LineHeight; lineHeight != 0 {
						pdf.Br(lineHeight)
					} else {
						pdf.Br(20)
					}
				}

				_ = pdf.Cell(&measureRect, decoded.Text)
			} else if linerLayout.IsVertical() {
				_ = pdf.Cell(&measureRect, decoded.Text)
				pdf.SetX(pdf.MarginLeft())
				pdf.SetY(pdf.GetY() + linerLayout.LineHeight)
			}

		case "image":
			var decoded types.ElementImage
			_ = json.Unmarshal(element.Attributes, &decoded)
			//_ = pdf.Cell(nil, decoded.Path)
			_ = pdf.Image(decoded.Path, 200, 50, nil)
		}
	}

	for _, linerLayout := range linerLayout.LinearLayouts {
		drawPdf(pdf, page, linerLayout)
	}
}
