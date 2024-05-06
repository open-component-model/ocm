package output

import (
	"github.com/open-component-model/ocm/pkg/errors"
)

type SingleElementOutput struct {
	Elem interface{}
}

var _ Output = &SingleElementOutput{}

func NewSingleElementOutput() *SingleElementOutput {
	return &SingleElementOutput{}
}

func (this *SingleElementOutput) Add(e interface{}) error {
	if this.Elem == nil {
		this.Elem = e
		return nil
	}
	return errors.Newf("only one element can be selected, but multiple elements selected/found")
}

func (this *SingleElementOutput) Close() error {
	return nil
}

func (this *SingleElementOutput) Out() error {
	return nil
}
