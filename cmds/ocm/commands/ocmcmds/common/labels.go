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

package common

import (
	"encoding/json"
	"strings"

	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/errors"
	"sigs.k8s.io/yaml"
)

func gkind(kind ...string) string {
	for _, k := range kind {
		if k != "" {
			return k
		}
	}
	return "label"
}

func ParseLabel(a string, kind ...string) (*metav1.Label, error) {
	i := strings.Index(a, "=")
	if i < 0 {
		return nil, errors.ErrInvalid(gkind(kind...), a)
	}
	label := a[:i]

	var value interface{}
	err := yaml.Unmarshal([]byte(a[i+1:]), &value)
	if err != nil {
		return nil, errors.Wrapf(errors.ErrInvalid(gkind(kind...), a), "no yaml or json")
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, errors.Wrapf(errors.ErrInvalid(gkind(kind...), a), err.Error())
	}
	return &metav1.Label{
		Name:  strings.TrimSpace(label),
		Value: json.RawMessage(data),
	}, nil
}

func AddParsedLabel(labels metav1.Labels, a string, kind ...string) (metav1.Labels, error) {
	l, err := ParseLabel(a, kind...)
	if err != nil {
		return nil, err
	}
	for _, c := range labels {
		if c.Name == l.Name {
			return nil, errors.Newf("duplicate %s %q", gkind(kind...), l.Name)
		}
	}
	return append(labels, *l), nil
}

func ParseLabels(labels []string, kind ...string) (metav1.Labels, error) {
	var err error
	result := metav1.Labels{}
	for _, l := range labels {
		result, err = AddParsedLabel(result, l, kind...)
		if err != nil {
			return nil, err
		}
	}
	return result, err
}

func SetParsedLabel(labels metav1.Labels, a string, kind ...string) (metav1.Labels, error) {
	l, err := ParseLabel(a, kind...)
	if err != nil {
		return nil, err
	}
	for i, c := range labels {
		if c.Name == l.Name {
			labels[i].Value = l.Value
			return labels, nil
		}
	}
	return append(labels, *l), nil
}
