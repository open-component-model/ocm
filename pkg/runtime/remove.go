// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"encoding/json"
)

// Remove removes additional top level attributes from a generic
// object specification based on an object covering the additional
// attributes.
func Remove(data UnstructuredMap, obj interface{}) error {
	var m UnstructuredMap
	bytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return err
	}
	for k := range m {
		delete(data, k)
	}
	return nil
}

// RemoveFromData removes additional top level attributes from an
// object specification based on an object covering the additional
// attributes.
func RemoveFromData(data []byte, obj interface{}, unmarshaler Unmarshaler) ([]byte, error) {
	var m UnstructuredMap
	bytes, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return nil, err
	}

	if unmarshaler == nil {
		unmarshaler = DefaultJSONEncoding
	}

	var d UnstructuredMap
	err = unmarshaler.Unmarshal(data, &d)
	if err != nil {
		return nil, err
	}

	for k := range m {
		delete(d, k)
	}
	return json.Marshal(d)
}

// UnmarshalAdditional unmarshalls an object serialization by a given
// unmarshaler by separating independent top-level property sets defined
// by a main object and several addenda given by a list of deserializable objects.
// The main object then gets all the properties not covered by the addenda,
// while the additional aspects are fed into the given objects.
func UnmarshalAdditional(data []byte, result Unstructured, unmarshaler Unmarshaler, additional ...interface{}) error {
	if unmarshaler == nil {
		unmarshaler = DefaultJSONEncoding
	}

	err := unmarshaler.Unmarshal(data, result)
	if err != nil {
		return err
	}

	for _, a := range additional {
		err := unmarshaler.Unmarshal(data, a)
		if err != nil {
			return err
		}
		err = Remove(result.GetObject(), a)
		if err != nil {
			return err
		}
	}
	return nil
}
