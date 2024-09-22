package transferhandler

import (
	"github.com/mandelsoft/goutils/errors"
)

//////////////////////////////////////////////////////////////////////////////

type ConfigOption interface {
	SetConfig([]byte)
	GetConfig() []byte
}

type configOption struct {
	config []byte
}

func (o *configOption) ApplyTransferOption(to TransferOptions) error {
	if eff, ok := to.(ConfigOption); ok {
		eff.SetConfig(o.config)
		return nil
	} else {
		return errors.ErrNotSupported(KIND_TRANSFEROPTION, "config")
	}
}

// WithConfig configures a handler specific configuration.
// It is accepted by all handler featuring such a config possibility.
func WithConfig(config []byte) TransferOption {
	return &configOption{
		config: config,
	}
}
