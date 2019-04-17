package types

import "encoding/json"

type PDF struct {
	Width         float64 `json:"width"`
	Height        float64 `json:"height"`
	LineHeight    float64 `json:"line_height"`
	TextSize      int     `json:"text_size"`
	TextColor     Color   `json:"text_color"`
	Pages         []Page  `json:"pages"`
	AutoPageBreak bool    `json:"auto_page_break,string"`
	textHeight    float64 `json:"text_cap_height"`
}

func (P *PDF) TextHeight() float64 {
	return P.textHeight
}
func (P *PDF) SetTextHeight(textHeight float64) {
	P.textHeight = textHeight
}

type Page struct {
	LinerLayout LinerLayout `json:"liner_layout"`
}

type LinerLayout struct {
	Orientation   string        `json:"orientation"`
	LineHeight    float64       `json:"line_height"`
	LinearLayouts []LinerLayout `json:"linear_layouts"`
	Elements      []Element     `json:"elements"`
}

func (L *LinerLayout) IsVertical() bool {
	return L.Orientation == "vertical"
}
func (L *LinerLayout) IsHorizontal() bool {
	return L.Orientation == "horizontal"
}

type Element struct {
	Type       string          `json:"type"`
	Attributes json.RawMessage `json:"attributes"`
}

func (E *Element) IsLineBreak() bool {
	return E.Type == "line_break"
}
func (E *Element) IsText() bool {
	return E.Type == "text"
}
func (E *Element) IsImage() bool {
	return E.Type == "image"
}

type ElementLineBreak struct {
	Height float64 `json:"height"`
}

type ElementText struct {
	Text            string  `json:"text"`
	Color           Color   `json:"color"`
	Width           float64 `json:"width"`
	Height          float64 `json:"height"`
	Border          Border  `json:"border"`
	BorderTop       Border  `json:"border_top"`
	BorderRight     Border  `json:"border_right"`
	BorderBottom    Border  `json:"border_bottom"`
	BorderLeft      Border  `json:"border_left"`
	BackgroundColor Color   `json:"background_color"`
	Align           string  `json:"align"`
	Valign          string  `json:"valign"`
}

func (ET *ElementText) IsAlignCenter() bool {
	return ET.Align == "center"
}
func (ET *ElementText) IsAlignRight() bool {
	return ET.Align == "right"
}
func (ET *ElementText) IsValignMiddle() bool {
	return ET.Valign == "middle"
}
func (ET *ElementText) IsValignBottom() bool {
	return ET.Valign == "bottom"
}

type ElementImage struct {
	Path   string  `json:"path"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
}

type Color struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

type Border struct {
	Width float64 `json:"width"`
	Color Color   `json:"color"`
}
