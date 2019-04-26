package types

const AlignLeft = "left"
const AlignCenter = "center"
const AlignRight = "right"

type Align string

func (A Align) IsLeft() bool {
	return A == AlignLeft || A == ""
}
func (A Align) IsCenter() bool {
	return A == AlignCenter
}
func (A Align) IsRight() bool {
	return A == AlignRight
}
