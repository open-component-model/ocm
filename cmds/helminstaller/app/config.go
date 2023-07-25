// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"

	v1 "github.com/open-component-model/ocm/v2/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/utils/localize"
	"github.com/open-component-model/ocm/v2/pkg/errors"
)

type Config struct {
	Chart           v1.ResourceReference            `json:"chart,omitempty"`
	SubCharts       map[string]v1.ResourceReference `json:"subCharts,omitempty"`
	Release         string                          `json:"release,omitempty"`
	Namespace       string                          `json:"namespace,omitempty"`
	CreateNamespace bool                            `json:"createNamespace,omitempty"`
	ImageMapping    []ImageMapping                  `json:"imageMapping"`
	Values          json.RawMessage                 `json:"values,omitempty"`
	KubeConfigName  string                          `json:"kubeConfigName,omitempty"`
}

type ImageMapping = localize.ImageMapping

func (c *Config) GetValues() (map[string]interface{}, error) {
	if len(c.Values) == 0 {
		return map[string]interface{}{}, nil
	}
	var result map[string]interface{}
	err := json.Unmarshal(c.Values, &result)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal values from bootstrap config")
	}
	return result, nil
}
