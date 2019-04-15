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
	fmt.Printf("%v\n", decoded)

	aaa(decoded)

	//for _, linerLayout := range decoded.LinearLayouts {
	//	fmt.Printf("orientation: %v\n", linerLayout.Orientation)
	//
	//	for _, element := range linerLayout.Elements {
	//		fmt.Printf("type: %v\n", element.Type)
	//	}
	//}

	//--------------------------------------------------------------------

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()
	if err := pdf.AddTTFFont("default", *ttfPath); err != nil {
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

func aaa(linerLayout types.LinerLayout) {

	fmt.Printf("orientation: %v\n", linerLayout.Orientation)

	for _, element := range linerLayout.Elements {
		fmt.Printf("%v\n", element)

		//fmt.Printf("element: %v\n", element)

		//fmt.Printf("type: %v\n", element.Type)
	}

	for _, linerLayout := range linerLayout.LinearLayouts {
		aaa(linerLayout)
	}
}
