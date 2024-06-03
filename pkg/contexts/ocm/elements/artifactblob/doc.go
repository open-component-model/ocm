// Package artifactblob hosts packages for
// ResourceAccess and SourceAccess builders used to add
// resources and sources as local blobs to a component version.
// There is a generic access just requiring blobaccess.BlobAccess objects
// and various specialized builders for various technologied
// and source locations.
//
// For example, the package helm provides a builder, which can be used
// to feed in helm charts from helm chart repositories or from the
// local filesystem.
package artifactblob
