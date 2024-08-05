package mock

import (
	"ocm.software/ocm/api/ocm/tools/toi/install"
	"ocm.software/ocm/api/utils"
)

type Driver struct {
	handler func(*install.Operation) (*install.OperationResult, error)
}

var _ install.Driver = (*Driver)(nil)

func New(handler ...func(*install.Operation) (*install.OperationResult, error)) install.Driver {
	return &Driver{utils.Optional(handler...)}
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
