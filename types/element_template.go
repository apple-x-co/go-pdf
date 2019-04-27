package types

import "encoding/json"

type ElementTemplate struct {
	Type       ElementType     `json:"type"`
	Id         string          `json:"id"`
	Attributes json.RawMessage `json:"attributes"`
}
