package types

type ContentMargin struct {
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
}

func (E *ContentMargin) Horizontal() float64 {
	return E.Right + E.Left
}

func (E *ContentMargin) Vertical() float64 {
	return E.Top + E.Bottom
}

func (R *Rect) ApplyContentMargin(contentMargin ContentMargin) Rect {
	return Rect{
		Origin: Origin{
			X: R.Origin.X + contentMargin.Left - contentMargin.Right,
			Y: R.Origin.Y + contentMargin.Top - contentMargin.Bottom,
		},
		Size: Size{
			Width:  R.Size.Width,
			Height: R.Size.Height,
		},
	}
}
