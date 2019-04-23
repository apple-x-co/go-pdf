package types

type DocumentConfigure struct {
	Width         float64    `json:"width"`
	Height        float64    `json:"height"`
	TextSize      int        `json:"text_size"`
	TextColor     Color      `json:"text_color"`
	Header        PageHeader `json:"header"`
	Footer        PageFooter `json:"footer"`
	Pages         []Page     `json:"pages"`
	AutoPageBreak bool       `json:"auto_page_break,string"`
	CompressLevel int        `json:"compress_level"`
	Password      string     `json:"password"`
	TTFPath       string     `json:"-"`
	fontHeight    float64    `json:"-"`
}

func (D *DocumentConfigure) FontHeight() float64 {
	return D.fontHeight
}
func (D *DocumentConfigure) SetFontHeight(textHeight float64) {
	D.fontHeight = textHeight
}
