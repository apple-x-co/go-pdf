package types

type Layout struct {
	Width  LayoutConstant `json:"width"`
	Height LayoutConstant `json:"height"`
	Weight float64        `json:"weight"`
}
