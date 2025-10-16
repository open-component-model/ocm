package stdopts

import (
	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/datacontext"
)

type DataContextOptionBag interface {
	SetDataContext(ctx datacontext.Context)
}

var _ DataContextOptionBag = (*DataContextOption)(nil)

type DataContextOption struct {
	Value datacontext.Context
}

func (d *DataContextOption) SetDataContext(ctx datacontext.Context) {
	d.Value = ctx
}

////////////////////////////////////////////////////////////////////////////////

type PathFileSystemOptionBag interface {
	SetPathFileSystem(v vfs.FileSystem)
}

var _ PathFileSystemOptionBag = (*PathFileSystem)(nil)

type PathFileSystem struct {
	Value vfs.FileSystem
}

func (d *PathFileSystem) SetPathFileSystem(v vfs.FileSystem) {
	d.Value = v
}

////////////////////////////////////////////////////////////////////////////////

type LoggingContextOptionBag interface {
	SetLoggingContext(ctx logging.ContextProvider)
}

var _ LoggingContextOptionBag = (*LoggingContext)(nil)

type LoggingContext struct {
	Value logging.Context
}

func (d *LoggingContext) SetLoggingContext(ctx logging.ContextProvider) {
	d.Value = ctx.LoggingContext()
}

////////////////////////////////////////////////////////////////////////////////

type CredentialContextOptionBag interface {
	SetCredentialContext(ctx credentials.ContextProvider)
}

var _ CredentialContextOptionBag = (*CredentialContext)(nil)

type CredentialContext struct {
	Value credentials.Context
}

func (d *CredentialContext) SetCredentialContext(ctx credentials.ContextProvider) {
	d.Value = ctx.CredentialsContext()
}

////////////////////////////////////////////////////////////////////////////////

type CredentialsOptionBag interface {
	SetCredentials(v cpi.Credentials)
}

var _ CredentialsOptionBag = (*Credentials)(nil)

type Credentials struct {
	Value credentials.Credentials
}

func (d *Credentials) SetCredentials(v credentials.Credentials) {
	d.Value = v
}

////////////////////////////////////////////////////////////////////////////////

type CachingContextOptionBag interface {
	SetCachingContext(v datacontext.ContextProvider)
}

var _ CachingContextOptionBag = (*CachingContext)(nil)

type CachingContext struct {
	Value datacontext.Context
}

func (d *CachingContext) SetCachingContext(ctx datacontext.ContextProvider) {
	d.Value = ctx.AttributesContext()
}

////////////////////////////////////////////////////////////////////////////////

type CachingPathOptionBag interface {
	SetCachingPath(v string)
}

var _ CachingPathOptionBag = (*CachingPath)(nil)

type CachingPath struct {
	Value string
}

func (d *CachingPath) SetCachingPath(v string) {
	d.Value = v
}

////////////////////////////////////////////////////////////////////////////////

type CachingFileSystemOptionBag interface {
	SetCachingFileSystem(v vfs.FileSystem)
}

var _ CachingFileSystemOptionBag = (*CachingFileSystem)(nil)

type CachingFileSystem struct {
	Value vfs.FileSystem
}

func (d *CachingFileSystem) SetCachingFileSystem(v vfs.FileSystem) {
	d.Value = v
}
