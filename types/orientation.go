package types

type Orientation string

func (O Orientation) IsVertical() bool {
	return O == "vertical"
}
func (O Orientation) IsHorizontal() bool {
	return O == "horizontal"
}
