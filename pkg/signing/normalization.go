// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package signing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/open-component-model/ocm/pkg/errors"
)

type Entries []Entry

func (l *Entries) Add(key string, value interface{}) {
	*l = append(*l, NewEntry(key, value))
}

func (l Entries) String() string {
	return l.ToString("")
}
func (l Entries) ToString(gap string) string {
	ngap := gap + "  "
	s := "{"
	sep := ""
	for _, v := range l {
		s = fmt.Sprintf("%s\n%s", s, v.ToString(ngap))
		sep = "\n" + gap
	}
	s += sep + "}"
	return s
}

func toString(v interface{}, gap string) string {
	switch castIn := v.(type) {
	case Entries:
		return castIn.ToString(gap)
	case []Entry:
		return Entries(castIn).ToString(gap)
	case Entry:
		return castIn.ToString(gap)
	case []interface{}:
		ngap := gap + "  "
		s := "["
		sep := ""
		for _, v := range castIn {
			s = fmt.Sprintf("%s\n%s%s", s, ngap, toString(v, ngap))
			sep = "\n" + gap
		}
		s += sep + "]"
		return s
	case string:
		return castIn
		break
	default:
		panic(fmt.Sprintf("unknown type %T in sorting. This should not happen", v))
	}
	return ""
}

// Entry is used to keep exactly one key/value pair
type Entry struct {
	key   string
	value interface{}
}

func (e Entry) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		e.key: e.value,
	})
}

func NewEntry(key string, value interface{}) Entry {
	return Entry{key: key, value: value}
}

func (e Entry) Get() (string, interface{}) {
	return e.key, e.value
}

func (e Entry) Key() string {
	return e.key
}

func (e Entry) Value() interface{} {
	return e.value
}

func (e Entry) ToString(gap string) string {
	return fmt.Sprintf("%s%s: %s", gap, e.Key(), toString(e.Value(), gap))
}

func PrepareNormalization(v interface{}, excludes ExcludeRules) (Entries, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	raw := map[string]interface{}{}

	err = json.Unmarshal(data, &raw)
	if err != nil {
		return nil, err
	}

	return prepareStruct(raw, excludes)
}

func prepare(v interface{}, ex ExcludeRules) (interface{}, error) {
	switch e := v.(type) {
	case map[string]interface{}:
		return prepareStruct(e, ex)
	case []interface{}:
		return prepareArray(e, ex)
	default:
		return v, nil
	}
}

func prepareStruct(v map[string]interface{}, ex ExcludeRules) ([]Entry, error) {
	entries := Entries{}
	for key, value := range v {
		mapped, prop := ex.Field(key, value)
		if mapped != "" {
			nested, err := prepare(value, prop)
			if err != nil {
				return nil, errors.Wrapf(err, "field %q", key)
			}
			entries.Add(mapped, nested)
		}
	}
	// sort the entries based on the key
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Key() < entries[j].Key()
	})
	return entries, nil
}

func prepareArray(v []interface{}, ex ExcludeRules) ([]interface{}, error) {
	entries := []interface{}{}
	for index, value := range v {
		exclude, prop := ex.Element(value)
		if !exclude {
			nested, err := prepare(value, prop)
			if err != nil {
				return nil, errors.Wrapf(err, "entry %d", index)
			}
			entries = append(entries, nested)
		}
	}
	return entries, nil
}

func Normalize(v interface{}, ex ExcludeRules) ([]byte, error) {
	entries, err := PrepareNormalization(v, ex)
	if err != nil {
		return nil, err
	}
	return Marshal("", entries)
}

func Marshal(gap string, entries Entries) ([]byte, error) {
	byteBuffer := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(byteBuffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", gap)

	if err := encoder.Encode(entries); err != nil {
		return nil, err
	}

	normalizedJson := byteBuffer.Bytes()

	// encoder.Encode appends a newline that we do not want
	if normalizedJson[len(normalizedJson)-1] == 10 {
		normalizedJson = normalizedJson[:len(normalizedJson)-1]
	}
	return normalizedJson, nil
}
