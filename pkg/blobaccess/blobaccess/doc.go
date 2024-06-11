// Package blobaccess provides the basic set of types and supporting
// functions for using BlobAccess implementations.
// It is intended to be used by generic BlobAccess users not referring to
// dedicated implementations.
// Dedicated implementations can be accessed using the appropriate implementation
// package. This separation is required to avoid cycles for BlobAccess implementations
// using again BlobAccess objects.
// Alternatively, the parent package can be used, if there are no cycles. The
// parent package provides the basic functionality plus
// access to various implementation flavors in one package to offer a
// simplified discovery of available implementations .
package blobaccess
