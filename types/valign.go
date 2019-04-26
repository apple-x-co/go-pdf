package types

const ValignMiddle = "middle"
const ValignBottom = "bottom"

type Valign string

func (V Valign) IsMiddle() bool {
	return V == ValignMiddle
}
func (V Valign) IsBottom() bool {
	return V == ValignBottom
}
