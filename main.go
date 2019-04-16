package main

import (
	"apple-x-co/go-pdf/types"
	"encoding/json"
	"fmt"
	"github.com/signintech/gopdf"
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

	var page = types.PDF{LineHeight: 20, TextColor: types.Color{R: 0, G: 0, B: 0}}
	bytes := []byte(string(b))
	if err := json.Unmarshal(bytes, &page); err != nil {
		fmt.Println("error:", err)
		return
	}
	//fmt.Printf("%v\n", page)

	gp := gopdf.GoPdf{}
	gp.Start(gopdf.Config{PageSize: gopdf.Rect{W: page.Width, H: page.Height}, Unit: gopdf.Unit_PT})
	if err := gp.AddTTFFont("default", *ttfPath); err != nil {
		log.Print(err.Error())
		return
	}
	if err := gp.SetFont("default", "", 14); err != nil {
		log.Print(err.Error())
		return
	}

	gp.AddPage()
	gp.SetTextColor(page.TextColor.R, page.TextColor.G, page.TextColor.B)
	drawPdf(&gp, page, page.LinerLayout)

	if err := gp.WritePdf(*outputPath); err != nil {
		log.Print(err.Error())
		return
	}

	_ = gp.Close()
}

func drawPdf(gp *gopdf.GoPdf, page types.PDF, linerLayout types.LinerLayout) {

	//fmt.Printf("orientation: %v\n", linerLayout.Orientation)

	width := page.Width - gp.MarginLeft() - gp.MarginRight()
	//height := page.Height - gp.MarginTop() - gp.MarginBottom()

	for _, element := range linerLayout.Elements {
		x := gp.GetX()
		//y := gp.GetY()

		switch element.Type {
		case "line_break":
			var decoded types.ElementLineBreak
			_ = json.Unmarshal(element.Attributes, &decoded)
			gp.Br(decoded.Height)

		case "text":
			var decoded = types.ElementText{Color: types.Color{R: page.TextColor.R, G: page.TextColor.G, B: page.TextColor.B}}
			_ = json.Unmarshal(element.Attributes, &decoded)

			measureWidth, _ := gp.MeasureTextWidth(decoded.Text)
			measureRect := gopdf.Rect{W: measureWidth, H: 0}

			gp.SetTextColor(decoded.Color.R, decoded.Color.G, decoded.Color.B)

			if linerLayout.IsHorizontal() {
				if x+measureWidth > width {
					if lineHeight := linerLayout.LineHeight; lineHeight != 0 {
						gp.Br(lineHeight)
					} else if lineHeight := page.LineHeight; lineHeight != 0 {
						gp.Br(lineHeight)
					} else {
						gp.Br(20)
					}
				}

				_ = gp.Cell(&measureRect, decoded.Text)
			} else if linerLayout.IsVertical() {
				_ = gp.Cell(&measureRect, decoded.Text)
				gp.SetX(gp.MarginLeft())
				gp.SetY(gp.GetY() + linerLayout.LineHeight)
			}

			gp.SetTextColor(page.TextColor.R, page.TextColor.G, page.TextColor.B)

		case "image":
			var decoded types.ElementImage
			_ = json.Unmarshal(element.Attributes, &decoded)

			imageRect := gopdf.Rect{}
			if decoded.Width != 0 && decoded.Height != 0 {
				imageRect.W = decoded.Width
				imageRect.H = decoded.Height
			} else if decoded.Width == 0 && decoded.Height == 0 {
				file, _ := os.Open(decoded.Path)
				img, _, _ := image.DecodeConfig(file)
				imageRect.W = float64(img.Width)
				imageRect.H = float64(img.Height)
			} else if decoded.Width == 0 && decoded.Height != 0 {
				file, _ := os.Open(decoded.Path)
				img, _, _ := image.DecodeConfig(file)
				imageRect.H = decoded.Height
				imageRect.W = float64(img.Width) * (imageRect.H / float64(img.Height))
			} else if decoded.Width != 0 && decoded.Height == 0 {
				file, _ := os.Open(decoded.Path)
				img, _, _ := image.DecodeConfig(file)
				imageRect.W = decoded.Width
				imageRect.H = float64(img.Height) * (imageRect.W / float64(img.Width))
			}

			if gp.GetX()+imageRect.W > page.Width {
				if lineHeight := linerLayout.LineHeight; lineHeight != 0 {
					gp.Br(lineHeight)
				} else if lineHeight := page.LineHeight; lineHeight != 0 {
					gp.Br(lineHeight)
				} else {
					gp.Br(20)
				}
			}

			if decoded.X != 0 || decoded.Y != 0 {
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
		drawPdf(gp, page, linerLayout)
	}
}
