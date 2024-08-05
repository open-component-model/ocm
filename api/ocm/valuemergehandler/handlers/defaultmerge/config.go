package defaultmerge

import (
	// special case to resolve dependency cycles.
	"github.com/mandelsoft/goutils/errors"

	hpi "ocm.software/ocm/api/ocm/valuemergehandler/internal"
)

type Mode string

func (m Mode) String() string {
	return string(m)
}

const (
	MODE_DEFAULT = Mode("")
	MODE_NONE    = Mode("none")
	MODE_LOCAL   = Mode("local")
	MODE_INBOUND = Mode("inbound")
)

func NewConfig(overwrite Mode) *Config {
	return &Config{
		Overwrite: overwrite,
	}
}

type Config struct {
	Overwrite Mode `json:"overwrite"`
}

func (c Config) Complete(ctx hpi.Context) error {
	switch c.Overwrite {
	case MODE_NONE, MODE_LOCAL, MODE_INBOUND:
	case "":
		// leave choice to using algorithm
	default:
		return errors.ErrInvalid("merge overwrite mode", string(c.Overwrite))
	}
	return nil
}
