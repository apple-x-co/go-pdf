package types

const AlignCenter = "center"
const AlignRight = "right"

type Align string

func (A Align) IsCenter() bool {
	return A == AlignCenter
}
func (A Align) IsRight() bool {
	return A == AlignRight
}
