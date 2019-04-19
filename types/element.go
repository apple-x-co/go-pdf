package types

import "encoding/json"

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
	Resize bool    `json:"resize,string"`
}
