package types

type Page struct {
	LinerLayout LinerLayout `json:"liner_layout"`
	PageHeader  Header      `json:"page_header"`
	PageFooter  Footer      `json:"page_footer"`
	FixedTitle  Header      `json:"fixed_title"`
}
