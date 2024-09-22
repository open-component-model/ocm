package runtime

import (
	"encoding/json"
)

func ToJSON(in interface{}) (json.RawMessage, error) {
	if in == nil {
		return nil, nil
	}

	var raw interface{}
	switch c := in.(type) {
	case json.RawMessage:
		return c, nil
	case []byte:
		err := DefaultYAMLEncoding.Unmarshal(c, &raw)
		if err != nil {
			return nil, err
		}
	case string:
		err := DefaultYAMLEncoding.Unmarshal([]byte(c), &raw)
		if err != nil {
			return nil, err
		}
	default:
		raw = c
	}
	return json.Marshal(raw)
}
