package types

const OrientationHorizontal = "horizontal"
const OrientationVertical = "vertical"

type Orientation string

func (O Orientation) IsHorizontal() bool {
	return O == OrientationHorizontal || O == ""
}
func (O Orientation) IsVertical() bool {
	return O == OrientationVertical
}
