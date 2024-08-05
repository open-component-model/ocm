// Package blobaccess provides various flavors of BlobAccess implementations.
// It is used to access blobs in a uniform way. Additionally, it provides
// the basic types provided by the sub package with the same name.
// This package just provides the basic types for
// BlobAccess implementations. It is provided for generic BlobAccess users not
// referring to dedicated implementation. This separation is required to avoid cycles
// for blobaccess users which are also used again to implement some of the BlobAccess
// variants.
// Alternatively, the sub package and the various implementation packages
// can be used in combination.
package blobaccess
