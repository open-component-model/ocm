package compdesc

import (
	"bytes"
)

type conversionError struct {
	error
}

func ThrowConversionError(err error) {
	panic(conversionError{err})
}

func (e conversionError) Error() string {
	return "conversion error: " + e.error.Error()
}

func CatchConversionError(errp *error) {
	if r := recover(); r != nil {
		if je, ok := r.(conversionError); ok {
			*errp = je
		} else {
			panic(r)
		}
	}
}

func Validate(desc *ComponentDescriptor) error {
	data, err := Encode(desc)
	if err != nil {
		return err
	}
	_, err = Decode(data)
	return err
}

// ElementIndex determines the index of an element in the element list
// for a given ElementMeta. If no element is found -1 is returned.
func ElementIndex(acc ElementAccessor, metaprovider ElementMetaProvider) int {
	meta := metaprovider.GetMeta()
	id := meta.GetIdentityDigest(acc)
	for i := 0; i < acc.Len(); i++ {
		if bytes.Equal(acc.Get(i).GetMeta().GetIdentityDigest(acc), id) {
			return i
		}
	}
	return -1
}
