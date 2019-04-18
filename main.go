package main

import (
	"apple-x-co/go-pdf/drawer"
	"apple-x-co/go-pdf/types"
	"encoding/json"
	"fmt"
	"github.com/signintech/gopdf"
	"github.com/signintech/gopdf/fontmaker/core"
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

	var pdf = types.PDF{
		LineHeight:    20,
		TextSize:      14,
		TextColor:     types.Color{R: 0, G: 0, B: 0},
		AutoPageBreak: true,
		CompressLevel: 0}
	bytes := []byte(string(b))
	if err := json.Unmarshal(bytes, &pdf); err != nil {
		fmt.Println("error:", err)
		return
	}
	//fmt.Printf("%v\n", pdf)

	gp := gopdf.GoPdf{}

	if pdf.Password == "" {
		gp.Start(gopdf.Config{PageSize: gopdf.Rect{W: pdf.Width, H: pdf.Height}, Unit: gopdf.Unit_PT})
	} else {
		gp.Start(
			gopdf.Config{
				PageSize: gopdf.Rect{W: pdf.Width, H: pdf.Height},
				Unit:     gopdf.Unit_PT,
				Protection: gopdf.PDFProtectionConfig{
					UseProtection: true,
					Permissions:   gopdf.PermissionsPrint | gopdf.PermissionsCopy | gopdf.PermissionsModify,
					OwnerPass:     []byte(pdf.Password),
					UserPass:      []byte(pdf.Password),
				},
			})
	}

	gp.SetCompressLevel(pdf.CompressLevel)

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
		drawer.Draw(&gp, pdf, page.LinerLayout)
	}

	if err := gp.WritePdf(*outputPath); err != nil {
		log.Print(err.Error())
		return
	}

	_ = gp.Close()
}
