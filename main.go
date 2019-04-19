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
		outputPath  = flag.StringP("out", "o", "output.configure", "file path of output configure.")
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

	var configure = types.DocumentConfigure{
		TextSize:      14,
		TextColor:     types.Color{R: 0, G: 0, B: 0},
		AutoPageBreak: true,
		CompressLevel: 0}
	bytes := []byte(string(b))
	if err := json.Unmarshal(bytes, &configure); err != nil {
		fmt.Println("error:", err)
		return
	}
	//fmt.Printf("%v\n", configure)

	gp := gopdf.GoPdf{}

	if configure.Password == "" {
		gp.Start(gopdf.Config{PageSize: gopdf.Rect{W: configure.Width, H: configure.Height}, Unit: gopdf.Unit_PT})
	} else {
		gp.Start(
			gopdf.Config{
				PageSize: gopdf.Rect{W: configure.Width, H: configure.Height},
				Unit:     gopdf.Unit_PT,
				Protection: gopdf.PDFProtectionConfig{
					UseProtection: true,
					Permissions:   gopdf.PermissionsPrint | gopdf.PermissionsCopy | gopdf.PermissionsModify,
					OwnerPass:     []byte(configure.Password),
					UserPass:      []byte(configure.Password),
				},
			})
	}

	gp.SetCompressLevel(configure.CompressLevel)

	if err := gp.AddTTFFont("default", *ttfPath); err != nil {
		log.Print(err.Error())
		return
	}
	if err := gp.SetFont("default", "", configure.TextSize); err != nil {
		log.Print(err.Error())
		return
	}

	var parser core.TTFParser
	if err := parser.Parse(*ttfPath); err != nil {
		log.Print(err.Error())
		return
	}
	configure.SetTextHeight(float64(float64(parser.Ascender()+parser.XHeight()+parser.Descender()) * 1000.00 / float64(parser.UnitsPerEm())))

	gp.SetTextColor(configure.TextColor.R, configure.TextColor.G, configure.TextColor.B)

	for _, page := range configure.Pages {
		gp.AddPage()
		drawer.Draw(&gp, configure, page.LinerLayout)
	}

	if err := gp.WritePdf(*outputPath); err != nil {
		log.Print(err.Error())
		return
	}

	_ = gp.Close()
}
