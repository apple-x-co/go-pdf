package main

import (
	"apple-x-co/go-pdf/types"
	"encoding/json"
	"fmt"
	"github.com/signintech/gopdf"
	"github.com/signintech/gopdf/fontmaker/core"
	"image"
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

	var pdf = types.PDF{LineHeight: 20, TextSize: 14, TextColor: types.Color{R: 0, G: 0, B: 0}, AutoPageBreak: true}
	bytes := []byte(string(b))
	if err := json.Unmarshal(bytes, &pdf); err != nil {
		fmt.Println("error:", err)
		return
	}
	//fmt.Printf("%v\n", pdf)

	gp := gopdf.GoPdf{}
	gp.Start(gopdf.Config{PageSize: gopdf.Rect{W: pdf.Width, H: pdf.Height}, Unit: gopdf.Unit_PT})
	if err := gp.AddTTFFont("default", *ttfPath); err != nil {
		log.Print(err.Error())
		return
	}
	if err := gp.SetFont("default", "", pdf.TextSize); err != nil {
		log.Print(err.Error())
		return
	}

	var parser core.TTFParser
	if err := parser.Parse(*ttfPath); err != nil {
		log.Print(err.Error())
		return
	}
	pdf.SetTextCapHeight(float64(float64(parser.CapHeight()) * 1000.00 / float64(parser.UnitsPerEm())))

	gp.SetTextColor(pdf.TextColor.R, pdf.TextColor.G, pdf.TextColor.B)

	for _, page := range pdf.Pages {
		gp.AddPage()
		drawPdf(&gp, pdf, page.LinerLayout)
	}

	if err := gp.WritePdf(*outputPath); err != nil {
		log.Print(err.Error())
		return
	}

	_ = gp.Close()
}

func drawPdf(gp *gopdf.GoPdf, pdf types.PDF, linerLayout types.LinerLayout) {

	//fmt.Printf("orientation: %v\n", linerLayout.Orientation)

	width := pdf.Width - gp.MarginLeft() - gp.MarginRight()
	//height := pdf.Height - gp.MarginTop() - gp.MarginBottom()

	for _, element := range linerLayout.Elements {
		x := gp.GetX()
		//y := gp.GetY()

		switch element.Type {
		case "line_break":
			var decoded types.ElementLineBreak
			_ = json.Unmarshal(element.Attributes, &decoded)
			gp.Br(decoded.Height)

		case "text":
			var decoded = types.ElementText{Color: types.Color{R: pdf.TextColor.R, G: pdf.TextColor.G, B: pdf.TextColor.B}, Width: -1, Height: -1}
			_ = json.Unmarshal(element.Attributes, &decoded)

			var textRect gopdf.Rect

			if decoded.Width != -1 || decoded.Height != -1 {
				textRect = gopdf.Rect{W: decoded.Width, H: decoded.Height}
			} else {
				measureWidth, _ := gp.MeasureTextWidth(decoded.Text)
				measureHeight := pdf.TextCapHeight() * (float64(pdf.TextSize) / 1000.0)
				textRect = gopdf.Rect{W: measureWidth, H: measureHeight}
			}

			gp.SetTextColor(decoded.Color.R, decoded.Color.G, decoded.Color.B)

			if linerLayout.IsHorizontal() {
				if x+textRect.W > width {
					if lineHeight := linerLayout.LineHeight; lineHeight != 0 {
						gp.Br(lineHeight)
					} else if lineHeight := pdf.LineHeight; lineHeight != 0 {
						gp.Br(lineHeight)
					} else {
						gp.Br(20)
					}
				}

				// todo: 実際のページの高さより早く改ページしてしまっている。
				if gp.GetY()+textRect.H > pdf.Height && pdf.AutoPageBreak {
					gp.AddPage()
				}

				_ = gp.Cell(&textRect, decoded.Text)
			} else if linerLayout.IsVertical() {
				if gp.GetY()+textRect.H > pdf.Height && pdf.AutoPageBreak {
					gp.AddPage()
				}

				_ = gp.Cell(&textRect, decoded.Text)
				gp.SetX(gp.MarginLeft())
				gp.SetY(gp.GetY() + linerLayout.LineHeight)
			}

			gp.SetTextColor(pdf.TextColor.R, pdf.TextColor.G, pdf.TextColor.B)

		case "image":
			var decoded = types.ElementImage{X: -1, Y: -1, Width: -1, Height: -1}
			_ = json.Unmarshal(element.Attributes, &decoded)

			imageRect := gopdf.Rect{}
			if decoded.Width != -1 && decoded.Height != -1 {
				imageRect.W = decoded.Width
				imageRect.H = decoded.Height
			} else if decoded.Width == -1 && decoded.Height == -1 {
				file, _ := os.Open(decoded.Path)
				img, _, _ := image.DecodeConfig(file)
				imageRect.W = float64(img.Width)
				imageRect.H = float64(img.Height)
			} else if decoded.Width == -1 && decoded.Height != -1 {
				file, _ := os.Open(decoded.Path)
				img, _, _ := image.DecodeConfig(file)
				imageRect.H = decoded.Height
				imageRect.W = float64(img.Width) * (imageRect.H / float64(img.Height))
			} else if decoded.Width != -1 && decoded.Height == -1 {
				file, _ := os.Open(decoded.Path)
				img, _, _ := image.DecodeConfig(file)
				imageRect.W = decoded.Width
				imageRect.H = float64(img.Height) * (imageRect.W / float64(img.Width))
			}

			if gp.GetX()+imageRect.W > pdf.Width {
				if lineHeight := linerLayout.LineHeight; lineHeight != 0 {
					gp.Br(lineHeight)
				} else if lineHeight := pdf.LineHeight; lineHeight != 0 {
					gp.Br(lineHeight)
				} else {
					gp.Br(20)
				}
			}

			if gp.GetY()+imageRect.H > pdf.Height && pdf.AutoPageBreak {
				gp.AddPage()
			}

			if decoded.X != -1 || decoded.Y != -1 {
				_ = gp.Image(decoded.Path, decoded.X, decoded.Y, &imageRect)
			} else {
				_ = gp.Image(decoded.Path, gp.GetX(), gp.GetY(), &imageRect)

				if linerLayout.IsHorizontal() {
					gp.SetX(gp.GetX() + imageRect.W)
				} else if linerLayout.IsVertical() {
					gp.SetY(gp.GetY() + imageRect.H)
				}
			}
		}
	}

	for _, linerLayout := range linerLayout.LinearLayouts {
		drawPdf(gp, pdf, linerLayout)
	}
}
