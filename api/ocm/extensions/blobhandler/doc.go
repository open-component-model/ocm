// Package blobhandler contains blobhandlers for handling local blobs
// for dedicated repository types. It is structured into sub packaged
// for all context types, for example the context type oci for
// implementations of the oci go binding interface
// In those sub packages there a handler packages for dedicated repository
// implementations for this type, for example the oci registry implementation
// for the connect type oci.
package blobhandler
