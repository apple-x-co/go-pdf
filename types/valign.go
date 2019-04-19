package types

type Valign string

func (v Valign) IsMiddle() bool {
	return v == "middle"
}
func (v Valign) IsBottom() bool {
	return v == "bottom"
}
