package compdesc

import (
	"encoding/json"

	"github.com/mandelsoft/goutils/generics"
)

type GenericComponentDescriptor ComponentDescriptor

var (
	_ json.Marshaler   = (*GenericComponentDescriptor)(nil)
	_ json.Unmarshaler = (*GenericComponentDescriptor)(nil)
)

func (g GenericComponentDescriptor) MarshalJSON() ([]byte, error) {
	return Encode(generics.Pointer(ComponentDescriptor(g)), DefaultJSONCodec)
}

func (g *GenericComponentDescriptor) UnmarshalJSON(bytes []byte) error {
	cd, err := Decode(bytes, DefaultJSONCodec)
	if err != nil {
		return err
	}
	*g = *((*GenericComponentDescriptor)(cd))
	return nil
}

func (g *GenericComponentDescriptor) Descriptor() *ComponentDescriptor {
	return (*ComponentDescriptor)(g)
}
