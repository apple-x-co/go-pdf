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
	pdf.SetTextHeight(float64(float64(parser.Ascender()+parser.XHeight()+parser.Descender()) * 1000.00 / float64(parser.UnitsPerEm())))

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

	for _, element := range linerLayout.Elements {
		switch element.Type {
		case "line_break":
			var decoded types.ElementLineBreak
			_ = json.Unmarshal(element.Attributes, &decoded)
			gp.Br(decoded.Height)

		case "text":
			var decoded = types.ElementText{Color: types.Color{R: pdf.TextColor.R, G: pdf.TextColor.G, B: pdf.TextColor.B}, BackgroundColor: types.Color{R: 0, G: 0, B: 0}, Width: -1, Height: -1, Border: types.Border{Width: -1, Color: types.Color{R: 0, B: 0, G: 0}}}
			_ = json.Unmarshal(element.Attributes, &decoded)
			drawText(gp, pdf, linerLayout, decoded)

		case "image":
			var decoded = types.ElementImage{X: -1, Y: -1, Width: -1, Height: -1}
			_ = json.Unmarshal(element.Attributes, &decoded)
			drawImage(gp, pdf, linerLayout, decoded)
		}
	}

	for _, linerLayout := range linerLayout.LinearLayouts {
		drawPdf(gp, pdf, linerLayout)
	}
}

func drawText(gp *gopdf.GoPdf, pdf types.PDF, linerLayout types.LinerLayout, decoded types.ElementText) {
	x := gp.GetX()
	width := pdf.Width - gp.MarginLeft() - gp.MarginRight()
	height := pdf.Height - gp.MarginTop() - gp.MarginBottom()

	measureWidth, _ := gp.MeasureTextWidth(decoded.Text)
	measureHeight := pdf.TextHeight() * (float64(pdf.TextSize) / 1000.0)

	var textRect gopdf.Rect

	if decoded.Width != -1 && decoded.Height != -1 {
		textRect = gopdf.Rect{W: decoded.Width, H: decoded.Height}
	} else if decoded.Width != -1 && decoded.Height == -1 {
		textRect = gopdf.Rect{W: decoded.Width, H: measureHeight}
	} else if decoded.Width == -1 && decoded.Height != -1 {
		textRect = gopdf.Rect{W: measureWidth, H: decoded.Height}
	} else {
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

		if gp.GetY()+textRect.H > height && pdf.AutoPageBreak {
			gp.AddPage()
		}

		if decoded.Border.Width != -1 {
			gp.SetLineWidth(decoded.Border.Width)
			gp.SetStrokeColor(decoded.Border.Color.R, decoded.Border.Color.G, decoded.Border.Color.B)
			if decoded.BackgroundColor.R != 0 || decoded.BackgroundColor.G != 0 || decoded.BackgroundColor.B != 0 {
				//gp.Line(gp.GetX(), gp.GetY(), gp.GetX()+textRect.W, gp.GetY())
				//gp.Line(gp.GetX()+textRect.W, gp.GetY(), gp.GetX()+textRect.W, gp.GetY()+textRect.H)
				//gp.Line(gp.GetX()+textRect.W, gp.GetY()+textRect.H, gp.GetX(), gp.GetY()+textRect.H)
				//gp.Line(gp.GetX(), gp.GetY()+textRect.H, gp.GetX(), gp.GetY())
				gp.SetFillColor(decoded.BackgroundColor.R, decoded.BackgroundColor.G, decoded.BackgroundColor.B)
				gp.RectFromUpperLeftWithStyle(gp.GetX(), gp.GetY(), textRect.W, textRect.H, "FD")
			} else {
				gp.RectFromUpperLeft(gp.GetX(), gp.GetY(), textRect.W, textRect.H)
			}
		}

		if decoded.Align == "center" {
			gp.SetX(gp.GetX() + ((textRect.W / 2) - (measureWidth / 2)))
		} else if decoded.Align == "right" {
			gp.SetX(gp.GetX() + textRect.W - measureWidth)
		}

		_ = gp.Cell(&textRect, decoded.Text)

		if decoded.Align == "center" {
			gp.SetX(gp.GetX() - ((textRect.W / 2) - (measureWidth / 2)))
		}
	} else if linerLayout.IsVertical() {
		if gp.GetY()+textRect.H > height && pdf.AutoPageBreak {
			gp.AddPage()
		}

		if decoded.Border.Width != -1 {
			gp.SetLineWidth(decoded.Border.Width)
			gp.SetStrokeColor(decoded.Border.Color.R, decoded.Border.Color.G, decoded.Border.Color.B)
			if decoded.BackgroundColor.R != 0 || decoded.BackgroundColor.G != 0 || decoded.BackgroundColor.B != 0 {
				//gp.Line(gp.GetX(), gp.GetY(), gp.GetX()+textRect.W, gp.GetY())
				//gp.Line(gp.GetX()+textRect.W, gp.GetY(), gp.GetX()+textRect.W, gp.GetY()+textRect.H)
				//gp.Line(gp.GetX()+textRect.W, gp.GetY()+textRect.H, gp.GetX(), gp.GetY()+textRect.H)
				//gp.Line(gp.GetX(), gp.GetY()+textRect.H, gp.GetX(), gp.GetY())
				gp.SetFillColor(decoded.BackgroundColor.R, decoded.BackgroundColor.G, decoded.BackgroundColor.B)
				gp.RectFromUpperLeftWithStyle(gp.GetX(), gp.GetY(), textRect.W, textRect.H, "FD")
			} else {
				gp.RectFromUpperLeft(gp.GetX(), gp.GetY(), textRect.W, textRect.H)
			}
		}

		_ = gp.Cell(&textRect, decoded.Text)
		gp.SetX(gp.MarginLeft())
		gp.SetY(gp.GetY() + linerLayout.LineHeight)
	}

	gp.SetTextColor(pdf.TextColor.R, pdf.TextColor.G, pdf.TextColor.B)
}

func drawImage(gp *gopdf.GoPdf, pdf types.PDF, linerLayout types.LinerLayout, decoded types.ElementImage) {
	height := pdf.Height - gp.MarginTop() - gp.MarginBottom()

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

	if gp.GetY()+imageRect.H > height && pdf.AutoPageBreak {
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
