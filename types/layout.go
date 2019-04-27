package types

type Layout struct {
	Width  LayoutConstant `json:"width"`
	Height LayoutConstant `json:"height"`
	Ratio  float64        `json:"ratio"`
}
