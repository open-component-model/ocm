package plugin

import (
	"github.com/open-component-model/ocm/api/ocm/cpi"
	"github.com/open-component-model/ocm/api/ocm/extensions/labels/routingslip/spi"
	"github.com/open-component-model/ocm/api/utils/runtime"
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
