package types

type Size struct {
	Width  float64
	Height float64
}

func (S *Size) IsZero() bool {
	return S.Width == 0 && S.Height == 0
}
