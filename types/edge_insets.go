package types

type EdgeInset struct {
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
}

func (E *EdgeInset) Horizontal() float64 {
	return E.Right + E.Left
}

func (E *EdgeInset) Vertical() float64 {
	return E.Top + E.Bottom
}

func (R *Rect) Inset(inset EdgeInset) Rect {
	return Rect{
		Origin: Origin{
			X: R.Origin.X + inset.Left,
			Y: R.Origin.Y + inset.Top,
		},
		Size: Size{
			Width:  R.Size.Width - inset.Horizontal(),
			Height: R.Size.Height - inset.Vertical(),
		},
	}
}
