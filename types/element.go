package types

import "encoding/json"

type Element struct {
	Type       ElementType     `json:"type"`
	TemplateId string          `json:"template_id"`
	Attributes json.RawMessage `json:"attributes"`
}

type ElementLineBreak struct {
	Height float64 `json:"height"`
}

type ElementText struct {
	Text            string    `json:"text"`
	TextSize        int       `json:"text_size"`
	Color           Color     `json:"color"`
	Size            Size      `json:"size"`
	Origin          Origin    `json:"origin"`
	Border          Border    `json:"border"`
	BorderTop       Border    `json:"border_top"`
	BorderRight     Border    `json:"border_right"`
	BorderBottom    Border    `json:"border_bottom"`
	BorderLeft      Border    `json:"border_left"`
	BackgroundColor Color     `json:"background_color"`
	Align           Align     `json:"align"`
	Valign          Valign    `json:"valign"`
	Wrap            bool      `json:"wrap,string"`
	Inset           EdgeInset `json:"inset"`
	Layout          Layout    `json:"layout"`
}

type ElementImage struct {
	Path         string    `json:"path"`
	Size         Size      `json:"size"`
	Origin       Origin    `json:"origin"`
	Resize       bool      `json:"resize,string"`
	Resolution   uint      `json:"resolution"`
	Inset        EdgeInset `json:"inset"`
	Border       Border    `json:"border"`
	BorderTop    Border    `json:"border_top"`
	BorderRight  Border    `json:"border_right"`
	BorderBottom Border    `json:"border_bottom"`
	BorderLeft   Border    `json:"border_left"`
	Layout       Layout    `json:"layout"`
}
