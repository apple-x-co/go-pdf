package pdf

import (
	"apple-x-co/go-pdf/types"
	"fmt"

	//"bytes"
	"encoding/json"
	//"github.com/nfnt/resize"
	"github.com/signintech/gopdf"
	"github.com/signintech/gopdf/fontmaker/core"
	//"image"
	//"image/jpeg"
	//"image/png"
	"log"
	//"os"
)

const unsetWidth float64 = 0
const unsetHeight float64 = 0
const unsetX float64 = 0
const unsetY float64 = 0

type PDF struct {
	gp gopdf.GoPdf
}

func (p *PDF) Draw(documentConfigure types.DocumentConfigure) {
	p.gp = gopdf.GoPdf{}

	if documentConfigure.Password == "" {
		p.gp.Start(
			gopdf.Config{
				PageSize: gopdf.Rect{W: documentConfigure.Width, H: documentConfigure.Height},
				Unit:     gopdf.Unit_PT,
			})
	} else {
		p.gp.Start(
			gopdf.Config{
				PageSize: gopdf.Rect{W: documentConfigure.Width, H: documentConfigure.Height},
				Unit:     gopdf.Unit_PT,
				Protection: gopdf.PDFProtectionConfig{
					UseProtection: true,
					Permissions:   gopdf.PermissionsPrint | gopdf.PermissionsCopy | gopdf.PermissionsModify,
					OwnerPass:     []byte(documentConfigure.Password),
					UserPass:      []byte(documentConfigure.Password),
				},
			})
	}

	p.gp.SetCompressLevel(documentConfigure.CompressLevel)

	if err := p.gp.AddTTFFont("default", documentConfigure.TTFPath); err != nil {
		log.Print(err.Error())
		return
	}
	if err := p.gp.SetFont("default", "", documentConfigure.TextSize); err != nil {
		log.Print(err.Error())
		return
	}

	var parser core.TTFParser
	if err := parser.Parse(documentConfigure.TTFPath); err != nil {
		log.Print(err.Error())
		return
	}
	documentConfigure.SetFontHeight(float64(float64(parser.Ascender()+parser.XHeight()+parser.Descender()) * 1000.00 / float64(parser.UnitsPerEm())))

	p.gp.SetTextColor(documentConfigure.TextColor.R, documentConfigure.TextColor.G, documentConfigure.TextColor.B)

	for _, page := range documentConfigure.Pages {
		p.gp.AddPage()
		wrapRect := p.draw(documentConfigure, page.LinerLayout)

		// > debug
		p.gp.SetStrokeColor(255, 0, 0)
		p.gp.RectFromUpperLeft(wrapRect.Origin.X, wrapRect.Origin.Y, wrapRect.Size.Width, wrapRect.Size.Height)
		// < debug
	}
}

func (p *PDF) draw(documentConfigure types.DocumentConfigure, linerLayout types.LinerLayout) types.Rect {
	var wrapRect = types.Rect{Origin: types.Origin{X: p.gp.GetX(), Y: p.gp.GetY()}}
	var lineWrapRect = types.Rect{Origin: types.Origin{X: p.gp.GetX(), Y: p.gp.GetY()}}

	if len(linerLayout.Elements) > 0 {

		for _, element := range linerLayout.Elements {
			if element.Type.IsLineBreak() {
				var decoded = types.ElementLineBreak{
					Height: unsetHeight,
				}
				_ = json.Unmarshal(element.Attributes, &decoded)
				p.lineBreak(&wrapRect, decoded.Height)

			} else if element.Type.IsText() {
				var decoded = types.ElementText{
					Color:           types.Color{R: documentConfigure.TextColor.R, G: documentConfigure.TextColor.G, B: documentConfigure.TextColor.B},
					BackgroundColor: types.Color{R: 0, G: 0, B: 0},
					Width:           unsetWidth,
					Height:          unsetHeight,
					Border:          types.Border{Width: unsetWidth, Color: types.Color{R: 0, B: 0, G: 0}},
					BorderTop:       types.Border{Width: unsetWidth, Color: types.Color{R: 0, B: 0, G: 0}},
					BorderRight:     types.Border{Width: unsetWidth, Color: types.Color{R: 0, B: 0, G: 0}},
					BorderBottom:    types.Border{Width: unsetWidth, Color: types.Color{R: 0, B: 0, G: 0}},
					BorderLeft:      types.Border{Width: unsetWidth, Color: types.Color{R: 0, B: 0, G: 0}},
				}
				_ = json.Unmarshal(element.Attributes, &decoded)

				fmt.Printf("---------------------------\n%v\n", decoded.Text)

				measureSize := p.measureText(documentConfigure, decoded)

				// VERTICAL
				if linerLayout.Orientation.IsVertical() {
					p.lineBreak(&lineWrapRect, linerLayout.LineHeight)
				}

				// LINE BREAK
				if p.needLineBreak(documentConfigure, lineWrapRect, measureSize) {
					fmt.Print("> line break\n")
					p.lineBreak(&lineWrapRect, linerLayout.LineHeight)
				}

				// PAGE BREAK
				if p.needPageBreak(documentConfigure, lineWrapRect, measureSize) {
					fmt.Print("> page break\n")
					p.gp.AddPage()
					p.pageBreak(&lineWrapRect, &wrapRect)
				}

				// DRAWABLE RECT
				textRect := types.Rect{Origin: types.Origin{X: lineWrapRect.MaxX(), Y: lineWrapRect.MinY()}, Size: measureSize}
				p.gp.SetX(textRect.MinX())
				p.gp.SetY(textRect.MinY())

				// DRAW
				fmt.Printf("textRect: %v\n", textRect)
				fmt.Printf("lineWrapRect: %v\n", lineWrapRect)
				p.drawText(documentConfigure, decoded, textRect)

				lineWrapRect = lineWrapRect.Merge(textRect)

				//} else if element.Type.IsImage() {
				//	var decoded = types.ElementImage{
				//		X:      unsetX,
				//		Y:      unsetY,
				//		Width:  unset_width,
				//		Height: unsetHeight,
				//		Resize: false,
				//	}
				//	_ = json.Unmarshal(element.Attributes, &decoded)
				//	rect := p.drawImage(documentConfigure, linerLayout, decoded, wrapRect)
				//	wrapRect = wrapRect.Merge(rect)
			}
		}

		return wrapRect
	}

	for _, _linerLayout := range linerLayout.LinearLayouts {
		rect := p.draw(documentConfigure, _linerLayout)
		wrapRect = wrapRect.Merge(rect)

		// > debug
		p.gp.SetStrokeColor(255, 255, 0)
		p.gp.RectFromUpperLeft(wrapRect.Origin.X, wrapRect.Origin.Y, wrapRect.Size.Width, wrapRect.Size.Height)
		// < debug
	}

	return wrapRect
}

func (p *PDF) measureText(documentConfigure types.DocumentConfigure, decoded types.ElementText) types.Size {
	measureWidth, _ := p.gp.MeasureTextWidth(decoded.Text)
	measureHeight := documentConfigure.FontHeight() * (float64(documentConfigure.TextSize) / 1000.0)

	var measureSize types.Size
	if decoded.Width != unsetWidth && decoded.Height != unsetHeight {
		measureSize = types.Size{Width: decoded.Width, Height: decoded.Height}
	} else if decoded.Width != unsetWidth && decoded.Height == unsetHeight {
		measureSize = types.Size{Width: decoded.Width, Height: measureHeight}
	} else if decoded.Width == unsetWidth && decoded.Height != unsetHeight {
		measureSize = types.Size{Width: measureWidth, Height: decoded.Height}
	} else {
		measureSize = types.Size{Width: measureWidth, Height: measureHeight}
	}

	return measureSize
}

func (p *PDF) drawText(documentConfigure types.DocumentConfigure, decoded types.ElementText, textRect types.Rect) {
	// BORDER, FILL
	if decoded.Border.Width != unsetWidth {
		p.gp.SetLineWidth(decoded.Border.Width)
		p.gp.SetStrokeColor(decoded.Border.Color.R, decoded.Border.Color.G, decoded.Border.Color.B)
		if decoded.BackgroundColor.R != 0 || decoded.BackgroundColor.G != 0 || decoded.BackgroundColor.B != 0 {
			p.gp.SetFillColor(decoded.BackgroundColor.R, decoded.BackgroundColor.G, decoded.BackgroundColor.B)
			p.gp.RectFromUpperLeftWithStyle(textRect.MinX(), textRect.MinY(), textRect.Width(), textRect.Height(), "FD")
		} else {
			p.gp.RectFromUpperLeft(textRect.MinX(), textRect.MinY(), textRect.Width(), textRect.Height())
		}
	} else if decoded.BorderTop.Width != unsetWidth {
		p.gp.SetLineWidth(decoded.BorderTop.Width)
		p.gp.SetStrokeColor(decoded.BorderTop.Color.R, decoded.BorderTop.Color.G, decoded.BorderTop.Color.B)
		p.gp.Line(textRect.MinX(), textRect.MinY(), textRect.MinX()+textRect.Width(), textRect.MinY())
	} else if decoded.BorderRight.Width != unsetWidth {
		p.gp.SetLineWidth(decoded.BorderRight.Width)
		p.gp.SetStrokeColor(decoded.BorderRight.Color.R, decoded.BorderRight.Color.G, decoded.BorderRight.Color.B)
		p.gp.Line(textRect.MinX()+textRect.Width(), textRect.MinY(), textRect.MinX()+textRect.Width(), textRect.MinY()+textRect.Height())
	} else if decoded.BorderBottom.Width != unsetWidth {
		p.gp.SetLineWidth(decoded.BorderBottom.Width)
		p.gp.SetStrokeColor(decoded.BorderBottom.Color.R, decoded.BorderBottom.Color.G, decoded.BorderBottom.Color.B)
		p.gp.Line(textRect.MinX()+textRect.Width(), textRect.MinY()+textRect.Height(), textRect.MinX(), textRect.MinY()+textRect.Height())
	} else if decoded.BorderLeft.Width != unsetWidth {
		p.gp.SetLineWidth(decoded.BorderLeft.Width)
		p.gp.SetStrokeColor(decoded.BorderLeft.Color.R, decoded.BorderLeft.Color.G, decoded.BorderLeft.Color.B)
		p.gp.Line(textRect.MinX(), textRect.MinY()+textRect.Height(), textRect.MinX(), textRect.MinY())
	}

	// FIXME: ALIGN & VALIGN
	//// ALIGN & VALIGN
	//if decoded.Align.IsCenter() {
	//	p.gp.SetX(p.gp.GetX() + ((gpRect.W / 2) - (measureWidth / 2)))
	//} else if decoded.Align.IsRight() {
	//	p.gp.SetX(p.gp.GetX() + gpRect.W - measureWidth)
	//}
	//if decoded.Valign.IsMiddle() {
	//	p.gp.SetY(p.gp.GetY() + ((gpRect.H / 2) - (measureHeight / 2)))
	//} else if decoded.Valign.IsBottom() {
	//	p.gp.SetY(p.gp.GetY() + gpRect.H - measureHeight)
	//}

	// DRAW TEXT
	var gpRect = gopdf.Rect{W: textRect.Width(), H: textRect.Height()}
	p.gp.SetTextColor(decoded.Color.R, decoded.Color.G, decoded.Color.B)
	_ = p.gp.Cell(&gpRect, decoded.Text+fmt.Sprintf("%v", textRect))
	p.gp.SetTextColor(documentConfigure.TextColor.R, documentConfigure.TextColor.G, documentConfigure.TextColor.B)

	// FIXME: RESET ALIGN & VALIGN
	//// RESET ALIGN & VALIGN
	//if decoded.Align.IsCenter() {
	//	p.gp.SetX(p.gp.GetX() - ((gpRect.W / 2) - (measureWidth / 2)))
	//}
	//if decoded.Valign.IsMiddle() {
	//	p.gp.SetY(p.gp.GetY() - ((gpRect.H / 2) - (measureHeight / 2)))
	//} else if decoded.Valign.IsMiddle() {
	//	p.gp.SetY(p.gp.GetY() - gpRect.H + measureHeight)
	//}
}

//func (p *PDF) drawImage(documentConfigure types.DocumentConfigure, linerLayout types.LinerLayout, decoded types.ElementImage, elementsRect types.Rect) types.Rect {
//	height := documentConfigure.Height - p.gp.MarginTop() - p.gp.MarginBottom()
//
//	file, _ := os.Open(decoded.Path)
//	imgConfig, _, _ := image.DecodeConfig(file)
//
//	_, _ = file.Seek(0, 0)
//	img, imgType, _ := image.Decode(file)
//	_ = file.Close()
//
//	var imageRectSize gopdf.Rect
//	if decoded.Width != unset_width && decoded.Height != unsetHeight {
//		imageRectSize.W = decoded.Width
//		imageRectSize.H = decoded.Height
//	} else if decoded.Width == unset_width && decoded.Height == unsetHeight {
//		imageRectSize.W = float64(imgConfig.Width)
//		imageRectSize.H = float64(imgConfig.Height)
//	} else if decoded.Width == unset_width && decoded.Height != unsetHeight {
//		imageRectSize.H = decoded.Height
//		imageRectSize.W = float64(imgConfig.Width) * (imageRectSize.H / float64(imgConfig.Height))
//	} else if decoded.Width != unset_width && decoded.Height == unsetHeight {
//		imageRectSize.W = decoded.Width
//		imageRectSize.H = float64(imgConfig.Height) * (imageRectSize.W / float64(imgConfig.Width))
//	}
//
//	// VERTICAL
//	if linerLayout.Orientation.IsVertical() && p.CurrentSize().Height != 0 {
//		p.gp.SetY(p.gp.GetY() + p.CurrentSize().Height)
//		p.clearCurrentSize()
//	}
//
//	// LINE BREAK
//	if p.gp.GetX()+imageRectSize.W > documentConfigure.Width {
//		if lineHeight := linerLayout.LineHeight; lineHeight != 0 {
//			p.gp.SetX(elementsRect.Origin.X)
//			p.gp.SetY(p.gp.GetY() + lineHeight)
//		} else {
//			p.gp.SetX(elementsRect.Origin.X)
//			p.gp.SetY(p.gp.GetY() + p.CurrentSize().Height)
//		}
//		p.clearCurrentSize()
//	}
//
//	// PAGE BREAK
//	if p.gp.GetY()+imageRectSize.H > height && documentConfigure.AutoPageBreak {
//		p.gp.AddPage()
//		p.clearCurrentSize()
//	}
//
//	// STORE MAX AXIS
//	p.setCurrentSize(imageRectSize.W, imageRectSize.H)
//
//	var imageHoloder gopdf.ImageHolder
//
//	// RESIZE
//	if decoded.Resize && ((decoded.Width != unset_width && decoded.Width < float64(imgConfig.Width)) || (decoded.Height != unsetHeight && decoded.Height < float64(imgConfig.Height))) {
//		resizedImg := resize.Resize(uint(imageRectSize.W)*2, uint(imageRectSize.H)*2, img, resize.Lanczos3)
//
//		resizedBuf := new(bytes.Buffer)
//		switch imgType {
//		case "png":
//			if err := png.Encode(resizedBuf, resizedImg); err != nil {
//				panic(err)
//			}
//		case "jpeg":
//			if err := jpeg.Encode(resizedBuf, resizedImg, nil); err != nil {
//				panic(err)
//			}
//		}
//
//		ih, err := gopdf.ImageHolderByBytes(resizedBuf.Bytes())
//		if err != nil {
//			panic(err)
//		}
//
//		imageHoloder = ih
//	} else {
//		ih, err := gopdf.ImageHolderByPath(decoded.Path)
//		if err != nil {
//			panic(err)
//		}
//
//		imageHoloder = ih
//	}
//
//	// DRAW IMAGE
//	if decoded.X != unsetX || decoded.Y != unsetY {
//		_ = p.gp.ImageByHolder(imageHoloder, decoded.X, decoded.Y, &imageRectSize)
//		return types.Rect{} // dummy
//	}
//	if linerLayout.Orientation.IsHorizontal() {
//		_ = p.gp.Image(decoded.Path, p.gp.GetX(), p.gp.GetY(), &imageRectSize)
//		p.gp.SetX(p.gp.GetX() + imageRectSize.W)
//	} else if linerLayout.Orientation.IsVertical() {
//		_ = p.gp.Image(decoded.Path, p.gp.GetX(), p.gp.GetY(), &imageRectSize)
//	}
//
//	return types.Rect{} // dummy
//}

func (p *PDF) verticalBreak(lineWrapRect *types.Rect) {
	lineWrapRect.Origin.X = p.gp.GetX()
	lineWrapRect.Origin.Y = p.gp.GetY()
	lineWrapRect.Size.Width = 0
	lineWrapRect.Size.Height = 0
}

func (p *PDF) lineBreak(lineWrapRect *types.Rect, lineHeight float64) {
	fmt.Printf("lineWrapRect(before): %v\n", lineWrapRect)

	fmt.Printf("y: %v\n", lineWrapRect.Origin.Y)
	fmt.Printf("h: %v\n", lineWrapRect.Size.Height)
	fmt.Printf("maxy: %v\n", lineWrapRect.MaxY())
	fmt.Printf("lineHeight: %v\n", lineHeight)

	lineWrapRect.Origin.X = lineWrapRect.MinX()
	if lineHeight == unsetHeight {
		lineWrapRect.Origin.Y = lineWrapRect.MaxY()
	} else {
		lineWrapRect.Origin.Y = lineWrapRect.MinY() + lineHeight
	}
	lineWrapRect.Size.Width = 0
	lineWrapRect.Size.Height = 0

	fmt.Printf("lineWrapRect(after): %v\n", lineWrapRect)
}

func (p *PDF) pageBreak(lineWrapRect *types.Rect, wrapRect *types.Rect) {
	lineWrapRect.Origin.X = p.gp.MarginLeft()
	lineWrapRect.Origin.Y = p.gp.MarginTop()
	lineWrapRect.Size.Width = 0
	lineWrapRect.Size.Width = 0

	wrapRect.Origin.X = p.gp.MarginLeft()
	wrapRect.Origin.Y = p.gp.MarginTop()
	wrapRect.Size.Width = 0
	wrapRect.Size.Width = 0
}

func (p *PDF) needLineBreak(documentConfigure types.DocumentConfigure, lineWrapRect types.Rect, measureSize types.Size) bool {
	if lineWrapRect.MaxX()+measureSize.Width > documentConfigure.Width-p.gp.MarginLeft()-p.gp.MarginRight() {
		return true
	}
	return false
}

func (p *PDF) needPageBreak(documentConfigure types.DocumentConfigure, lineWrapRect types.Rect, measureSize types.Size) bool {
	if lineWrapRect.MinY()+measureSize.Height > documentConfigure.Height-p.gp.MarginTop()-p.gp.MarginBottom() {
		return true
	}
	return false
}

func (p *PDF) Save(outputPath string) error {
	return p.gp.WritePdf(outputPath)
}

func (p *PDF) Destroy() {
	_ = p.gp.Close()
}
