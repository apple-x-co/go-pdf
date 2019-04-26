package types

type LinerLayout struct {
	Orientation  Orientation   `json:"orientation"`
	LineHeight   float64       `json:"line_height"`
	LinerLayouts []LinerLayout `json:"liner_layouts"`
	Elements     []Element     `json:"elements"`
	Layout       Layout        `json:"layout"`
}
