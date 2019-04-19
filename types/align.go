package types

type Align string

func (a Align) IsCenter() bool {
	return a == "center"
}
func (a Align) IsRight() bool {
	return a == "right"
}
