package types

import "encoding/json"

type Element struct {
	Type       ElementType     `json:"type"`
	Attributes json.RawMessage `json:"attributes"`
}

type ElementLineBreak struct {
	Height float64 `json:"height"`
}

type ElementText struct {
	Text            string `json:"text"`
	Color           Color  `json:"color"`
	Size            Size   `json:"size"`
	Origin          Origin `json:"origin"`
	Border          Border `json:"border"`
	BorderTop       Border `json:"border_top"`
	BorderRight     Border `json:"border_right"`
	BorderBottom    Border `json:"border_bottom"`
	BorderLeft      Border `json:"border_left"`
	BackgroundColor Color  `json:"background_color"`
	Align           Align  `json:"align"`
	Valign          Valign `json:"valign"`
}

type ElementImage struct {
	Path   string `json:"path"`
	Size   Size   `json:"size"`
	Origin Origin `json:"origin"`
	Resize bool   `json:"resize,string"`
}
