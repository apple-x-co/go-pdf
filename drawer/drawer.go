package drawer

import (
	"apple-x-co/go-pdf/types"
	"bytes"
	"encoding/json"
	"github.com/nfnt/resize"
	"github.com/signintech/gopdf"
	"image"
	"image/jpeg"
	"image/png"
	"os"
)

type context struct {
	currentHeight float64
}

func (c *context) MaxHeight() float64 {
	return c.currentHeight
}
func (c *context) SetMaxHeight(lineHeight float64) {
	if c.currentHeight < lineHeight {
		c.currentHeight = lineHeight
	}
}
func (c *context) ClearCurrentHeight() {
	c.currentHeight = 0
}

func Draw(gp *gopdf.GoPdf, pdf types.DocumentConfigure, linerLayout types.LinerLayout) {
	//fmt.Printf("orientation: %v\n", linerLayout.Orientation)

	ctx := context{currentHeight: 0}

	draw(gp, pdf, linerLayout, &ctx)
}

func draw(gp *gopdf.GoPdf, pdf types.DocumentConfigure, linerLayout types.LinerLayout, ctx *context) {
	for _, element := range linerLayout.Elements {
		switch element.Type {
		case "line_break":
			var decoded types.ElementLineBreak
			_ = json.Unmarshal(element.Attributes, &decoded)
			gp.Br(decoded.Height)
			ctx.ClearCurrentHeight()

		case "text":
			var decoded = types.ElementText{
				Color:           types.Color{R: pdf.TextColor.R, G: pdf.TextColor.G, B: pdf.TextColor.B},
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
			drawText(gp, pdf, linerLayout, decoded, ctx)

		case "image":
			var decoded = types.ElementImage{
				X:      -1,
				Y:      -1,
				Width:  -1,
				Height: -1,
				Resize: false,
			}
			_ = json.Unmarshal(element.Attributes, &decoded)
			drawImage(gp, pdf, linerLayout, decoded, ctx)
		}
	}

	for _, linerLayout := range linerLayout.LinearLayouts {
		draw(gp, pdf, linerLayout, ctx)
	}

}

func drawText(gp *gopdf.GoPdf, pdf types.DocumentConfigure, linerLayout types.LinerLayout, decoded types.ElementText, ctx *context) {
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
		// LINE BREAK
		if x+textRect.W > width {
			if lineHeight := linerLayout.LineHeight; lineHeight != 0 {
				gp.Br(lineHeight)
			} else {
				gp.Br(ctx.MaxHeight())
			}
			ctx.ClearCurrentHeight()
		}

		// PAGE BREAK
		if gp.GetY()+textRect.H > height && pdf.AutoPageBreak {
			gp.AddPage()
			ctx.ClearCurrentHeight()
		}

		// BORDER, FILL
		if decoded.Border.Width != -1 {
			gp.SetLineWidth(decoded.Border.Width)
			gp.SetStrokeColor(decoded.Border.Color.R, decoded.Border.Color.G, decoded.Border.Color.B)
			if decoded.BackgroundColor.R != 0 || decoded.BackgroundColor.G != 0 || decoded.BackgroundColor.B != 0 {
				gp.SetFillColor(decoded.BackgroundColor.R, decoded.BackgroundColor.G, decoded.BackgroundColor.B)
				gp.RectFromUpperLeftWithStyle(gp.GetX(), gp.GetY(), textRect.W, textRect.H, "FD")
			} else {
				gp.RectFromUpperLeft(gp.GetX(), gp.GetY(), textRect.W, textRect.H)
			}
		} else if decoded.BorderTop.Width != -1 {
			gp.SetLineWidth(decoded.BorderTop.Width)
			gp.SetStrokeColor(decoded.BorderTop.Color.R, decoded.BorderTop.Color.G, decoded.BorderTop.Color.B)
			gp.Line(gp.GetX(), gp.GetY(), gp.GetX()+textRect.W, gp.GetY())
		} else if decoded.BorderRight.Width != -1 {
			gp.SetLineWidth(decoded.BorderRight.Width)
			gp.SetStrokeColor(decoded.BorderRight.Color.R, decoded.BorderRight.Color.G, decoded.BorderRight.Color.B)
			gp.Line(gp.GetX()+textRect.W, gp.GetY(), gp.GetX()+textRect.W, gp.GetY()+textRect.H)
		} else if decoded.BorderBottom.Width != -1 {
			gp.SetLineWidth(decoded.BorderBottom.Width)
			gp.SetStrokeColor(decoded.BorderBottom.Color.R, decoded.BorderBottom.Color.G, decoded.BorderBottom.Color.B)
			gp.Line(gp.GetX()+textRect.W, gp.GetY()+textRect.H, gp.GetX(), gp.GetY()+textRect.H)
		} else if decoded.BorderLeft.Width != -1 {
			gp.SetLineWidth(decoded.BorderLeft.Width)
			gp.SetStrokeColor(decoded.BorderLeft.Color.R, decoded.BorderLeft.Color.G, decoded.BorderLeft.Color.B)
			gp.Line(gp.GetX(), gp.GetY()+textRect.H, gp.GetX(), gp.GetY())
		}

		// ALIGN & VALIGN
		if decoded.IsAlignCenter() {
			gp.SetX(gp.GetX() + ((textRect.W / 2) - (measureWidth / 2)))
		} else if decoded.IsAlignRight() {
			gp.SetX(gp.GetX() + textRect.W - measureWidth)
		}
		if decoded.IsValignMiddle() {
			gp.SetY(gp.GetY() + ((textRect.H / 2) - (measureHeight / 2)))
		} else if decoded.IsValignBottom() {
			gp.SetY(gp.GetY() + textRect.H - measureHeight)
		}

		// DRAW TEXT
		_ = gp.Cell(&textRect, decoded.Text)

		// STORE MAX HEIGHT
		ctx.SetMaxHeight(textRect.H)

		// RESET ALIGN & VALIGN
		if decoded.IsAlignCenter() {
			gp.SetX(gp.GetX() - ((textRect.W / 2) - (measureWidth / 2)))
		}
		if decoded.IsValignMiddle() {
			gp.SetY(gp.GetY() - ((textRect.H / 2) - (measureHeight / 2)))
		} else if decoded.IsValignBottom() {
			gp.SetY(gp.GetY() - textRect.H + measureHeight)
		}
	} else if linerLayout.IsVertical() {
		ctx.ClearCurrentHeight()

		// PAGE BREAK
		if gp.GetY()+textRect.H > height && pdf.AutoPageBreak {
			gp.AddPage()
		}

		// TODO: horizontal で実装した機能をこちらでも実装

		// DRAW TEXT
		_ = gp.Cell(&textRect, decoded.Text)

		gp.Br(textRect.H)
	}

	gp.SetTextColor(pdf.TextColor.R, pdf.TextColor.G, pdf.TextColor.B)
}

func drawImage(gp *gopdf.GoPdf, pdf types.DocumentConfigure, linerLayout types.LinerLayout, decoded types.ElementImage, ctx *context) {
	height := pdf.Height - gp.MarginTop() - gp.MarginBottom()

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
	if gp.GetX()+imageRect.W > pdf.Width {
		if lineHeight := linerLayout.LineHeight; lineHeight != 0 {
			gp.Br(lineHeight)
		} else {
			gp.Br(ctx.MaxHeight())
		}
		ctx.ClearCurrentHeight()
	}

	// PAGE BREAK
	if gp.GetY()+imageRect.H > height && pdf.AutoPageBreak {
		gp.AddPage()
		ctx.ClearCurrentHeight()
	}

	// STORE MAX HEIGHT
	ctx.SetMaxHeight(imageRect.H)

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
			_ = gp.ImageByHolder(imageHoloder, decoded.X, decoded.Y, &imageRect)
		} else {
			_ = gp.ImageByHolder(imageHoloder, gp.GetX(), gp.GetY(), &imageRect)

			// TODO: vertical のときの動きを修正

			if linerLayout.IsHorizontal() {
				gp.SetX(gp.GetX() + imageRect.W)
			} else if linerLayout.IsVertical() {
				gp.SetY(gp.GetY() + imageRect.H)
			}
		}

		return
	}

	// DRAW IMAGE
	if decoded.X != -1 || decoded.Y != -1 {
		_ = gp.Image(decoded.Path, decoded.X, decoded.Y, &imageRect)
	} else {
		_ = gp.Image(decoded.Path, gp.GetX(), gp.GetY(), &imageRect)

		// TODO: vertical のときの動きを修正

		if linerLayout.IsHorizontal() {
			gp.SetX(gp.GetX() + imageRect.W)
		} else if linerLayout.IsVertical() {
			gp.SetY(gp.GetY() + imageRect.H)
		}
	}
}
