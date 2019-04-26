package types

const ValignTop = "top"
const ValignMiddle = "middle"
const ValignBottom = "bottom"

type Valign string

func (V Valign) IsTop() bool {
	return V == ValignTop || V == ""
}
func (V Valign) IsMiddle() bool {
	return V == ValignMiddle
}
func (V Valign) IsBottom() bool {
	return V == ValignBottom
}
