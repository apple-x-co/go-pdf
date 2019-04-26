package types

type LinerLayout struct {
	Orientation  Orientation   `json:"orientation"`
	LineHeight   float64       `json:"line_height"`
	LinerLayouts []LinerLayout `json:"liner_layouts"`
	Elements     []Element     `json:"elements"`
	LayoutWidth  LayoutSize    `json:"layout_width"`
	LayoutHeight LayoutSize    `json:"layout_height"`
	LayoutWeight float64       `json:"layout_weight"`
}
