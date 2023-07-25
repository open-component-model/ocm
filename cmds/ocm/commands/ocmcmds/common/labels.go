// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"encoding/json"
	"strings"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"sigs.k8s.io/yaml"

	metav1 "github.com/open-component-model/ocm/v2/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/v2/pkg/errors"
)

func gkind(kind ...string) string {
	for _, k := range kind {
		if k != "" {
			return k
		}
	}
	return "label"
}

func ParseLabel(fs vfs.FileSystem, a string, kind ...string) (*metav1.Label, error) {
	var err error

	if fs == nil {
		fs = osfs.New()
	}
	i := strings.Index(a, "=")
	if i < 0 {
		return nil, errors.ErrInvalid(gkind(kind...), a)
	}
	label := a[:i]

	var data []byte
	if strings.HasPrefix(a[i+1:], "@") {
		data, err = vfs.ReadFile(fs, a[i+2:])
		if err != nil {
			return nil, errors.Wrapf(err, "cannot read file %q", a[i+2:])
		}
	} else {
		data = []byte(a[i+1:])
	}

	var value interface{}
	err = yaml.Unmarshal(data, &value)
	if err != nil {
		return nil, errors.Wrapf(errors.ErrInvalid(gkind(kind...), a), "no yaml or json")
	}
	data, err = json.Marshal(value)
	if err != nil {
		return nil, errors.Wrapf(errors.ErrInvalid(gkind(kind...), a), err.Error())
	}
	return &metav1.Label{
		Name:  strings.TrimSpace(label),
		Value: json.RawMessage(data),
	}, nil
}

func AddParsedLabel(fs vfs.FileSystem, labels metav1.Labels, a string, kind ...string) (metav1.Labels, error) {
	l, err := ParseLabel(fs, a, kind...)
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

func ParseLabels(fs vfs.FileSystem, labels []string, kind ...string) (metav1.Labels, error) {
	var err error
	result := metav1.Labels{}
	for _, l := range labels {
		result, err = AddParsedLabel(fs, result, l, kind...)
		if err != nil {
			return nil, err
		}
	}
	return result, err
}

func SetParsedLabel(fs vfs.FileSystem, labels metav1.Labels, a string, kind ...string) (metav1.Labels, error) {
	l, err := ParseLabel(fs, a, kind...)
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
