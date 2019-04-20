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
	currentSize types.Size
	gp          gopdf.GoPdf
}

func (p *PDF) CurrentSize() types.Size {
	return p.currentSize
}
func (p *PDF) setCurrentSize(width float64, height float64) {
	if p.currentSize.Width < width {
		p.currentSize.Width = width
	}
	if p.currentSize.Height < height {
		p.currentSize.Height = height
	}
}
func (p *PDF) clearCurrentSize() {
	p.currentSize.Width = 0
	p.currentSize.Height = 0
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

	width := documentConfigure.Width - p.gp.MarginLeft() - p.gp.MarginRight()
	height := documentConfigure.Height - p.gp.MarginTop() - p.gp.MarginBottom()

	for _, page := range documentConfigure.Pages {
		p.gp.AddPage()
		p.clearCurrentSize()
		containerRect := types.Rect{
			Origin: types.Origin{X: p.gp.GetX(), Y: p.gp.GetY()},
			Size:   types.Size{Width: width, Height: height},
		}
		p.gp.SetStrokeColor(255, 0, 0)                                                                                              // debug
		p.gp.RectFromUpperLeft(containerRect.Origin.X, containerRect.Origin.Y, containerRect.Size.Width, containerRect.Size.Height) // debug
		p.draw(documentConfigure, page.LinerLayout, containerRect)
	}
}

func (p *PDF) draw(documentConfigure types.DocumentConfigure, linerLayout types.LinerLayout, containerRect types.Rect) {
	for _, element := range linerLayout.Elements {
		if element.Type.IsLineBreak() {
			var decoded types.ElementLineBreak
			_ = json.Unmarshal(element.Attributes, &decoded)
			p.gp.Br(decoded.Height)
			p.clearCurrentSize()

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
			p.drawText(documentConfigure, linerLayout, decoded, containerRect)

		} else if element.Type.IsImage() {
			var decoded = types.ElementImage{
				X:      -1,
				Y:      -1,
				Width:  -1,
				Height: -1,
				Resize: false,
			}
			_ = json.Unmarshal(element.Attributes, &decoded)
			p.drawImage(documentConfigure, linerLayout, decoded, containerRect)

		}
	}

	for _, _linerLayout := range linerLayout.LinearLayouts {
		_containerRect := types.Rect{
			Origin: types.Origin{X: p.gp.GetX(), Y: p.gp.GetY()},
			Size:   types.Size{Width: containerRect.Size.Width - p.gp.GetX() + p.gp.MarginLeft(), Height: containerRect.Size.Height - p.gp.GetY() + p.gp.MarginTop()},
		}

		p.gp.SetStrokeColor(255, 255, 0)                                                                                                // debug
		p.gp.RectFromUpperLeft(_containerRect.Origin.X, _containerRect.Origin.Y, _containerRect.Size.Width, _containerRect.Size.Height) // debug
		p.draw(documentConfigure, _linerLayout, _containerRect)
	}
}

func (p *PDF) drawText(documentConfigure types.DocumentConfigure, linerLayout types.LinerLayout, decoded types.ElementText, containerRect types.Rect) {
	x := p.gp.GetX()
	width := documentConfigure.Width - p.gp.MarginLeft() - p.gp.MarginRight()
	height := documentConfigure.Height - p.gp.MarginTop() - p.gp.MarginBottom()

	measureWidth, _ := p.gp.MeasureTextWidth(decoded.Text)
	measureHeight := documentConfigure.FontHeight() * (float64(documentConfigure.TextSize) / 1000.0)

	var textRectSize gopdf.Rect
	if decoded.Width != -1 && decoded.Height != -1 {
		textRectSize = gopdf.Rect{W: decoded.Width, H: decoded.Height}
	} else if decoded.Width != -1 && decoded.Height == -1 {
		textRectSize = gopdf.Rect{W: decoded.Width, H: measureHeight}
	} else if decoded.Width == -1 && decoded.Height != -1 {
		textRectSize = gopdf.Rect{W: measureWidth, H: decoded.Height}
	} else {
		textRectSize = gopdf.Rect{W: measureWidth, H: measureHeight}
	}

	p.gp.SetTextColor(decoded.Color.R, decoded.Color.G, decoded.Color.B)

	// VERTICAL
	if linerLayout.Orientation.IsVertical() && p.CurrentSize().Height != 0 {
		p.gp.SetY(p.gp.GetY() + p.CurrentSize().Height)
		p.clearCurrentSize()
	}

	// LINE BREAK
	if x+textRectSize.W > width {
		if lineHeight := linerLayout.LineHeight; lineHeight != 0 {
			p.gp.SetX(containerRect.Origin.X)
			p.gp.SetY(p.gp.GetY() + lineHeight)
		} else {
			p.gp.SetX(containerRect.Origin.X)
			p.gp.SetY(p.gp.GetY() + p.CurrentSize().Height)
		}
		p.clearCurrentSize()
	}

	// PAGE BREAK
	if p.gp.GetY()+textRectSize.H > height && documentConfigure.AutoPageBreak {
		p.gp.AddPage()
		p.clearCurrentSize()
	}

	// BORDER, FILL
	if decoded.Border.Width != -1 {
		p.gp.SetLineWidth(decoded.Border.Width)
		p.gp.SetStrokeColor(decoded.Border.Color.R, decoded.Border.Color.G, decoded.Border.Color.B)
		if decoded.BackgroundColor.R != 0 || decoded.BackgroundColor.G != 0 || decoded.BackgroundColor.B != 0 {
			p.gp.SetFillColor(decoded.BackgroundColor.R, decoded.BackgroundColor.G, decoded.BackgroundColor.B)
			p.gp.RectFromUpperLeftWithStyle(p.gp.GetX(), p.gp.GetY(), textRectSize.W, textRectSize.H, "FD")
		} else {
			p.gp.RectFromUpperLeft(p.gp.GetX(), p.gp.GetY(), textRectSize.W, textRectSize.H)
		}
	} else if decoded.BorderTop.Width != -1 {
		p.gp.SetLineWidth(decoded.BorderTop.Width)
		p.gp.SetStrokeColor(decoded.BorderTop.Color.R, decoded.BorderTop.Color.G, decoded.BorderTop.Color.B)
		p.gp.Line(p.gp.GetX(), p.gp.GetY(), p.gp.GetX()+textRectSize.W, p.gp.GetY())
	} else if decoded.BorderRight.Width != -1 {
		p.gp.SetLineWidth(decoded.BorderRight.Width)
		p.gp.SetStrokeColor(decoded.BorderRight.Color.R, decoded.BorderRight.Color.G, decoded.BorderRight.Color.B)
		p.gp.Line(p.gp.GetX()+textRectSize.W, p.gp.GetY(), p.gp.GetX()+textRectSize.W, p.gp.GetY()+textRectSize.H)
	} else if decoded.BorderBottom.Width != -1 {
		p.gp.SetLineWidth(decoded.BorderBottom.Width)
		p.gp.SetStrokeColor(decoded.BorderBottom.Color.R, decoded.BorderBottom.Color.G, decoded.BorderBottom.Color.B)
		p.gp.Line(p.gp.GetX()+textRectSize.W, p.gp.GetY()+textRectSize.H, p.gp.GetX(), p.gp.GetY()+textRectSize.H)
	} else if decoded.BorderLeft.Width != -1 {
		p.gp.SetLineWidth(decoded.BorderLeft.Width)
		p.gp.SetStrokeColor(decoded.BorderLeft.Color.R, decoded.BorderLeft.Color.G, decoded.BorderLeft.Color.B)
		p.gp.Line(p.gp.GetX(), p.gp.GetY()+textRectSize.H, p.gp.GetX(), p.gp.GetY())
	}

	// ALIGN & VALIGN
	if decoded.Align.IsCenter() {
		p.gp.SetX(p.gp.GetX() + ((textRectSize.W / 2) - (measureWidth / 2)))
	} else if decoded.Align.IsRight() {
		p.gp.SetX(p.gp.GetX() + textRectSize.W - measureWidth)
	}
	if decoded.Valign.IsMiddle() {
		p.gp.SetY(p.gp.GetY() + ((textRectSize.H / 2) - (measureHeight / 2)))
	} else if decoded.Valign.IsBottom() {
		p.gp.SetY(p.gp.GetY() + textRectSize.H - measureHeight)
	}

	// DRAW TEXT
	_ = p.gp.Cell(&textRectSize, decoded.Text)

	// STORE MAX AXIS
	p.setCurrentSize(textRectSize.W, textRectSize.H)

	// RESET ALIGN & VALIGN
	if decoded.Align.IsCenter() {
		p.gp.SetX(p.gp.GetX() - ((textRectSize.W / 2) - (measureWidth / 2)))
	}
	if decoded.Valign.IsMiddle() {
		p.gp.SetY(p.gp.GetY() - ((textRectSize.H / 2) - (measureHeight / 2)))
	} else if decoded.Valign.IsMiddle() {
		p.gp.SetY(p.gp.GetY() - textRectSize.H + measureHeight)
	}

	p.gp.SetTextColor(documentConfigure.TextColor.R, documentConfigure.TextColor.G, documentConfigure.TextColor.B)

	// VERTICAL
	if linerLayout.Orientation.IsVertical() {
		p.gp.SetX(p.gp.GetX() - textRectSize.W)
	}
}

func (p *PDF) drawImage(documentConfigure types.DocumentConfigure, linerLayout types.LinerLayout, decoded types.ElementImage, containerRect types.Rect) {
	height := documentConfigure.Height - p.gp.MarginTop() - p.gp.MarginBottom()

	file, _ := os.Open(decoded.Path)
	imgConfig, _, _ := image.DecodeConfig(file)

	_, _ = file.Seek(0, 0)
	img, imgType, _ := image.Decode(file)
	_ = file.Close()

	var imageRectSize gopdf.Rect
	if decoded.Width != -1 && decoded.Height != -1 {
		imageRectSize.W = decoded.Width
		imageRectSize.H = decoded.Height
	} else if decoded.Width == -1 && decoded.Height == -1 {
		imageRectSize.W = float64(imgConfig.Width)
		imageRectSize.H = float64(imgConfig.Height)
	} else if decoded.Width == -1 && decoded.Height != -1 {
		imageRectSize.H = decoded.Height
		imageRectSize.W = float64(imgConfig.Width) * (imageRectSize.H / float64(imgConfig.Height))
	} else if decoded.Width != -1 && decoded.Height == -1 {
		imageRectSize.W = decoded.Width
		imageRectSize.H = float64(imgConfig.Height) * (imageRectSize.W / float64(imgConfig.Width))
	}

	// VERTICAL
	if linerLayout.Orientation.IsVertical() && p.CurrentSize().Height != 0 {
		p.gp.SetY(p.gp.GetY() + p.CurrentSize().Height)
		p.clearCurrentSize()
	}

	// LINE BREAK
	if p.gp.GetX()+imageRectSize.W > documentConfigure.Width {
		if lineHeight := linerLayout.LineHeight; lineHeight != 0 {
			p.gp.SetX(containerRect.Origin.X)
			p.gp.SetY(p.gp.GetY() + lineHeight)
		} else {
			p.gp.SetX(containerRect.Origin.X)
			p.gp.SetY(p.gp.GetY() + p.CurrentSize().Height)
		}
		p.clearCurrentSize()
	}

	// PAGE BREAK
	if p.gp.GetY()+imageRectSize.H > height && documentConfigure.AutoPageBreak {
		p.gp.AddPage()
		p.clearCurrentSize()
	}

	// STORE MAX AXIS
	p.setCurrentSize(imageRectSize.W, imageRectSize.H)

	var imageHoloder gopdf.ImageHolder

	// RESIZE
	if decoded.Resize && ((decoded.Width != -1 && decoded.Width < float64(imgConfig.Width)) || (decoded.Height != -1 && decoded.Height < float64(imgConfig.Height))) {
		resizedImg := resize.Resize(uint(imageRectSize.W)*2, uint(imageRectSize.H)*2, img, resize.Lanczos3)

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

		ih, err := gopdf.ImageHolderByBytes(resizedBuf.Bytes())
		if err != nil {
			panic(err)
		}

		imageHoloder = ih
	} else {
		ih, err := gopdf.ImageHolderByPath(decoded.Path)
		if err != nil {
			panic(err)
		}

		imageHoloder = ih
	}

	// DRAW IMAGE
	if decoded.X != -1 || decoded.Y != -1 {
		_ = p.gp.ImageByHolder(imageHoloder, decoded.X, decoded.Y, &imageRectSize)
		return
	}
	if linerLayout.Orientation.IsHorizontal() {
		_ = p.gp.Image(decoded.Path, p.gp.GetX(), p.gp.GetY(), &imageRectSize)
		p.gp.SetX(p.gp.GetX() + imageRectSize.W)
	} else if linerLayout.Orientation.IsVertical() {
		_ = p.gp.Image(decoded.Path, p.gp.GetX(), p.gp.GetY(), &imageRectSize)
	}
}

func (p *PDF) Save(outputPath string) error {
	return p.gp.WritePdf(outputPath)
}

func (p *PDF) Destroy() {
	_ = p.gp.Close()
}
