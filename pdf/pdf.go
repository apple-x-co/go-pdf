package pdf

import (
	"apple-x-co/go-pdf/types"
	"bytes"
	"encoding/json"
	"github.com/nfnt/resize"
	"github.com/signintech/gopdf"
	"github.com/signintech/gopdf/fontmaker/core"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"strings"
)

const UnsetWidth float64 = 0
const UnsetHeight float64 = 0
const UnsetX float64 = 0
const UnsetY float64 = 0
const DefaultColorR uint8 = 0
const DefaultColorG uint8 = 0
const DefaultColorB uint8 = 0
const DefaultTextSize int = 14
const DefaultCompressLevel int = -1

type PDF struct {
	gp          gopdf.GoPdf
	contentRect types.Rect
	headerRect  types.Rect
	footerRect  types.Rect
}

func (p *PDF) Draw(documentConfigure types.DocumentConfigure) {
	p.gp = gopdf.GoPdf{}

	//fmt.Printf("%v\n", documentConfigure)

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

	// FONT
	var parser core.TTFParser
	if err := parser.Parse(documentConfigure.TTFPath); err != nil {
		log.Print(err.Error())
		return
	}
	documentConfigure.SetFontHeight(float64(float64(parser.Ascender()+parser.XHeight()+parser.Descender()) * 1000.00 / float64(parser.UnitsPerEm())))

	p.gp.SetTextColor(documentConfigure.TextColor.R, documentConfigure.TextColor.G, documentConfigure.TextColor.B)

	// RECT
	if documentConfigure.Header.Size.IsSet() {
		p.headerRect = types.Rect{
			Origin: types.Origin{
				X: p.gp.MarginLeft(),
				Y: p.gp.MarginTop(),
			},
			Size: documentConfigure.Header.Size,
		}
	}
	if documentConfigure.Footer.Size.IsSet() {
		p.footerRect = types.Rect{
			Origin: types.Origin{
				X: p.gp.MarginLeft(),
				Y: documentConfigure.Height - p.gp.MarginBottom() - documentConfigure.Footer.Size.Height,
			},
			Size: documentConfigure.Footer.Size,
		}
	}
	p.contentRect = types.Rect{
		Origin: types.Origin{
			X: p.gp.MarginLeft(),
			Y: p.gp.MarginTop() + p.headerRect.Height(),
		},
		Size: types.Size{
			Width:  documentConfigure.Width - p.gp.MarginLeft() - p.gp.MarginRight(),
			Height: documentConfigure.Height - p.gp.MarginTop() - p.gp.MarginBottom() - p.headerRect.Height() - p.footerRect.Height(),
		},
	}

	// DRAW
	for _, page := range documentConfigure.Pages {
		p.gp.AddPage()

		if p.headerRect.Size.IsSet() {
			p.drawHeader(documentConfigure)
		}
		if p.footerRect.Size.IsSet() {
			p.drawFooter(documentConfigure)
		}

		p.gp.SetX(p.contentRect.MinX())
		p.gp.SetY(p.contentRect.MinY())
		p.draw(documentConfigure, page.LinerLayout)
		//fmt.Printf("rect: %v\n", rect)
	}
}

func (p *PDF) Save(outputPath string) error {
	return p.gp.WritePdf(outputPath)
}

func (p *PDF) Destroy() {
	_ = p.gp.Close()
}

// 描画要素のループ
func (p *PDF) draw(documentConfigure types.DocumentConfigure, linerLayout types.LinerLayout) types.Rect {
	var wrapRect = types.Rect{Origin: types.Origin{X: p.gp.GetX(), Y: p.gp.GetY()}}
	var lineWrapRect = types.Rect{Origin: types.Origin{X: p.gp.GetX(), Y: p.gp.GetY()}}

	if len(linerLayout.Elements) > 0 {

		for _, element := range linerLayout.Elements {
			if element.Type.IsLineBreak() {
				var decoded = types.ElementLineBreak{
					Height: UnsetHeight,
				}
				_ = json.Unmarshal(element.Attributes, &decoded)
				p.breakLine(&lineWrapRect, decoded.Height)

			} else if element.Type.IsText() {
				var decoded = types.ElementText{
					Color:           types.Color{R: documentConfigure.TextColor.R, G: documentConfigure.TextColor.G, B: documentConfigure.TextColor.B},
					BackgroundColor: types.Color{R: DefaultColorR, B: DefaultColorG, G: DefaultColorB},
					Size:            types.Size{Width: UnsetWidth, Height: UnsetWidth},
					Origin:          types.Origin{X: UnsetWidth, Y: UnsetHeight},
					Border:          types.Border{Width: UnsetWidth, Color: types.Color{R: DefaultColorR, B: DefaultColorG, G: DefaultColorB}},
					BorderTop:       types.Border{Width: UnsetWidth, Color: types.Color{R: DefaultColorR, B: DefaultColorG, G: DefaultColorB}},
					BorderRight:     types.Border{Width: UnsetWidth, Color: types.Color{R: DefaultColorR, B: DefaultColorG, G: DefaultColorB}},
					BorderBottom:    types.Border{Width: UnsetWidth, Color: types.Color{R: DefaultColorR, B: DefaultColorG, G: DefaultColorB}},
					BorderLeft:      types.Border{Width: UnsetWidth, Color: types.Color{R: DefaultColorR, B: DefaultColorG, G: DefaultColorB}},
				}
				_ = json.Unmarshal(element.Attributes, &decoded)

				//fmt.Printf("---------------------------\n%v\n", decoded.Text)

				if decoded.Align != "" && decoded.Size.Width == UnsetWidth {
					panic("aligns need width.")
				}
				if decoded.Valign != "" && decoded.Size.Height == UnsetHeight {
					panic("valigns need height.")
				}

				measureSize := p.measureText(documentConfigure, decoded)

				// FIX POSITION
				if decoded.Origin.X != UnsetX && decoded.Origin.Y != UnsetY {
					textFrame := types.Rect{Origin: types.Origin{X: decoded.Origin.X, Y: decoded.Origin.Y}, Size: measureSize}
					textRect := textFrame.Inset(types.EdgeInset{
						Top:    decoded.Margin.Top * -1,
						Right:  decoded.Margin.Right * -1,
						Bottom: decoded.Margin.Bottom * -1,
						Left:   decoded.Margin.Left * -1,
					})
					p.gp.SetX(textRect.MinX())
					p.gp.SetY(textRect.MinY())
					p.drawText(documentConfigure, decoded, textRect, textFrame)
					continue
				}

				// VERTICAL
				if linerLayout.Orientation.IsVertical() {
					p.breakLine(&lineWrapRect, linerLayout.LineHeight)
				}

				// LINE BREAK
				if p.needLineBreak(documentConfigure, lineWrapRect, measureSize) {
					//fmt.Print("> line break\n")
					p.breakLine(&lineWrapRect, linerLayout.LineHeight)
				}

				// PAGE BREAK
				if p.needPageBreak(documentConfigure, lineWrapRect, measureSize) {
					//fmt.Print("> page break\n")
					p.gp.AddPage()
					p.breakPage(&lineWrapRect, &wrapRect)

					if p.headerRect.Size.IsSet() {
						p.drawHeader(documentConfigure)
					}
					if p.footerRect.Size.IsSet() {
						p.drawFooter(documentConfigure)
					}

					p.gp.SetX(wrapRect.MinX())
					p.gp.SetY(wrapRect.MinY())
				}

				// DRAWABLE RECT
				textFrame := types.Rect{Origin: types.Origin{X: lineWrapRect.MaxX(), Y: lineWrapRect.MinY()}, Size: measureSize}
				textRect := textFrame.Inset(decoded.Margin)
				p.gp.SetX(textRect.MinX())
				p.gp.SetY(textRect.MinY())

				// DRAW
				//fmt.Printf("textRect: %v\n", textRect)
				//fmt.Printf("lineWrapRect: %v\n", lineWrapRect)
				p.drawText(documentConfigure, decoded, textRect, textFrame)

				lineWrapRect = lineWrapRect.Merge(textFrame)

			} else if element.Type.IsImage() {
				var decoded = types.ElementImage{
					Size:       types.Size{Width: UnsetWidth, Height: UnsetWidth},
					Origin:     types.Origin{X: UnsetWidth, Y: UnsetHeight},
					Resize:     false,
					Resolution: 2,
				}
				_ = json.Unmarshal(element.Attributes, &decoded)

				//fmt.Printf("---------------------------\n%v\n", decoded.Path)

				measureSize := p.measureImage(documentConfigure, decoded)

				// FIX POSITION
				if decoded.Origin.X != UnsetX && decoded.Origin.Y != UnsetY {
					imageFrame := types.Rect{Origin: types.Origin{X: decoded.Origin.X, Y: decoded.Origin.Y}, Size: measureSize}
					imageRect := imageFrame.Inset(decoded.Margin)
					p.gp.SetX(imageRect.MinX())
					p.gp.SetY(imageRect.MinY())
					p.drawImage(documentConfigure, decoded, imageRect, imageFrame)
					continue
				}

				// VERTICAL
				if linerLayout.Orientation.IsVertical() {
					p.breakLine(&lineWrapRect, linerLayout.LineHeight)
				}

				// LINE BREAK
				if p.needLineBreak(documentConfigure, lineWrapRect, measureSize) {
					//fmt.Print("> line break\n")
					p.breakLine(&lineWrapRect, linerLayout.LineHeight)
				}

				// PAGE BREAK
				if p.needPageBreak(documentConfigure, lineWrapRect, measureSize) {
					//fmt.Print("> page break\n")
					p.gp.AddPage()
					p.breakPage(&lineWrapRect, &wrapRect)

					if p.headerRect.Size.IsSet() {
						p.drawHeader(documentConfigure)
					}
					if p.footerRect.Size.IsSet() {
						p.drawFooter(documentConfigure)
					}

					p.gp.SetX(wrapRect.MinX())
					p.gp.SetY(wrapRect.MinY())
				}

				// DRAWABLE RECT
				imageFrame := types.Rect{Origin: types.Origin{X: lineWrapRect.MaxX(), Y: lineWrapRect.MinY()}, Size: measureSize}
				imageRect := imageFrame.Inset(decoded.Margin)
				p.gp.SetX(imageRect.MinX())
				p.gp.SetY(imageRect.MinY())

				// DRAW
				//fmt.Printf("textRect: %v\n", imageRect)
				//fmt.Printf("lineWrapRect: %v\n", lineWrapRect)
				// > debug
				//p.gp.SetStrokeColor(255, 255, 0)
				//p.gp.RectFromUpperLeft(imageRect.Origin.X, imageRect.Origin.Y, imageRect.Size.Width, imageRect.Size.Height)
				// < debug
				p.drawImage(documentConfigure, decoded, imageRect, imageFrame)

				lineWrapRect = lineWrapRect.Merge(imageFrame)
			}

			wrapRect = wrapRect.Merge(lineWrapRect)
		}

		// > debug
		//p.gp.SetStrokeColor(255, 0, 0)
		//p.gp.RectFromUpperLeft(wrapRect.Origin.X, wrapRect.Origin.Y, wrapRect.Width(), wrapRect.Height())
		// < debug

		return wrapRect
	}

	for _, _linerLayout := range linerLayout.LinerLayouts {
		drawnRect := p.draw(documentConfigure, _linerLayout)
		wrapRect = wrapRect.Merge(drawnRect)

		// > debug
		//p.gp.SetStrokeColor(255, 255, 0)
		//p.gp.RectFromUpperLeft(wrapRect.Origin.X, wrapRect.Origin.Y, wrapRect.Size.Width, wrapRect.Size.Height)
		// < debug

		if linerLayout.Orientation.IsHorizontal() {
			p.gp.SetX(wrapRect.MaxX())
			p.gp.SetY(wrapRect.MinY())
		} else if linerLayout.Orientation.IsVertical() {
			p.gp.SetX(wrapRect.MinX())
			p.gp.SetY(wrapRect.MaxY())
		}
	}

	return wrapRect
}

// 描画要素のループ（ヘッターフッター）
func (p *PDF) drawHeaderOrFooter(documentConfigure types.DocumentConfigure, linerLayout types.LinerLayout) types.Rect {
	var wrapRect = types.Rect{Origin: types.Origin{X: p.gp.GetX(), Y: p.gp.GetY()}}
	var lineWrapRect = types.Rect{Origin: types.Origin{X: p.gp.GetX(), Y: p.gp.GetY()}}

	if len(linerLayout.Elements) > 0 {

		for _, element := range linerLayout.Elements {
			if element.Type.IsLineBreak() {
				var decoded = types.ElementLineBreak{
					Height: UnsetHeight,
				}
				_ = json.Unmarshal(element.Attributes, &decoded)
				p.breakLine(&lineWrapRect, decoded.Height)

			} else if element.Type.IsText() {
				var decoded = types.ElementText{
					Color:           types.Color{R: documentConfigure.TextColor.R, G: documentConfigure.TextColor.G, B: documentConfigure.TextColor.B},
					BackgroundColor: types.Color{R: DefaultColorR, B: DefaultColorG, G: DefaultColorB},
					Size:            types.Size{Width: UnsetWidth, Height: UnsetWidth},
					Origin:          types.Origin{X: UnsetWidth, Y: UnsetHeight},
					Border:          types.Border{Width: UnsetWidth, Color: types.Color{R: DefaultColorR, B: DefaultColorG, G: DefaultColorB}},
					BorderTop:       types.Border{Width: UnsetWidth, Color: types.Color{R: DefaultColorR, B: DefaultColorG, G: DefaultColorB}},
					BorderRight:     types.Border{Width: UnsetWidth, Color: types.Color{R: DefaultColorR, B: DefaultColorG, G: DefaultColorB}},
					BorderBottom:    types.Border{Width: UnsetWidth, Color: types.Color{R: DefaultColorR, B: DefaultColorG, G: DefaultColorB}},
					BorderLeft:      types.Border{Width: UnsetWidth, Color: types.Color{R: DefaultColorR, B: DefaultColorG, G: DefaultColorB}},
				}
				_ = json.Unmarshal(element.Attributes, &decoded)

				if decoded.Align != "" && decoded.Size.Width == UnsetWidth {
					panic("aligns need width.")
				}
				if decoded.Valign != "" && decoded.Size.Height == UnsetHeight {
					panic("valigns need height.")
				}

				measureSize := p.measureText(documentConfigure, decoded)

				// FIX POSITION
				if decoded.Origin.X != UnsetX && decoded.Origin.Y != UnsetY {
					textFrame := types.Rect{Origin: types.Origin{X: decoded.Origin.X, Y: decoded.Origin.Y}, Size: measureSize}
					textRect := textFrame.Inset(decoded.Margin)
					p.gp.SetX(textRect.MinX())
					p.gp.SetY(textRect.MinY())
					p.drawText(documentConfigure, decoded, textRect, textFrame)
					continue
				}

				// VERTICAL
				if linerLayout.Orientation.IsVertical() {
					p.breakLine(&lineWrapRect, linerLayout.LineHeight)
				}

				// LINE BREAK
				if p.needLineBreak(documentConfigure, lineWrapRect, measureSize) {
					//fmt.Print("> line break\n")
					p.breakLine(&lineWrapRect, linerLayout.LineHeight)
				}

				// DRAWABLE RECT
				textFrame := types.Rect{Origin: types.Origin{X: lineWrapRect.MaxX(), Y: lineWrapRect.MinY()}, Size: measureSize}
				textRect := textFrame.Inset(decoded.Margin)
				p.gp.SetX(textRect.MinX())
				p.gp.SetY(textRect.MinY())

				// DRAW
				p.drawText(documentConfigure, decoded, textRect, textFrame)

				lineWrapRect = lineWrapRect.Merge(textFrame)

			} else if element.Type.IsImage() {
				var decoded = types.ElementImage{
					Size:   types.Size{Width: UnsetWidth, Height: UnsetWidth},
					Origin: types.Origin{X: UnsetWidth, Y: UnsetHeight},
					Resize: false,
				}
				_ = json.Unmarshal(element.Attributes, &decoded)

				measureSize := p.measureImage(documentConfigure, decoded)

				// FIX POSITION
				if decoded.Origin.X != UnsetX && decoded.Origin.Y != UnsetY {
					imageFrame := types.Rect{Origin: types.Origin{X: decoded.Origin.X, Y: decoded.Origin.Y}, Size: measureSize}
					imageRect := imageFrame.Inset(decoded.Margin)
					p.gp.SetX(imageRect.MinX())
					p.gp.SetY(imageRect.MinY())
					p.drawImage(documentConfigure, decoded, imageRect, imageFrame)
					continue
				}

				// VERTICAL
				if linerLayout.Orientation.IsVertical() {
					p.breakLine(&lineWrapRect, linerLayout.LineHeight)
				}

				// LINE BREAK
				if p.needLineBreak(documentConfigure, lineWrapRect, measureSize) {
					p.breakLine(&lineWrapRect, linerLayout.LineHeight)
				}

				// DRAWABLE RECT
				imageFrame := types.Rect{Origin: types.Origin{X: lineWrapRect.MaxX(), Y: lineWrapRect.MinY()}, Size: measureSize}
				imageRect := imageFrame.Inset(decoded.Margin)
				p.gp.SetX(imageRect.MinX())
				p.gp.SetY(imageRect.MinY())

				// DRAW
				p.drawImage(documentConfigure, decoded, imageRect, imageFrame)

				lineWrapRect = lineWrapRect.Merge(imageFrame)
			}

			wrapRect = wrapRect.Merge(lineWrapRect)
		}

	}

	return wrapRect
}

// 計算：テキストのサイズ
func (p *PDF) measureText(documentConfigure types.DocumentConfigure, decoded types.ElementText) types.Size {
	if p.isMultiLineText(decoded.Text) {
		measureSize := types.Size{}
		measureHeight := documentConfigure.FontHeight() * (float64(documentConfigure.TextSize) / 1000.0)

		texts := strings.Split(decoded.Text, "\n")
		for _, text := range texts {
			measureWidth, _ := p.gp.MeasureTextWidth(text)
			if measureSize.Width < measureWidth {
				measureSize.Width = measureWidth
			}
			measureSize.Height += measureHeight
		}

		if decoded.Size.Width != UnsetWidth && decoded.Size.Height == UnsetHeight {
			measureSize.Width = decoded.Size.Width
		}
		if decoded.Size.Width == UnsetWidth && decoded.Size.Height != UnsetHeight {
			measureSize.Height = decoded.Size.Height
		}

		measureSize.Width += decoded.Margin.Horizontal()
		measureSize.Height += decoded.Margin.Vertical()

		return measureSize
	}

	measureWidth, _ := p.gp.MeasureTextWidth(decoded.Text)
	measureHeight := documentConfigure.FontHeight() * (float64(documentConfigure.TextSize) / 1000.0)

	var measureSize types.Size
	if decoded.Size.Width != UnsetWidth && decoded.Size.Height != UnsetHeight {
		measureSize = types.Size{Width: decoded.Size.Width, Height: decoded.Size.Height}
	} else if decoded.Size.Width != UnsetWidth && decoded.Size.Height == UnsetHeight {
		measureSize = types.Size{Width: decoded.Size.Width, Height: measureHeight}
	} else if decoded.Size.Width == UnsetWidth && decoded.Size.Height != UnsetHeight {
		measureSize = types.Size{Width: measureWidth, Height: decoded.Size.Height}
	} else {
		measureSize = types.Size{Width: measureWidth, Height: measureHeight}
	}

	measureSize.Width += decoded.Margin.Horizontal()
	measureSize.Height += decoded.Margin.Vertical()

	return measureSize
}

// 計算：画像のサイズ
func (p *PDF) measureImage(documentConfigure types.DocumentConfigure, decoded types.ElementImage) types.Size {
	file, _ := os.Open(decoded.Path)
	imgConfig, _, _ := image.DecodeConfig(file)
	_ = file.Close()

	var measureSize types.Size
	if decoded.Size.Width != UnsetWidth && decoded.Size.Height != UnsetHeight && decoded.Size.Width < float64(imgConfig.Width) && decoded.Size.Height < float64(imgConfig.Height) {
		measureSize.Width = decoded.Size.Width
		measureSize.Height = decoded.Size.Height
	} else if decoded.Size.Width == UnsetWidth && decoded.Size.Height != UnsetHeight && decoded.Size.Height < float64(imgConfig.Height) {
		measureSize.Height = decoded.Size.Height
		measureSize.Width = float64(imgConfig.Width) * (measureSize.Height / float64(imgConfig.Height))
	} else if decoded.Size.Width != UnsetWidth && decoded.Size.Height == UnsetHeight && decoded.Size.Width < float64(imgConfig.Width) {
		measureSize.Width = decoded.Size.Width
		measureSize.Height = float64(imgConfig.Height) * (measureSize.Width / float64(imgConfig.Width))
	} else {
		if float64(imgConfig.Width) <= documentConfigure.Width-p.gp.MarginLeft()-p.gp.MarginRight() {
			measureSize.Width = float64(imgConfig.Width)
			measureSize.Height = float64(imgConfig.Height)
		} else {
			measureSize.Width = documentConfigure.Width - p.gp.MarginLeft() - p.gp.MarginRight()
			measureSize.Height = float64(imgConfig.Height) * (measureSize.Width / float64(imgConfig.Width))
		}
	}

	measureSize.Width += decoded.Margin.Horizontal()
	measureSize.Height += decoded.Margin.Vertical()

	return measureSize
}

// 描画：テキスト
func (p *PDF) drawText(documentConfigure types.DocumentConfigure, decoded types.ElementText, textRect types.Rect, textFrame types.Rect) {
	// BORDER, FILL
	if decoded.Border.Width != UnsetWidth {
		p.gp.SetLineWidth(decoded.Border.Width)
		p.gp.SetStrokeColor(decoded.Border.Color.R, decoded.Border.Color.G, decoded.Border.Color.B)
		if decoded.BackgroundColor.R != DefaultColorR || decoded.BackgroundColor.G != DefaultColorG || decoded.BackgroundColor.B != DefaultColorB {
			p.gp.SetFillColor(decoded.BackgroundColor.R, decoded.BackgroundColor.G, decoded.BackgroundColor.B)
			p.gp.RectFromUpperLeftWithStyle(textFrame.MinX(), textFrame.MinY(), textFrame.Width(), textFrame.Height(), "FD")
		} else {
			p.gp.RectFromUpperLeft(textFrame.MinX(), textFrame.MinY(), textFrame.Width(), textFrame.Height())
		}
	} else if decoded.BorderTop.Width != UnsetWidth {
		p.gp.SetLineWidth(decoded.BorderTop.Width)
		p.gp.SetStrokeColor(decoded.BorderTop.Color.R, decoded.BorderTop.Color.G, decoded.BorderTop.Color.B)
		p.gp.Line(textFrame.MinX(), textFrame.MinY(), textFrame.MinX()+textFrame.Width(), textFrame.MinY())
	} else if decoded.BorderRight.Width != UnsetWidth {
		p.gp.SetLineWidth(decoded.BorderRight.Width)
		p.gp.SetStrokeColor(decoded.BorderRight.Color.R, decoded.BorderRight.Color.G, decoded.BorderRight.Color.B)
		p.gp.Line(textFrame.MinX()+textFrame.Width(), textFrame.MinY(), textFrame.MinX()+textFrame.Width(), textFrame.MinY()+textFrame.Height())
	} else if decoded.BorderBottom.Width != UnsetWidth {
		p.gp.SetLineWidth(decoded.BorderBottom.Width)
		p.gp.SetStrokeColor(decoded.BorderBottom.Color.R, decoded.BorderBottom.Color.G, decoded.BorderBottom.Color.B)
		p.gp.Line(textFrame.MinX()+textFrame.Width(), textFrame.MinY()+textFrame.Height(), textFrame.MinX(), textFrame.MinY()+textFrame.Height())
	} else if decoded.BorderLeft.Width != UnsetWidth {
		p.gp.SetLineWidth(decoded.BorderLeft.Width)
		p.gp.SetStrokeColor(decoded.BorderLeft.Color.R, decoded.BorderLeft.Color.G, decoded.BorderLeft.Color.B)
		p.gp.Line(textFrame.MinX(), textFrame.MinY()+textFrame.Height(), textFrame.MinX(), textFrame.MinY())
	} else if decoded.BackgroundColor.R != DefaultColorR || decoded.BackgroundColor.G != DefaultColorG || decoded.BackgroundColor.B != DefaultColorB {
		p.gp.SetFillColor(decoded.BackgroundColor.R, decoded.BackgroundColor.G, decoded.BackgroundColor.B)
		p.gp.RectFromUpperLeftWithStyle(textFrame.MinX(), textFrame.MinY(), textFrame.Width(), textFrame.Height(), "F")
	}

	var gpRect = gopdf.Rect{W: textRect.Width(), H: textRect.Height()}

	// WRAP TEXT
	if decoded.Wrap {
		p.gp.SetTextColor(decoded.Color.R, decoded.Color.G, decoded.Color.B)
		_ = p.gp.MultiCell(&gpRect, decoded.Text)
		p.gp.SetTextColor(documentConfigure.TextColor.R, documentConfigure.TextColor.G, documentConfigure.TextColor.B)
		return
	}

	// TEXT OPTION
	var option = gopdf.CellOption{
		Border: 0,
		Float:  gopdf.Right,
	}
	if decoded.Align.IsCenter() {
		option.Align = gopdf.Center
	} else if decoded.Align.IsRight() {
		option.Align = gopdf.Right
	} else {
		option.Align = gopdf.Left
	}
	if decoded.Valign.IsMiddle() {
		option.Align = option.Align | gopdf.Middle
	} else if decoded.Valign.IsBottom() {
		option.Align = option.Align | gopdf.Bottom
	} else {
		option.Align = option.Align | gopdf.Top
	}

	// TEXT COLOR
	p.gp.SetTextColor(decoded.Color.R, decoded.Color.G, decoded.Color.B)

	if p.isMultiLineText(decoded.Text) {
		option.Float = gopdf.Bottom
		texts := strings.Split(decoded.Text, "\n")
		gpRect = gopdf.Rect{W: textRect.Width(), H: textRect.Height() / float64(len(texts))}
		for _, text := range texts {
			_ = p.gp.CellWithOption(&gpRect, text, option)
		}
	} else {
		_ = p.gp.CellWithOption(&gpRect, decoded.Text, option)
	}

	// RESET TEXT COLOR
	p.gp.SetTextColor(documentConfigure.TextColor.R, documentConfigure.TextColor.G, documentConfigure.TextColor.B)
}

// 描画：画像
func (p *PDF) drawImage(documentConfigure types.DocumentConfigure, decoded types.ElementImage, imageRect types.Rect, imageFrame types.Rect) {
	file, _ := os.Open(decoded.Path)
	img, imgType, _ := image.Decode(file)
	_ = file.Close()

	var imageHoloder gopdf.ImageHolder

	// RESIZE
	if decoded.Resize {
		resizedImg := resize.Resize(uint(imageRect.Width())*decoded.Resolution, uint(imageRect.Height())*decoded.Resolution, img, resize.Lanczos3)

		resizedBuf := new(bytes.Buffer)
		switch imgType {
		case "png":
			if err := png.Encode(resizedBuf, resizedImg); err != nil {
				panic(err)
			}
		case "jpeg":
			if err := jpeg.Encode(resizedBuf, resizedImg, nil); err != nil {
				panic(err)
			}
		}

		_imageHolder, err := gopdf.ImageHolderByBytes(resizedBuf.Bytes())
		if err != nil {
			panic(err)
		}

		imageHoloder = _imageHolder
	} else {
		_imageHolder, err := gopdf.ImageHolderByPath(decoded.Path)
		if err != nil {
			panic(err)
		}

		imageHoloder = _imageHolder
	}

	// DRAW IMAGE
	var gpRect = gopdf.Rect{W: imageRect.Width(), H: imageRect.Height()}
	_ = p.gp.ImageByHolder(imageHoloder, imageRect.MinX(), imageRect.MinY(), &gpRect)

	// BORDER
	if decoded.Border.Width != UnsetWidth {
		p.gp.SetLineWidth(decoded.Border.Width)
		p.gp.SetStrokeColor(decoded.Border.Color.R, decoded.Border.Color.G, decoded.Border.Color.B)
		p.gp.RectFromUpperLeft(imageFrame.MinX(), imageFrame.MinY(), imageFrame.Width(), imageFrame.Height())
	} else if decoded.BorderTop.Width != UnsetWidth {
		p.gp.SetLineWidth(decoded.BorderTop.Width)
		p.gp.SetStrokeColor(decoded.BorderTop.Color.R, decoded.BorderTop.Color.G, decoded.BorderTop.Color.B)
		p.gp.Line(imageFrame.MinX(), imageFrame.MinY(), imageFrame.MinX()+imageFrame.Width(), imageFrame.MinY())
	} else if decoded.BorderRight.Width != UnsetWidth {
		p.gp.SetLineWidth(decoded.BorderRight.Width)
		p.gp.SetStrokeColor(decoded.BorderRight.Color.R, decoded.BorderRight.Color.G, decoded.BorderRight.Color.B)
		p.gp.Line(imageFrame.MinX()+imageFrame.Width(), imageFrame.MinY(), imageFrame.MinX()+imageFrame.Width(), imageFrame.MinY()+imageFrame.Height())
	} else if decoded.BorderBottom.Width != UnsetWidth {
		p.gp.SetLineWidth(decoded.BorderBottom.Width)
		p.gp.SetStrokeColor(decoded.BorderBottom.Color.R, decoded.BorderBottom.Color.G, decoded.BorderBottom.Color.B)
		p.gp.Line(imageFrame.MinX()+imageFrame.Width(), imageFrame.MinY()+imageFrame.Height(), imageFrame.MinX(), imageFrame.MinY()+imageFrame.Height())
	} else if decoded.BorderLeft.Width != UnsetWidth {
		p.gp.SetLineWidth(decoded.BorderLeft.Width)
		p.gp.SetStrokeColor(decoded.BorderLeft.Color.R, decoded.BorderLeft.Color.G, decoded.BorderLeft.Color.B)
		p.gp.Line(imageFrame.MinX(), imageFrame.MinY()+imageFrame.Height(), imageFrame.MinX(), imageFrame.MinY())
	}
}

// ヘッダー
func (p *PDF) drawHeader(documentConfigure types.DocumentConfigure) {
	p.gp.SetX(p.headerRect.MinX())
	p.gp.SetY(p.headerRect.MinY())

	// > debug
	//p.gp.SetStrokeColor(255, 0, 0)
	//p.gp.RectFromUpperLeft(p.headerRect.Origin.X, p.headerRect.Origin.Y, p.headerRect.Width(), p.headerRect.Height())
	// < debug

	p.drawHeaderOrFooter(documentConfigure, documentConfigure.Header.LinerLayout)
}

// フッター
func (p *PDF) drawFooter(documentConfigure types.DocumentConfigure) {
	p.gp.SetX(p.footerRect.MinX())
	p.gp.SetY(p.footerRect.MinY())

	// > debug
	//p.gp.SetStrokeColor(0, 255, 0)
	//p.gp.RectFromUpperLeft(p.footerRect.Origin.X, p.footerRect.Origin.Y, p.footerRect.Width(), p.footerRect.Height())
	// < debug

	p.drawHeaderOrFooter(documentConfigure, documentConfigure.Footer.LinerLayout)
}

// 縦
func (p *PDF) breakVertical(lineWrapRect *types.Rect) {
	lineWrapRect.Origin.X = p.gp.GetX()
	lineWrapRect.Origin.Y = p.gp.GetY()
	lineWrapRect.Size.Width = 0
	lineWrapRect.Size.Height = 0
}

// 改行
func (p *PDF) breakLine(lineWrapRect *types.Rect, lineHeight float64) {
	//fmt.Printf("lineWrapRect(before): %v\n", lineWrapRect)

	//fmt.Printf("y: %v\n", lineWrapRect.Origin.Y)
	//fmt.Printf("h: %v\n", lineWrapRect.Size.Height)
	//fmt.Printf("maxy: %v\n", lineWrapRect.MaxY())
	//fmt.Printf("lineHeight: %v\n", lineHeight)

	lineWrapRect.Origin.X = lineWrapRect.MinX()
	if lineHeight == UnsetHeight {
		lineWrapRect.Origin.Y = lineWrapRect.MaxY()
	} else {
		lineWrapRect.Origin.Y = lineWrapRect.MinY() + lineHeight
	}
	lineWrapRect.Size.Width = 0
	lineWrapRect.Size.Height = 0

	//fmt.Printf("lineWrapRect(after): %v\n", lineWrapRect)
}

// 改ページ
func (p *PDF) breakPage(lineWrapRect *types.Rect, wrapRect *types.Rect) {
	lineWrapRect.Origin.X = p.contentRect.MinX()
	lineWrapRect.Origin.Y = p.contentRect.MinY()
	lineWrapRect.Size.Width = 0
	lineWrapRect.Size.Width = 0

	wrapRect.Origin.X = p.contentRect.MinX()
	wrapRect.Origin.Y = p.contentRect.MinY()
	wrapRect.Size.Width = 0
	wrapRect.Size.Width = 0
}

// 判定：複数行
func (p *PDF) isMultiLineText(text string) bool {
	return strings.Contains(text, "\n")
}

// 判定：改行
func (p *PDF) needLineBreak(documentConfigure types.DocumentConfigure, lineWrapRect types.Rect, measureSize types.Size) bool {
	if lineWrapRect.MaxX()+measureSize.Width > p.contentRect.MaxX() {
		return true
	}
	return false
}

// 判定：ページ
func (p *PDF) needPageBreak(documentConfigure types.DocumentConfigure, lineWrapRect types.Rect, measureSize types.Size) bool {
	if lineWrapRect.MinY()+measureSize.Height > p.contentRect.MaxY() {
		return true
	}
	return false
}
