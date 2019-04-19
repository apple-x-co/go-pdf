package types

type LinerLayout struct {
	Orientation   Orientation   `json:"orientation"`
	LineHeight    float64       `json:"line_height"`
	LinearLayouts []LinerLayout `json:"linear_layouts"`
	Elements      []Element     `json:"elements"`
}
