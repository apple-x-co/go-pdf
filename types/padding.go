package types

type Padding struct {
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
}

func (E *Padding) Horizontal() float64 {
	return E.Right + E.Left
}

func (E *Padding) Vertical() float64 {
	return E.Top + E.Bottom
}

func (R *Rect) ApplyPadding(padding Padding) Rect {
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
