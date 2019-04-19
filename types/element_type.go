package types

type ElementType string

func (E ElementType) IsLineBreak() bool {
	return E == "line_break"
}
func (E ElementType) IsText() bool {
	return E == "text"
}
func (E ElementType) IsImage() bool {
	return E == "image"
}
