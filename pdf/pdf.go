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
	"math"
	"os"
	"strings"
	"text/template"
	"time"
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
const DefaultImageResolution uint = 2

type PDF struct {
	gp               gopdf.GoPdf
	contentRect      types.Rect
	commonHeaderRect types.Rect
	commonFooterRect types.Rect
	templates        map[string]interface{}
	pageNumber       uint
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
				Unit:     gopdf.UnitPT,
				Protection: gopdf.PDFProtectionConfig{
					UseProtection: true,
					Permissions:   gopdf.PermissionsPrint | gopdf.PermissionsCopy | gopdf.PermissionsModify,
					OwnerPass:     []byte(documentConfigure.Password),
					UserPass:      []byte(documentConfigure.Password),
				},
			})
	}

	p.gp.SetMargins(
		documentConfigure.Margin.Left,
		documentConfigure.Margin.Top,
		documentConfigure.Margin.Right,
		documentConfigure.Margin.Bottom,
	)
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
	documentConfigure.SetFontHeight(float64(parser.Ascender()+parser.XHeight()+parser.Descender()) * 1000.00 / float64(parser.UnitsPerEm()))

	p.gp.SetTextColor(documentConfigure.TextColor.R, documentConfigure.TextColor.G, documentConfigure.TextColor.B)

	// RECT
	if !documentConfigure.CommonHeader.Size.IsZero() {
		p.commonHeaderRect = types.Rect{
			Origin: types.Origin{
				X: p.gp.MarginLeft(),
				Y: p.gp.MarginTop(),
			},
			Size: documentConfigure.CommonHeader.Size,
		}
		if p.commonHeaderRect.Size.Width == 0 {
			p.commonHeaderRect.Size.Width = documentConfigure.Width - p.gp.MarginLeft() - p.gp.MarginRight()
		}
	}
	if !documentConfigure.CommonFooter.Size.IsZero() {
		p.commonFooterRect = types.Rect{
			Origin: types.Origin{
				X: p.gp.MarginLeft(),
				Y: documentConfigure.Height - p.gp.MarginBottom() - documentConfigure.CommonFooter.Size.Height,
			},
			Size: documentConfigure.CommonFooter.Size,
		}
		if p.commonFooterRect.Size.Width == 0 {
			p.commonFooterRect.Size.Width = documentConfigure.Width - p.gp.MarginLeft() - p.gp.MarginRight()
		}
	}
	p.contentRect = types.Rect{
		Origin: types.Origin{
			X: p.gp.MarginLeft(),
			Y: p.gp.MarginTop() + p.commonHeaderRect.Height(),
		},
		Size: types.Size{
			Width:  documentConfigure.Width - p.gp.MarginLeft() - p.gp.MarginRight(),
			Height: documentConfigure.Height - p.gp.MarginTop() - p.gp.MarginBottom() - p.commonHeaderRect.Height() - p.commonFooterRect.Height(),
		},
	}

	// ELEMENT TEMPLATES
	p.templates = map[string]interface{}{}

	for _, elementTemplate := range documentConfigure.Templates {
		//fmt.Printf("%v\n", elementTemplate.Id)
		if elementTemplate.Type.IsText() {
			var decoded = types.ElementText{
				TextSize:        documentConfigure.TextSize,
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
			_ = json.Unmarshal(elementTemplate.Attributes, &decoded)
			p.templates[elementTemplate.Id] = decoded
		} else if elementTemplate.Type.IsImage() {
			var decoded = types.ElementImage{
				Size:       types.Size{Width: UnsetWidth, Height: UnsetWidth},
				Origin:     types.Origin{X: UnsetWidth, Y: UnsetHeight},
				Resize:     false,
				Resolution: DefaultImageResolution,
			}
			_ = json.Unmarshal(elementTemplate.Attributes, &decoded)
			p.templates[elementTemplate.Id] = decoded
		}
	}
	//fmt.Printf("templates: %v\n", p.templates)

	// DRAW
	for _, page := range documentConfigure.Pages {
		p.gp.AddPage()
		p.pageNumber += 1

		// GLOBAL HEADER & FOOTER
		if !p.commonHeaderRect.Size.IsZero() {
			p.draw(documentConfigure, page, documentConfigure.CommonHeader.LinerLayout, p.commonHeaderRect, true, false)
		}
		if !p.commonFooterRect.Size.IsZero() {
			p.draw(documentConfigure, page, documentConfigure.CommonFooter.LinerLayout, p.commonFooterRect, true, true)
		}

		pageHeaderRect := types.Rect{}
		titleRect := types.Rect{}
		contentRect := p.contentRect

		// DRAW PAGE HEADER
		if !page.PageHeader.Size.IsZero() {
			pageHeaderRect = types.Rect{
				Origin: types.Origin{X: contentRect.MinX(), Y: contentRect.MinY()},
				Size:   page.PageHeader.Size,
			}
			if pageHeaderRect.Size.Width == UnsetWidth {
				pageHeaderRect.Size.Width = p.contentRect.Width()
			}
			contentRect = contentRect.ApplyMargin(types.Margin{
				Top: page.PageHeader.Size.Height,
			})
			p.draw(documentConfigure, page, page.PageHeader.LinerLayout, pageHeaderRect, true, false)
		}

		// DRAW FIXED TITLE
		if !page.FixedTitle.Size.IsZero() {
			titleRect = types.Rect{
				Origin: types.Origin{X: contentRect.MinX(), Y: contentRect.MinY()},
				Size:   page.FixedTitle.Size,
			}
			if titleRect.Size.Width == UnsetWidth {
				titleRect.Size.Width = p.contentRect.Width()
			}
			contentRect = contentRect.ApplyMargin(types.Margin{
				Top: page.FixedTitle.Size.Height,
			})
			p.draw(documentConfigure, page, page.FixedTitle.LinerLayout, titleRect, true, false)
		}

		// DRAW PAGE CONTENT
		wrapRect := p.draw(documentConfigure, page, page.LinerLayout, contentRect, true, false)
		//fmt.Printf("rect: %v\n", rect)

		// DRAW PAGE FOOTER
		pageFooterRect := types.Rect{}
		if !page.PageFooter.Size.IsZero() {
			pageFooterRect = types.Rect{
				Origin: types.Origin{X: wrapRect.MinX(), Y: wrapRect.MaxY()},
				Size:   page.PageFooter.Size,
			}
			if pageFooterRect.Size.Width == UnsetWidth {
				pageFooterRect.Size.Width = p.contentRect.Width()
			}
		}
		if !pageFooterRect.Size.IsZero() {
			p.draw(documentConfigure, page, page.PageFooter.LinerLayout, pageFooterRect, true, true)
		}
	}
}

func (p *PDF) Save(outputPath string) error {
	return p.gp.WritePdf(outputPath)
}

func (p *PDF) Destroy() {
	_ = p.gp.Close()
}

// 描画要素のループ
func (p *PDF) draw(documentConfigure types.DocumentConfigure, page types.Page, linerLayout types.LinerLayout, parentRect types.Rect, needMoveAxis bool, isFooter bool) types.Rect {
	if needMoveAxis {
		p.gp.SetX(parentRect.MinX())
		p.gp.SetY(parentRect.MinY())
	}

	var wrapRect = types.Rect{Origin: types.Origin{X: p.gp.GetX(), Y: p.gp.GetY()}}
	var lineWrapRect = types.Rect{Origin: types.Origin{X: p.gp.GetX(), Y: p.gp.GetY()}}
	var parentLayoutSize = p.calcLayoutSize(parentRect.Size, linerLayout.Layout)
	//fmt.Printf("parentLayoutSize: %v\n", parentLayoutSize)

	if len(linerLayout.Elements) > 0 {

		for _, element := range linerLayout.Elements {
			if element.Type.IsLineBreak() {
				var decoded = types.ElementLineBreak{
					Height: UnsetHeight,
				}
				_ = json.Unmarshal(element.Attributes, &decoded)
				p.breakLine(&lineWrapRect, decoded.Height)

			} else if element.Type.IsText() {
				var decoded types.ElementText
				if element.TemplateId != "" {
					templateText, ok := p.templates[element.TemplateId].(types.ElementText)
					if ok {
						decoded = templateText
					}
				}
				if decoded.TextSize == 0 {
					decoded = types.ElementText{
						TextSize:        documentConfigure.TextSize,
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
				}
				_ = json.Unmarshal(element.Attributes, &decoded)

				//fmt.Printf("---------------------------\n%v\n", decoded.Text)

				// BUILD TEXT
				vars := struct {
					PageNumber uint
					Now        string
				}{
					p.pageNumber,
					time.Now().Format("2006-01-02 15:04:05"),
				}
				tmpl, err := template.New("text").Parse(decoded.Text)
				if err != nil {
					panic(err)
				}
				var buf bytes.Buffer
				err = tmpl.Execute(&buf, vars)
				if err != nil {
					panic(err)
				}
				decoded.Text = buf.String()

				// ACTUAL SIZE
				measureSize := p.measureText(documentConfigure, decoded)

				// LAYOUT SIZE
				if decoded.Layout.Width.IsMatchParent() || decoded.Layout.Height.IsMatchParent() {
					elementLayoutSize := p.calcLayoutSize(parentLayoutSize, decoded.Layout)
					//fmt.Printf("elementLayoutSize: %v\n", elementLayoutSize)
					if elementLayoutSize.Width != UnsetWidth {
						measureSize.Width = elementLayoutSize.Width
					}
					if elementLayoutSize.Height != UnsetHeight {
						measureSize.Height = elementLayoutSize.Height
					}

					if decoded.Wrap && decoded.Size.IsZero() {
						texts, _ := p.gp.SplitText(decoded.Text, elementLayoutSize.Width)
						measureSize.Height = measureSize.Height * float64(len(texts))
					}
				}

				// TOTAL SIZE
				size := types.Size{Width: measureSize.Width + decoded.Margin.Horizontal(), Height: measureSize.Height + decoded.Margin.Vertical()}

				// DRAW FIXED POSITION
				if decoded.Origin.X != UnsetX && decoded.Origin.Y != UnsetY {
					textFrame := types.Rect{Origin: types.Origin{X: decoded.Origin.X, Y: decoded.Origin.Y}, Size: size}
					textRect := textFrame.ApplyMargin(decoded.Margin)
					textRect = textRect.ApplyContentMargin(decoded.ContentMargin)
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
				if p.needLineBreak(lineWrapRect, size) {
					//fmt.Print("> line break\n")
					p.breakLine(&lineWrapRect, linerLayout.LineHeight)
				}

				// PAGE BREAK
				if p.needPageBreak(lineWrapRect, size) && !isFooter {
					//fmt.Print("> page break\n")
					p.gp.AddPage()
					p.pageNumber += 1
					p.breakPage(&lineWrapRect, &wrapRect)

					if !p.commonHeaderRect.Size.IsZero() {
						p.draw(documentConfigure, page, documentConfigure.CommonHeader.LinerLayout, p.commonHeaderRect, true, false)
					}
					if !p.commonFooterRect.Size.IsZero() {
						p.draw(documentConfigure, page, documentConfigure.CommonFooter.LinerLayout, p.commonFooterRect, true, true)
					}

					// DRAW FIXED TITLE
					if !page.FixedTitle.Size.IsZero() {
						titleRect := types.Rect{
							Origin: types.Origin{X: wrapRect.MinX(), Y: wrapRect.MinY()},
							Size:   page.FixedTitle.Size,
						}
						if titleRect.Size.Width == UnsetWidth {
							titleRect.Size.Width = p.contentRect.Width()
						}
						p.draw(documentConfigure, page, page.FixedTitle.LinerLayout, titleRect, true, false)
						lineWrapRect = lineWrapRect.ApplyMargin(types.Margin{
							Top: titleRect.Size.Height,
						})
						wrapRect = wrapRect.ApplyMargin(types.Margin{
							Top: titleRect.Size.Height,
						})
					}

					p.gp.SetX(wrapRect.MinX())
					p.gp.SetY(wrapRect.MinY())
				}

				// DRAWABLE RECT
				textFrame := types.Rect{Origin: types.Origin{X: lineWrapRect.MaxX(), Y: lineWrapRect.MinY()}, Size: size}
				textRect := textFrame.ApplyMargin(decoded.Margin)
				textRect = textRect.ApplyContentMargin(decoded.ContentMargin)
				p.gp.SetX(textRect.MinX())
				p.gp.SetY(textRect.MinY())

				// DRAW
				//fmt.Printf("textRect: %v\n", textRect)
				//fmt.Printf("lineWrapRect: %v\n", lineWrapRect)
				p.drawText(documentConfigure, decoded, textRect, textFrame)

				lineWrapRect = lineWrapRect.Merge(textFrame)

			} else if element.Type.IsImage() {
				var decoded types.ElementImage
				if element.TemplateId != "" {
					templateImage, ok := p.templates[element.TemplateId].(types.ElementImage)
					if ok {
						decoded = templateImage
					}
				}
				if decoded.Resolution == 0 {
					decoded = types.ElementImage{
						Size:       types.Size{Width: UnsetWidth, Height: UnsetWidth},
						Origin:     types.Origin{X: UnsetWidth, Y: UnsetHeight},
						Resize:     false,
						Resolution: DefaultImageResolution,
					}
				}
				_ = json.Unmarshal(element.Attributes, &decoded)

				//fmt.Printf("---------------------------\n%v\n", decoded.Path)

				// Actual Size
				measureSize := p.measureImage(documentConfigure, decoded)

				// Layout Size
				if decoded.Layout.Width.IsMatchParent() || decoded.Layout.Height.IsMatchParent() {
					elementLayoutSize := p.calcLayoutSize(parentLayoutSize, decoded.Layout)
					//fmt.Printf("elementLayoutSize: %v\n", elementLayoutSize)
					if elementLayoutSize.Width != UnsetWidth && elementLayoutSize.Height == UnsetHeight {
						measureSize.Height = measureSize.Height * (elementLayoutSize.Width / measureSize.Width)
						measureSize.Width = elementLayoutSize.Width
					} else if elementLayoutSize.Width == UnsetWidth && elementLayoutSize.Height != UnsetHeight {
						measureSize.Width = measureSize.Width * (elementLayoutSize.Height / measureSize.Height)
						measureSize.Height = elementLayoutSize.Height
					}
				}

				// TOTAL SIZE
				size := types.Size{Width: measureSize.Width, Height: measureSize.Height}

				// DRAW FIXED POSITION
				if decoded.Origin.X != UnsetX && decoded.Origin.Y != UnsetY {
					imageFrame := types.Rect{Origin: types.Origin{X: decoded.Origin.X, Y: decoded.Origin.Y}, Size: size}
					imageRect := imageFrame.ApplyMargin(decoded.Margin)
					imageRect = imageRect.ApplyContentMargin(decoded.ContentMargin)
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
				if p.needLineBreak(lineWrapRect, size) {
					//fmt.Print("> line break\n")
					p.breakLine(&lineWrapRect, linerLayout.LineHeight)
				}

				// PAGE BREAK
				if p.needPageBreak(lineWrapRect, size) {
					//fmt.Print("> page break\n")
					p.gp.AddPage()
					p.pageNumber += 1
					p.breakPage(&lineWrapRect, &wrapRect)

					if !p.commonHeaderRect.Size.IsZero() {
						p.draw(documentConfigure, page, documentConfigure.CommonHeader.LinerLayout, p.commonHeaderRect, true, false)
					}
					if !p.commonFooterRect.Size.IsZero() {
						p.draw(documentConfigure, page, documentConfigure.CommonFooter.LinerLayout, p.commonFooterRect, true, true)
					}

					// DRAW FIXED TITLE
					if !page.FixedTitle.Size.IsZero() {
						titleRect := types.Rect{
							Origin: types.Origin{X: wrapRect.MinX(), Y: wrapRect.MinY()},
							Size:   page.FixedTitle.Size,
						}
						if titleRect.Size.Width == UnsetWidth {
							titleRect.Size.Width = p.contentRect.Width()
						}
						p.draw(documentConfigure, page, page.FixedTitle.LinerLayout, titleRect, true, false)
						lineWrapRect = lineWrapRect.ApplyMargin(types.Margin{
							Top: titleRect.Size.Height,
						})
						wrapRect = wrapRect.ApplyMargin(types.Margin{
							Top: titleRect.Size.Height,
						})
					}

					p.gp.SetX(wrapRect.MinX())
					p.gp.SetY(wrapRect.MinY())
				}

				// DRAWABLE RECT
				imageFrame := types.Rect{Origin: types.Origin{X: lineWrapRect.MaxX(), Y: lineWrapRect.MinY()}, Size: size}
				imageRect := imageFrame.ApplyMargin(decoded.Margin)
				imageRect = imageRect.ApplyContentMargin(decoded.ContentMargin)
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
		drawnRect := p.draw(documentConfigure, page, _linerLayout, parentRect, false, false)
		wrapRect = wrapRect.Merge(drawnRect)

		// > debug
		//p.gp.SetStrokeColor(255, 255, 0)
		//p.gp.RectFromUpperLeft(wrapRect.Origin.X, wrapRect.Origin.Y, wrapRect.Size.Width, wrapRect.Size.Height)
		// < debug

		if linerLayout.Orientation.IsHorizontal() {
			p.gp.SetX(drawnRect.MaxX())
			p.gp.SetY(drawnRect.MinY())
		} else if linerLayout.Orientation.IsVertical() {
			p.gp.SetX(drawnRect.MinX())
			p.gp.SetY(drawnRect.MaxY())
		}
	}

	return wrapRect
}

// 計算：テキストのサイズ
func (p *PDF) measureText(documentConfigure types.DocumentConfigure, decoded types.ElementText) types.Size {
	if err := p.gp.SetFont("default", "", decoded.TextSize); err != nil {
		log.Print(err.Error())
	}

	if p.isMultiLineText(decoded.Text) {
		measureSize := types.Size{}
		measureHeight := documentConfigure.FontHeight() * (float64(decoded.TextSize) / 1000.0)

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

		//measureSize.Width += decoded.Margin.Horizontal()
		//measureSize.Height += decoded.Margin.Vertical()

		return measureSize
	}

	measureWidth, _ := p.gp.MeasureTextWidth(decoded.Text)
	measureHeight := documentConfigure.FontHeight() * (float64(decoded.TextSize) / 1000.0)

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

	//measureSize.Width += decoded.Margin.Horizontal()
	//measureSize.Height += decoded.Margin.Vertical()

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

	//measureSize.Width += decoded.Margin.Horizontal()
	//measureSize.Height += decoded.Margin.Vertical()

	return measureSize
}

// 計算：レイアウトサイズ
func (p *PDF) calcLayoutSize(size types.Size, layout types.Layout) types.Size {
	var layoutSize = types.Size{Width: UnsetWidth, Height: UnsetHeight}
	if layout.Width.IsMatchParent() {
		layoutSize.Width = math.Trunc(size.Width * layout.Ratio)
	}
	if layout.Height.IsMatchParent() {
		layoutSize.Height = math.Trunc(size.Height * layout.Ratio)
	}
	return layoutSize
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
	} else {
		if decoded.BorderTop.Width != UnsetWidth {
			p.gp.SetLineWidth(decoded.BorderTop.Width)
			p.gp.SetStrokeColor(decoded.BorderTop.Color.R, decoded.BorderTop.Color.G, decoded.BorderTop.Color.B)
			p.gp.Line(textFrame.MinX(), textFrame.MinY(), textFrame.MinX()+textFrame.Width(), textFrame.MinY())
		}
		if decoded.BorderRight.Width != UnsetWidth {
			p.gp.SetLineWidth(decoded.BorderRight.Width)
			p.gp.SetStrokeColor(decoded.BorderRight.Color.R, decoded.BorderRight.Color.G, decoded.BorderRight.Color.B)
			p.gp.Line(textFrame.MinX()+textFrame.Width(), textFrame.MinY(), textFrame.MinX()+textFrame.Width(), textFrame.MinY()+textFrame.Height())
		}
		if decoded.BorderBottom.Width != UnsetWidth {
			p.gp.SetLineWidth(decoded.BorderBottom.Width)
			p.gp.SetStrokeColor(decoded.BorderBottom.Color.R, decoded.BorderBottom.Color.G, decoded.BorderBottom.Color.B)
			p.gp.Line(textFrame.MinX()+textFrame.Width(), textFrame.MinY()+textFrame.Height(), textFrame.MinX(), textFrame.MinY()+textFrame.Height())
		}
		if decoded.BorderLeft.Width != UnsetWidth {
			p.gp.SetLineWidth(decoded.BorderLeft.Width)
			p.gp.SetStrokeColor(decoded.BorderLeft.Color.R, decoded.BorderLeft.Color.G, decoded.BorderLeft.Color.B)
			p.gp.Line(textFrame.MinX(), textFrame.MinY()+textFrame.Height(), textFrame.MinX(), textFrame.MinY())
		}
	}

	if decoded.BackgroundColor.R != DefaultColorR || decoded.BackgroundColor.G != DefaultColorG || decoded.BackgroundColor.B != DefaultColorB {
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

	// TEXT SIZE
	if err := p.gp.SetFont("default", "", decoded.TextSize); err != nil {
		log.Print(err.Error())
	}

	// TEXT COLOR
	p.gp.SetFillColor(decoded.Color.R, decoded.Color.G, decoded.Color.B)
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

	// RESET COLOR
	p.gp.SetStrokeColor(documentConfigure.TextColor.R, documentConfigure.TextColor.G, documentConfigure.TextColor.B)
	p.gp.SetFillColor(documentConfigure.TextColor.R, documentConfigure.TextColor.G, documentConfigure.TextColor.B)
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
	lineWrapRect.Size.Height = 0

	wrapRect.Origin.X = p.contentRect.MinX()
	wrapRect.Origin.Y = p.contentRect.MinY()
	wrapRect.Size.Width = 0
	wrapRect.Size.Height = 0
}

// 判定：複数行
func (p *PDF) isMultiLineText(text string) bool {
	return strings.Contains(text, "\n")
}

// 判定：改行
func (p *PDF) needLineBreak(lineWrapRect types.Rect, measureSize types.Size) bool {
	if lineWrapRect.MaxX()+measureSize.Width > p.contentRect.MaxX() {
		return true
	}
	return false
}

// 判定：ページ
func (p *PDF) needPageBreak(lineWrapRect types.Rect, measureSize types.Size) bool {
	if lineWrapRect.MinY()+measureSize.Height > p.contentRect.MaxY() {
		return true
	}
	return false
}
