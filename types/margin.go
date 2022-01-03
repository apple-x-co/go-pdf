package types

type Margin struct {
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
}

func (E *Margin) Horizontal() float64 {
	return E.Right + E.Left
}

func (E *Margin) Vertical() float64 {
	return E.Top + E.Bottom
}

func (R *Rect) ApplyMargin(padding Margin) Rect {
	return Rect{
		Origin: Origin{
			X: R.Origin.X + padding.Left,
			Y: R.Origin.Y + padding.Top,
		},
		Size: Size{
			Width:  R.Size.Width - padding.Horizontal(),
			Height: R.Size.Height - padding.Vertical(),
		},
	}
}
