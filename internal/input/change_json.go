package input

import (
	"bytes"
	"encoding/json"
)

// Preserve explicit null evidence without retaining raw input downstream.
func (c *changeJSON) UnmarshalJSON(data []byte) error {
	type plain changeJSON
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode((*plain)(c)); err != nil {
		return err
	}
	var sides map[string]json.RawMessage
	if err := json.Unmarshal(data, &sides); err != nil {
		return err
	}
	c.BeforeAbsent = bytes.Equal(bytes.TrimSpace(sides["before"]), []byte("null"))
	c.AfterAbsent = bytes.Equal(bytes.TrimSpace(sides["after"]), []byte("null"))
	return nil
}
