package types

type Rect struct {
	Origin Origin
	Size   Size
}

func (R *Rect) Merge(anotherRect Rect) Rect {
	var newRect = Rect{Origin: Origin{X: R.Origin.X, Y: R.Origin.Y}, Size: Size{Width: R.Size.Width, Height: R.Size.Height}}

	if R.MaxX() < anotherRect.MaxX() {
		newRect.Size.Width = anotherRect.MaxX() - R.Origin.X
	}
	if R.MaxY() < anotherRect.MaxY() {
		newRect.Size.Height = anotherRect.MaxY() - R.Origin.Y
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
