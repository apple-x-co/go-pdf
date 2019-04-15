package types

type LinerLayout struct {
	Orientation   string        `json:"orientation"`
	LinearLayouts []LinerLayout `json:"linear_layouts"`
	Elements      []interface{} `json:"elements"`
	//Elements []Element `json:"elements"`
}

//type Element struct {
//	Type string `json:"type"`
//	Attributes json.RawMessage `json:"attributes"`
//}
