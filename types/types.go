package types

import "encoding/json"

type Page struct {
	Width       float64     `json:"width"`
	Height      float64     `json:"height"`
	LineHeight  float64     `json:"line_height"`
	LinerLayout LinerLayout `json:"liner_layout"`
}

type LinerLayout struct {
	Orientation   string        `json:"orientation"`
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
	Text string `json:"text"`
}

type ElementImage struct {
	Path string `json:"path"`
}
