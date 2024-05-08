package template

import (
	"encoding/json"

	"github.com/imdario/mergo"
	"sigs.k8s.io/yaml"
)

type MergeTemplater struct{}

func NewMerge() Templater {
	return &MergeTemplater{}
}

func (g MergeTemplater) Process(data string, values Values) (string, error) {
	var dst map[string]interface{}

	var err error
	wasjson := false
	if err = json.Unmarshal([]byte(data), &dst); err == nil {
		wasjson = true
	} else {
		err = yaml.Unmarshal([]byte(data), &dst)
	}
	if err != nil {
		return "", err
	}
	err = mergo.Merge(&dst, map[string]interface{}(values))
	if err != nil {
		return "", err
	}

	var r []byte
	if wasjson {
		r, err = json.Marshal(dst)
	} else {
		r, err = yaml.Marshal(dst)
	}
	if err != nil {
		return "", err
	}
	return string(r), nil
}
