package clisupport

import (
	"encoding/json"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/utils"
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

	data, err := utils.ResolveData(a[i+1:], fs)
	if err != nil {
		return nil, err
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
