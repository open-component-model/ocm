// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"io"

	"github.com/modern-go/reflect2"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
)

type AccessMethodSource interface {
	AccessMethod() (AccessMethod, error)
}

// ResourceReader gets a Reader for a given resource/source access.
// It provides a Reader handling the Close contract for the access method
// by connecting the access method's Close method to the Readers Close method .
func ResourceReader(s AccessMethodSource) (io.ReadCloser, error) {
	meth, err := s.AccessMethod()
	if err != nil {
		return nil, err
	}
	return ResourceReaderForMethod(meth)
}

// ResourceMimeReader gets a Reader for a given resource/source access.
// It provides a Reader handling the Close contract for the access method
// by connecting the access method's Close method to the Readers Close method.
// Additionally, the mime type is returned.
func ResourceMimeReader(s AccessMethodSource) (io.ReadCloser, string, error) {
	meth, err := s.AccessMethod()
	if err != nil {
		return nil, "", err
	}
	r, err := ResourceReaderForMethod(meth)
	return r, meth.MimeType(), err
}

func ResourceReaderForMethod(meth AccessMethod) (io.ReadCloser, error) {
	r, err := meth.Reader()
	if err != nil {
		meth.Close()
		return nil, err
	}
	return accessio.AddCloser(r, meth, "access method"), nil
}

// ResourceData extracts the data for a given resource/source access.
// It handles the Close contract for the access method for a singular use.
func ResourceData(s AccessMethodSource) ([]byte, error) {
	meth, err := s.AccessMethod()
	if err != nil {
		return nil, err
	}
	defer meth.Close()
	return meth.Get()
}

type WrappedAccessSpec interface {
	UnwrapAccessSpec() AccessSpec
}

func EffecticveAccessSpec(spec AccessSpec) AccessSpec {
	for {
		if reflect2.IsNil(spec) {
			return nil
		}
		if w, ok := spec.(WrappedAccessSpec); ok {
			spec = w.UnwrapAccessSpec()
		} else {
			break
		}
	}
	return spec
}

func effecticveAccessSpec(spec compdesc.AccessSpec) compdesc.AccessSpec {
	for {
		if reflect2.IsNil(spec) {
			return nil
		}
		if w, ok := spec.(WrappedAccessSpec); ok {
			spec = w.UnwrapAccessSpec()
		} else {
			break
		}
	}
	return spec
}
