package types

import "encoding/json"

type PDF struct {
	Width      float64 `json:"width"`
	Height     float64 `json:"height"`
	LineHeight float64 `json:"line_height"`
	TextColor  Color   `json:"text_color"`
	Pages      []Page  `json:"pages"`
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

func (linerLayout *LinerLayout) IsVertical() bool {
	return linerLayout.Orientation == "vertical"
}
func (linerLayout *LinerLayout) IsHorizontal() bool {
	return linerLayout.Orientation == "horizontal"
}

type Element struct {
	Type       string          `json:"type"`
	Attributes json.RawMessage `json:"attributes"`
}

func (element *Element) IsLineBreak() bool {
	return element.Type == "line_break"
}
func (element *Element) IsText() bool {
	return element.Type == "text"
}
func (element *Element) IsImage() bool {
	return element.Type == "image"
}

type ElementLineBreak struct {
	Height float64 `json:"height"`
}

type ElementText struct {
	Text  string `json:"text"`
	Color Color  `json:"color"`
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
