// Package blobaccess provides the basic set of types and supporting
// functions for using BlobAccess implementations.
// It is intended to be used by generic BlobAccess users not referring to
// dedicated implementations.
// Dedicated implementations can be accessed using the appropriate implementation
// package. THis separation is required to avoid cycle for BlobAccess implementations
// using again BlobAccess consumers.
// Alternatively, the parent package can be used, if there are no cycles. The
// parent package provide thes ebasic functionality plus
// access the to various implementation flavors in one package.
package blobaccess
