package utils

import (
	"io"

	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/iotools"
)

func GetResourceData(acc ocm.AccessProvider) ([]byte, error) {
	return blobaccess.DataFromProvider(acc)
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

func GetResourceReader(acc ocm.AccessProvider) (io.ReadCloser, error) {
	return blobaccess.ReaderFromProvider(acc)
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
	return iotools.AddReaderCloser(reader, c), nil
}
