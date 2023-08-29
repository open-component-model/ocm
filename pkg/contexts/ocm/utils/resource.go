// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"io"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
)

func GetResourceData(acc ocm.ResourceAccess) ([]byte, error) {
	m, err := acc.AccessMethod()
	if err != nil {
		return nil, err
	}
	defer m.Close()
	return m.Get()
}

func GetResourceDataForPath(cv ocm.ComponentVersionAccess, id metav1.Identity, path []metav1.Identity, resolvers ...ocm.ComponentVersionResolver) ([]byte, error) {
	return GetResourceDataForRef(cv, metav1.NewNestedResourceRef(id, path), resolvers...)
}

func GetResourceDataForRef(cv ocm.ComponentVersionAccess, ref metav1.ResourceReference, resolvers ...ocm.ComponentVersionResolver) ([]byte, error) {
	var res ocm.ComponentVersionResolver
	if len(resolvers) > 0 {
		res = ocm.NewCompoundResolver(resolvers...)
	}
	a, c, err := ResolveResourceReference(cv, ref, res)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	return GetResourceData(a)
}

func GetResourceReader(acc ocm.ResourceAccess) (io.ReadCloser, error) {
	m, err := acc.AccessMethod()
	if err != nil {
		return nil, err
	}
	return m.Reader()
}

func GetResourceReaderForPath(cv ocm.ComponentVersionAccess, id metav1.Identity, path []metav1.Identity, resolvers ...ocm.ComponentVersionResolver) (io.ReadCloser, error) {
	return GetResourceReaderForRef(cv, metav1.NewNestedResourceRef(id, path), resolvers...)
}

func GetResourceReaderForRef(cv ocm.ComponentVersionAccess, ref metav1.ResourceReference, resolvers ...ocm.ComponentVersionResolver) (io.ReadCloser, error) {
	var res ocm.ComponentVersionResolver
	if len(resolvers) > 0 {
		res = ocm.NewCompoundResolver(resolvers...)
	}
	a, c, err := ResolveResourceReference(cv, ref, res)
	if err != nil {
		return nil, err
	}

	reader, err := GetResourceReader(a)
	if err != nil {
		c.Close()
		return nil, err
	}
	return accessio.AddCloser(reader, c), nil
}
