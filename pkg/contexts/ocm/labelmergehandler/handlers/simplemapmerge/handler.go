// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package simplemapmerge

import (
	"fmt"
	"reflect"

	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labelmergehandler"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const ALGORITHM = "simpleMapMerge"

func init() {
	cpi.RegisterLabelMergeHandler(Handler{})
}

// LabelValue is the minimal structure of label values usable with the merge algorithm.
type LabelValue map[string]interface{}

// Handler is the merge algorithm implementation for simple list value merging.
type Handler struct{}

var _ labelmergehandler.LabelMergeHandler = (*Handler)(nil)

func (h Handler) Algorithm() string {
	return ALGORITHM
}

func (h Handler) Description() string {
	return `
This handler merges simple map labels values.

It supports the following config structure:
- *<code>overwrite</code>* *string* (optional) (default <code>none</code>.

  - <code>none</code> no change possible, if entry differs the merge is rejected.
  - <code>local</code> the local value is preserved.
  - <code>inbound</code> the inbound value overwrites the local one.
`
}

func (h Handler) DecodeConfig(data []byte) (labelmergehandler.LabelMergeHandlerConfig, error) {
	var cfg Config
	err := runtime.DefaultYAMLEncoding.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (h Handler) Merge(ctx cpi.Context, local *metav1.Label, inbound *metav1.Label, cfg labelmergehandler.LabelMergeHandlerConfig) error {
	var c *Config

	if cfg == nil {
		c = &Config{}
		c.Complete(ctx)
	} else {
		var ok bool

		c, ok = cfg.(*Config)
		if !ok {
			return errors.ErrInvalid("label merge config type", fmt.Sprintf("%T", cfg))
		}
	}

	var lv LabelValue
	if err := local.GetValue(&lv); err != nil {
		return errors.Wrapf(err, "local label value is no map object")
	}

	var tv LabelValue
	if err := inbound.GetValue(&tv); err != nil {
		return errors.Wrapf(err, "inbound label value is no map object")
	}

	modified := false
	for lk, le := range lv {
		if te, ok := tv[lk]; ok {
			if !reflect.DeepEqual(le, te) {
				switch c.Overwrite {
				case MODE_NONE:
					return fmt.Errorf("target value for %q changed", lk)
				case MODE_LOCAL:
					tv[lk] = le
					modified = true
				}
			}
		} else {
			tv[lk] = le
			modified = true
		}
	}
	if modified {
		inbound.SetValue(tv)
	}
	return nil
}
