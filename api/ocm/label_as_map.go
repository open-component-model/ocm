package ocm

import (
	"encoding/json"

	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
)

// AsMap return an unmarshalled map representation.
func AsMap(l metav1.Labels, acc ComponentVersionAccess) map[string]interface{} {
	labels := map[string]interface{}{}
	for _, label := range l {
		var m interface{}
		switch {
		case label.Value != nil:
			json.Unmarshal(label.Value, &m)
		case label.Access != nil:
			blob, _ := GetBlobValue(&label, acc)
			json.Unmarshal(blob, &m)
		}

		labels[label.Name] = m
	}

	return labels
}
