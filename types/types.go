package types

import "encoding/json"

type LinerLayout struct {
	Orientation   string        `json:"orientation"`
	LinearLayouts []LinerLayout `json:"linear_layouts"`
	Elements      []Element     `json:"elements"`
}

type Element struct {
	Type       string          `json:"type"`
	Attributes json.RawMessage `json:"attributes"`
}

type ElementText struct {
	Text string `json:"text"`
}

type ElementImage struct {
	Path string `json:"path"`
}
