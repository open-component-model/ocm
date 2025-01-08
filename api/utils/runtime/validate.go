package runtime

import (
	"github.com/mandelsoft/goutils/errors"
)

type Validater interface { // codespell:ignore
	Validate() error
}

func Validate(o interface{}) error {
	if t, ok := o.(TypedObject); ok {
		if t.GetType() == "" {
			return errors.New("type missing")
		}
	}
	if v, ok := o.(Validater); ok { // codespell:ignore
		return v.Validate()
	}
	return nil
}
