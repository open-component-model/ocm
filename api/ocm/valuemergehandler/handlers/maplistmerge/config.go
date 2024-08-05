package maplistmerge

import (
	"ocm.software/ocm/api/ocm/valuemergehandler/handlers/defaultmerge"
	"ocm.software/ocm/api/ocm/valuemergehandler/hpi"
	"ocm.software/ocm/api/utils"
)

type Mode = defaultmerge.Mode

const (
	MODE_DEFAULT = defaultmerge.MODE_DEFAULT
	MODE_NONE    = defaultmerge.MODE_NONE
	MODE_LOCAL   = defaultmerge.MODE_LOCAL
	MODE_INBOUND = defaultmerge.MODE_INBOUND
)

func NewConfig(field string, overwrite Mode, entries ...*hpi.Specification) *Config {
	return &Config{
		KeyField: field,
		Config:   *defaultmerge.NewConfig(overwrite),
		Entries:  utils.Optional(entries...),
	}
}

type Config struct {
	defaultmerge.Config
	KeyField string             `json:"keyField"`
	Entries  *hpi.Specification `json:"entries,omitempty"`
}

func (c *Config) Complete(ctx hpi.Context) error {
	err := c.Config.Complete(ctx)
	if err != nil {
		return err
	}
	if c.KeyField == "" {
		c.KeyField = "name"
	}
	return nil
}
