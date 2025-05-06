package mock

import (
	"github.com/mandelsoft/goutils/general"

	"ocm.software/ocm/api/ocm/tools/toi/install"
)

type Driver struct {
	handler func(*install.Operation) (*install.OperationResult, error)
}

var _ install.Driver = (*Driver)(nil)

func New(handler ...func(*install.Operation) (*install.OperationResult, error)) install.Driver {
	return &Driver{general.Optional(handler...)}
}

func (d *Driver) SetConfig(props map[string]string) error {
	return nil
}

func (d *Driver) Exec(op *install.Operation) (*install.OperationResult, error) {
	if d.handler != nil {
		return d.handler(op)
	}
	return &install.OperationResult{}, nil
}
