// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip/spi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type Entry struct {
	runtime.UnstructuredVersionedTypedObject `json:",inline"`
	handler                                  *PluginHandler
}

var _ spi.Entry = &Entry{}

func (s *Entry) Describe(ctx cpi.Context) string {
	return s.handler.Describe(s, ctx)
}

func (s *Entry) Validate(ctx spi.Context) error {
	_, err := s.handler.Validate(s)
	return err
}

func (s *Entry) Handler() *PluginHandler {
	return s.handler
}
