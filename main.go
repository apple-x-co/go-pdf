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
		TextSize:      14,
		TextColor:     types.Color{R: 0, G: 0, B: 0},
		AutoPageBreak: true,
		CompressLevel: 0,
		TTFPath:       ttfPath,
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
