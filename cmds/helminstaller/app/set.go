package app

import (
	"fmt"
	"strings"
)

// TODO: support better path expressions

func Set(values map[string]interface{}, path string, value interface{}) error {
	fields := strings.Split(path, ".")
	i := 0
	for ; i < len(fields)-1; i++ {
		f := strings.TrimSpace(fields[i])
		v, ok := values[f]
		if !ok {
			v = map[string]interface{}{}
			values[f] = v
		} else {
			if _, ok := v.(map[string]interface{}); !ok {
				return fmt.Errorf("invalid field path %s", strings.Join(fields[:i+1], "."))
			}
		}
		values = v.(map[string]interface{})
	}
	values[fields[len(fields)-1]] = value
	return nil
}
