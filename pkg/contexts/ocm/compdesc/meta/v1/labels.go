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

package v1

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/pkg/errors"
)

// Label is a label that can be set on objects.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type Label struct {
	// Name is the unique name of the label.
	Name string `json:"name"`
	// Value is the json/yaml data of the label
	Value json.RawMessage `json:"value"`

	// Version is the optional specification version of the attribute value
	// +kubebuilder:validation:Pattern:=`^v[0-9]+$`
	Version string `json:"version,omitempty"`
	// Signing describes whether the label should be included into the signature
	Signing bool `json:"signing,omitempty"`
}

// DeepCopyInto copies labels.
func (in *Label) DeepCopyInto(out *Label) {
	*out = *in
	out.Value = append(out.Value[:0:0], in.Value...)
}

var versionRegex = regexp.MustCompile("^v[0-9]+$")

func NewLabel(name string, value interface{}, opts ...interface{}) (*Label, error) {
	var data []byte
	var err error
	var ok bool

	if data, ok = value.([]byte); ok {
		var v interface{}
		err = json.Unmarshal(data, &v)
		if err != nil {
			return nil, errors.ErrInvalid("label value", string(data), name)
		}
	} else {
		data, err = json.Marshal(value)
		if err != nil {
			return nil, errors.ErrInvalid("label value", "<object>", name)
		}
	}
	l := &Label{Name: name, Value: data}
	for _, o := range opts {
		switch v := o.(type) {
		case bool:
			l.Signing = v
		case string:
			if versionRegex.MatchString(v) {
				l.Version = v
			} else {
				b, err := strconv.ParseBool(v)
				if err != nil {
					return nil, fmt.Errorf("invalid label option '%v'[%T]", v, v)
				}
				l.Signing = b
			}
		default:
			return nil, fmt.Errorf("invalid label option '%v'[%T]", v, v)
		}
	}
	return l, nil
}

// Labels describe a list of labels
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type Labels []Label

// Get returns the label with the given name.
func (l Labels) Get(name string) ([]byte, bool) {
	for _, label := range l {
		if label.Name == name {
			return label.Value, true
		}
	}
	return nil, false
}

func (l *Labels) Set(name string, value interface{}) error {
	newLabel, err := NewLabel(name, value)
	if err != nil {
		return err
	}
	for _, label := range *l {
		if label.Name == name {
			label.Value = newLabel.Value
			return nil
		}
	}
	*l = append(*l, *newLabel)
	return nil
}

func (l *Labels) Remove(name string) bool {
	for i, label := range *l {
		if label.Name == name {
			*l = append((*l)[:i], (*l)[i+1:]...)
			return true
		}
	}
	return false
}

// AsMap return an unmarshalled map representation.
func (l *Labels) AsMap() map[string]interface{} {
	labels := map[string]interface{}{}
	if l != nil {
		for _, label := range *l {
			var m interface{}
			json.Unmarshal(label.Value, &m)
			labels[label.Name] = m
		}
	}
	return labels
}

// Copy copies labels.
func (l Labels) Copy() Labels {
	if l == nil {
		return nil
	}
	n := make(Labels, len(l))
	copy(n, l)
	return n
}

// ValidateLabels validates a list of labels.
func ValidateLabels(fldPath *field.Path, labels Labels) field.ErrorList {
	allErrs := field.ErrorList{}
	labelNames := make(map[string]struct{})
	for i, label := range labels {
		labelPath := fldPath.Index(i)
		if len(label.Name) == 0 {
			allErrs = append(allErrs, field.Required(labelPath.Child("name"), "must specify a name"))
			continue
		}

		if _, ok := labelNames[label.Name]; ok {
			allErrs = append(allErrs, field.Duplicate(labelPath, "duplicate label name"))
			continue
		}
		labelNames[label.Name] = struct{}{}
	}
	return allErrs
}
