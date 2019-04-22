package types

type Rect struct {
	Origin Origin
	Size   Size
}

func (R *Rect) Merge(anotherRect Rect) Rect {
	var newRect = Rect{Origin: Origin{X: R.Origin.X, Y: R.Origin.Y}, Size: Size{Width: R.Size.Width, Height: R.Size.Height}}

	if R.Origin.X+R.Size.Width < anotherRect.Origin.X+anotherRect.Size.Width {
		newRect.Size.Width = anotherRect.Origin.X + anotherRect.Size.Width - R.Origin.X
	}
	if R.Origin.Y+R.Size.Height < anotherRect.Origin.Y+anotherRect.Size.Height {
		newRect.Size.Height = anotherRect.Origin.Y + anotherRect.Size.Height - R.Origin.Y
	}

	return newRect
}

func (R *Rect) Width() float64 {
	return R.Size.Width
}

func (R *Rect) Height() float64 {
	return R.Size.Height
}

func (R *Rect) MinX() float64 {
	return R.Origin.X
}

func (R *Rect) MaxX() float64 {
	return R.MinX() + R.Width()
}

func (R *Rect) MinY() float64 {
	return R.Origin.Y
}

func (R *Rect) MaxY() float64 {
	return R.MinY() + R.Height()
}
