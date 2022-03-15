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

package ocmcmds

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/gardener/ocm/pkg/errors"
	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"
	"sigs.k8s.io/yaml"
)

func ParseLabel(a string) (*metav1.Label, error) {
	i := strings.Index(a, "=")
	if i < 0 {
		return nil, errors.ErrInvalid("label", a)
	}
	label := a[:i]

	var value interface{}
	err := yaml.Unmarshal([]byte(a[i+1:]), &value)
	if err != nil {
		return nil, errors.Wrapf(errors.ErrInvalid("label", a), "no yaml or json")
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, errors.Wrapf(errors.ErrInvalid("label", a), err.Error())
	}
	return &metav1.Label{
		Name:  strings.TrimSpace(label),
		Value: json.RawMessage(data),
	}, nil
}

func AddParsedLabel(labels metav1.Labels, a string) (metav1.Labels, error) {
	l, err := ParseLabel(a)
	if err != nil {
		return nil, err
	}
	for _, c := range labels {
		if c.Name == l.Name {
			return nil, errors.Newf("duplicate label %q", l.Name)
		}
	}
	return append(labels, *l), nil
}

func SetParsedLabel(labels metav1.Labels, a string) (metav1.Labels, error) {
	l, err := ParseLabel(a)
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

func ReadEnv(path string) (map[string]string, error) {
	var (
		part   []byte
		prefix bool
	)

	result := map[string]string{}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)
	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	for {
		if part, prefix, err = reader.ReadLine(); err != nil {
			break
		}
		buffer.Write(part)
		if !prefix {
			line := strings.TrimSpace(buffer.String())
			if line != "" && !strings.HasPrefix(line, "#") {
				i := strings.Index(line, "=")
				if i <= 0 {
					return nil, errors.Newf("invalid variable syntax %q", line)
				}
				result[strings.TrimSpace(line[:i])] = strings.TrimSpace(line[i+1:])
			}
			buffer.Reset()
		}
	}
	if err == io.EOF {
		err = nil
	}
	return result, err
}
