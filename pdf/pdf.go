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
)

type PDF struct {
	currentHeight float64
	gp            gopdf.GoPdf
}

func (p *PDF) maxHeight() float64 {
	return p.currentHeight
}

func (p *PDF) setMaxHeight(lineHeight float64) {
	if p.currentHeight < lineHeight {
		p.currentHeight = lineHeight
	}
}

func (p *PDF) clearCurrentHeight() {
	p.currentHeight = 0
}

func (p *PDF) Draw(documentConfigure types.DocumentConfigure) {
	p.gp = gopdf.GoPdf{}

	if documentConfigure.Password == "" {
		p.gp.Start(gopdf.Config{PageSize: gopdf.Rect{W: documentConfigure.Width, H: documentConfigure.Height}, Unit: gopdf.Unit_PT})
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
		p.draw(documentConfigure, page.LinerLayout)
	}
}

func (p *PDF) draw(documentConfigure types.DocumentConfigure, linerLayout types.LinerLayout) {
	for _, element := range linerLayout.Elements {
		if element.Type.IsLineBreak() {
			var decoded types.ElementLineBreak
			_ = json.Unmarshal(element.Attributes, &decoded)
			p.gp.Br(decoded.Height)
			p.clearCurrentHeight()

		} else if element.Type.IsText() {
			var decoded = types.ElementText{
				Color:           types.Color{R: documentConfigure.TextColor.R, G: documentConfigure.TextColor.G, B: documentConfigure.TextColor.B},
				BackgroundColor: types.Color{R: 0, G: 0, B: 0},
				Width:           -1,
				Height:          -1,
				Border:          types.Border{Width: -1, Color: types.Color{R: 0, B: 0, G: 0}},
				BorderTop:       types.Border{Width: -1, Color: types.Color{R: 0, B: 0, G: 0}},
				BorderRight:     types.Border{Width: -1, Color: types.Color{R: 0, B: 0, G: 0}},
				BorderBottom:    types.Border{Width: -1, Color: types.Color{R: 0, B: 0, G: 0}},
				BorderLeft:      types.Border{Width: -1, Color: types.Color{R: 0, B: 0, G: 0}},
			}
			_ = json.Unmarshal(element.Attributes, &decoded)
			p.drawText(documentConfigure, linerLayout, decoded)

		} else if element.Type.IsImage() {
			var decoded = types.ElementImage{
				X:      -1,
				Y:      -1,
				Width:  -1,
				Height: -1,
				Resize: false,
			}
			_ = json.Unmarshal(element.Attributes, &decoded)
			p.drawImage(documentConfigure, linerLayout, decoded)

		}
	}

	for _, linerLayout := range linerLayout.LinearLayouts {
		p.draw(documentConfigure, linerLayout)
	}
}

func (p *PDF) drawText(documentConfigure types.DocumentConfigure, linerLayout types.LinerLayout, decoded types.ElementText) {
	x := p.gp.GetX()
	width := documentConfigure.Width - p.gp.MarginLeft() - p.gp.MarginRight()
	height := documentConfigure.Height - p.gp.MarginTop() - p.gp.MarginBottom()

	measureWidth, _ := p.gp.MeasureTextWidth(decoded.Text)
	measureHeight := documentConfigure.FontHeight() * (float64(documentConfigure.TextSize) / 1000.0)

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

	p.gp.SetTextColor(decoded.Color.R, decoded.Color.G, decoded.Color.B)

	if linerLayout.Orientation.IsHorizontal() {
		// LINE BREAK
		if x+textRect.W > width {
			if lineHeight := linerLayout.LineHeight; lineHeight != 0 {
				p.gp.Br(lineHeight)
			} else {
				p.gp.Br(p.maxHeight())
			}
			p.clearCurrentHeight()
		}

		// PAGE BREAK
		if p.gp.GetY()+textRect.H > height && documentConfigure.AutoPageBreak {
			p.gp.AddPage()
			p.clearCurrentHeight()
		}

		// BORDER, FILL
		if decoded.Border.Width != -1 {
			p.gp.SetLineWidth(decoded.Border.Width)
			p.gp.SetStrokeColor(decoded.Border.Color.R, decoded.Border.Color.G, decoded.Border.Color.B)
			if decoded.BackgroundColor.R != 0 || decoded.BackgroundColor.G != 0 || decoded.BackgroundColor.B != 0 {
				p.gp.SetFillColor(decoded.BackgroundColor.R, decoded.BackgroundColor.G, decoded.BackgroundColor.B)
				p.gp.RectFromUpperLeftWithStyle(p.gp.GetX(), p.gp.GetY(), textRect.W, textRect.H, "FD")
			} else {
				p.gp.RectFromUpperLeft(p.gp.GetX(), p.gp.GetY(), textRect.W, textRect.H)
			}
		} else if decoded.BorderTop.Width != -1 {
			p.gp.SetLineWidth(decoded.BorderTop.Width)
			p.gp.SetStrokeColor(decoded.BorderTop.Color.R, decoded.BorderTop.Color.G, decoded.BorderTop.Color.B)
			p.gp.Line(p.gp.GetX(), p.gp.GetY(), p.gp.GetX()+textRect.W, p.gp.GetY())
		} else if decoded.BorderRight.Width != -1 {
			p.gp.SetLineWidth(decoded.BorderRight.Width)
			p.gp.SetStrokeColor(decoded.BorderRight.Color.R, decoded.BorderRight.Color.G, decoded.BorderRight.Color.B)
			p.gp.Line(p.gp.GetX()+textRect.W, p.gp.GetY(), p.gp.GetX()+textRect.W, p.gp.GetY()+textRect.H)
		} else if decoded.BorderBottom.Width != -1 {
			p.gp.SetLineWidth(decoded.BorderBottom.Width)
			p.gp.SetStrokeColor(decoded.BorderBottom.Color.R, decoded.BorderBottom.Color.G, decoded.BorderBottom.Color.B)
			p.gp.Line(p.gp.GetX()+textRect.W, p.gp.GetY()+textRect.H, p.gp.GetX(), p.gp.GetY()+textRect.H)
		} else if decoded.BorderLeft.Width != -1 {
			p.gp.SetLineWidth(decoded.BorderLeft.Width)
			p.gp.SetStrokeColor(decoded.BorderLeft.Color.R, decoded.BorderLeft.Color.G, decoded.BorderLeft.Color.B)
			p.gp.Line(p.gp.GetX(), p.gp.GetY()+textRect.H, p.gp.GetX(), p.gp.GetY())
		}

		// ALIGN & VALIGN
		if decoded.Align.IsCenter() {
			p.gp.SetX(p.gp.GetX() + ((textRect.W / 2) - (measureWidth / 2)))
		} else if decoded.Align.IsRight() {
			p.gp.SetX(p.gp.GetX() + textRect.W - measureWidth)
		}
		if decoded.Valign.IsMiddle() {
			p.gp.SetY(p.gp.GetY() + ((textRect.H / 2) - (measureHeight / 2)))
		} else if decoded.Valign.IsBottom() {
			p.gp.SetY(p.gp.GetY() + textRect.H - measureHeight)
		}

		// DRAW TEXT
		_ = p.gp.Cell(&textRect, decoded.Text)

		// STORE MAX HEIGHT
		p.setMaxHeight(textRect.H)

		// RESET ALIGN & VALIGN
		if decoded.Align.IsCenter() {
			p.gp.SetX(p.gp.GetX() - ((textRect.W / 2) - (measureWidth / 2)))
		}
		if decoded.Valign.IsMiddle() {
			p.gp.SetY(p.gp.GetY() - ((textRect.H / 2) - (measureHeight / 2)))
		} else if decoded.Valign.IsMiddle() {
			p.gp.SetY(p.gp.GetY() - textRect.H + measureHeight)
		}
	} else if linerLayout.Orientation.IsVertical() {
		p.clearCurrentHeight()

		// PAGE BREAK
		if p.gp.GetY()+textRect.H > height && documentConfigure.AutoPageBreak {
			p.gp.AddPage()
		}

		// TODO: horizontal で実装した機能をこちらでも実装

		// DRAW TEXT
		_ = p.gp.Cell(&textRect, decoded.Text)

		p.gp.Br(textRect.H)
	}

	p.gp.SetTextColor(documentConfigure.TextColor.R, documentConfigure.TextColor.G, documentConfigure.TextColor.B)
}

func (p *PDF) drawImage(documentConfigure types.DocumentConfigure, linerLayout types.LinerLayout, decoded types.ElementImage) {
	height := documentConfigure.Height - p.gp.MarginTop() - p.gp.MarginBottom()

	file, _ := os.Open(decoded.Path)
	imgConfig, _, _ := image.DecodeConfig(file)

	_, _ = file.Seek(0, 0)
	img, imgType, _ := image.Decode(file)
	_ = file.Close()

	imageRect := gopdf.Rect{}
	if decoded.Width != -1 && decoded.Height != -1 {
		imageRect.W = decoded.Width
		imageRect.H = decoded.Height
	} else if decoded.Width == -1 && decoded.Height == -1 {
		imageRect.W = float64(imgConfig.Width)
		imageRect.H = float64(imgConfig.Height)
	} else if decoded.Width == -1 && decoded.Height != -1 {
		imageRect.H = decoded.Height
		imageRect.W = float64(imgConfig.Width) * (imageRect.H / float64(imgConfig.Height))
	} else if decoded.Width != -1 && decoded.Height == -1 {
		imageRect.W = decoded.Width
		imageRect.H = float64(imgConfig.Height) * (imageRect.W / float64(imgConfig.Width))
	}

	// LINE BREAK
	if p.gp.GetX()+imageRect.W > documentConfigure.Width {
		if lineHeight := linerLayout.LineHeight; lineHeight != 0 {
			p.gp.Br(lineHeight)
		} else {
			p.gp.Br(p.maxHeight())
		}
		p.clearCurrentHeight()
	}

	// PAGE BREAK
	if p.gp.GetY()+imageRect.H > height && documentConfigure.AutoPageBreak {
		p.gp.AddPage()
		p.clearCurrentHeight()
	}

	// STORE MAX HEIGHT
	p.setMaxHeight(imageRect.H)

	// RESIZE
	if decoded.Resize && ((decoded.Width != -1 && decoded.Width < float64(imgConfig.Width)) || (decoded.Height != -1 && decoded.Height < float64(imgConfig.Height))) {
		resizedImg := resize.Resize(uint(imageRect.W)*2, uint(imageRect.H)*2, img, resize.Lanczos3)

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

		imageHoloder, err := gopdf.ImageHolderByBytes(resizedBuf.Bytes())
		if err != nil {
			panic(err)
		}

		// DRAW IMAGE
		if decoded.X != -1 || decoded.Y != -1 {
			_ = p.gp.ImageByHolder(imageHoloder, decoded.X, decoded.Y, &imageRect)
		} else {
			_ = p.gp.ImageByHolder(imageHoloder, p.gp.GetX(), p.gp.GetY(), &imageRect)

			// TODO: vertical のときの動きを修正

			if linerLayout.Orientation.IsHorizontal() {
				p.gp.SetX(p.gp.GetX() + imageRect.W)
			} else if linerLayout.Orientation.IsVertical() {
				p.gp.SetY(p.gp.GetY() + imageRect.H)
			}
		}

		return
	}

	// DRAW IMAGE
	if decoded.X != -1 || decoded.Y != -1 {
		_ = p.gp.Image(decoded.Path, decoded.X, decoded.Y, &imageRect)
	} else {
		_ = p.gp.Image(decoded.Path, p.gp.GetX(), p.gp.GetY(), &imageRect)

		// TODO: vertical のときの動きを修正

		if linerLayout.Orientation.IsHorizontal() {
			p.gp.SetX(p.gp.GetX() + imageRect.W)
		} else if linerLayout.Orientation.IsVertical() {
			p.gp.SetY(p.gp.GetY() + imageRect.H)
		}
	}
}

func (p *PDF) Save(outputPath string) error {
	return p.gp.WritePdf(outputPath)
}

func (p *PDF) Destroy() {
	_ = p.gp.Close()
}
