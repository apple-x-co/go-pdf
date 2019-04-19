package types

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
