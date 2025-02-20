package standard

import (
	"slices"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/maputils"
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/goutils/sliceutils"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/utils/runtime"
)

func init() {
	transferhandler.RegisterHandler(1000, &TransferOptionsCreator{})
}

type Options struct {
	retries           *int
	recursive         *bool
	resourcesByValue  *bool
	localByValue      *bool
	sourcesByValue    *bool
	keepGlobalAccess  *bool
	stopOnExisting    *bool
	enforceTransport  *bool
	overwrite         *bool
	skipUpdate        *bool
	omitAccessTypes   set.Set[string]
	omitArtifactTypes set.Set[string]
	resolver          ocm.ComponentVersionResolver
}

var (
	_ transferhandler.TransferOption = (*Options)(nil)

	_ RetryOption                 = (*Options)(nil)
	_ ResourcesByValueOption      = (*Options)(nil)
	_ LocalResourcesByValueOption = (*Options)(nil)
	_ EnforceTransportOption      = (*Options)(nil)
	_ OverwriteOption             = (*Options)(nil)
	_ SkipUpdateOption            = (*Options)(nil)
	_ SourcesByValueOption        = (*Options)(nil)
	_ RecursiveOption             = (*Options)(nil)
	_ ResolverOption              = (*Options)(nil)
	_ KeepGlobalAccessOption      = (*Options)(nil)
	_ OmitAccessTypesOption       = (*Options)(nil)
	_ OmitArtifactTypesOption     = (*Options)(nil)
)

type TransferOptionsCreator = transferhandler.SpecializedOptionsCreator[*Options, Options]

func (o *Options) NewOptions() transferhandler.TransferHandlerOptions {
	return &Options{}
}

func (o *Options) NewTransferHandler() (transferhandler.TransferHandler, error) {
	return New(o)
}

func (o *Options) ApplyTransferOption(target transferhandler.TransferOptions) error {
	if o.retries != nil {
		if opts, ok := target.(RetryOption); ok {
			opts.SetRetries(*o.retries)
		}
	}
	if o.recursive != nil {
		if opts, ok := target.(RecursiveOption); ok {
			opts.SetRecursive(*o.recursive)
		}
	}
	if o.skipUpdate != nil {
		if opts, ok := target.(SkipUpdateOption); ok {
			opts.SetSkipUpdate(*o.skipUpdate)
		}
	}
	if o.resourcesByValue != nil {
		if opts, ok := target.(ResourcesByValueOption); ok {
			opts.SetResourcesByValue(*o.resourcesByValue)
		}
	}
	if o.localByValue != nil {
		if opts, ok := target.(LocalResourcesByValueOption); ok {
			opts.SetLocalResourcesByValue(*o.localByValue)
		}
	}
	if o.sourcesByValue != nil {
		if opts, ok := target.(SourcesByValueOption); ok {
			opts.SetSourcesByValue(*o.sourcesByValue)
		}
	}
	if o.keepGlobalAccess != nil {
		if opts, ok := target.(KeepGlobalAccessOption); ok {
			opts.SetKeepGlobalAccess(*o.keepGlobalAccess)
		}
	}
	if o.stopOnExisting != nil {
		if opts, ok := target.(StopOnExistingVersionOption); ok {
			opts.SetStopOnExistingVersion(*o.stopOnExisting)
		}
	}
	if o.enforceTransport != nil {
		if opts, ok := target.(EnforceTransportOption); ok {
			opts.SetEnforceTransport(*o.enforceTransport)
		}
	}
	if o.overwrite != nil {
		if opts, ok := target.(OverwriteOption); ok {
			opts.SetOverwrite(*o.overwrite)
		}
	}
	if o.omitAccessTypes != nil {
		if opts, ok := target.(OmitAccessTypesOption); ok {
			opts.SetOmittedAccessTypes(maputils.OrderedKeys(o.omitAccessTypes)...)
		}
	}
	if o.omitArtifactTypes != nil {
		if opts, ok := target.(OmitArtifactTypesOption); ok {
			opts.SetOmittedArtifactTypes(maputils.OrderedKeys(o.omitAccessTypes)...)
		}
	}
	if o.resolver != nil {
		if opts, ok := target.(ResolverOption); ok {
			opts.SetResolver(o.resolver)
		}
	}
	return nil
}

func (o *Options) Apply(opts ...transferhandler.TransferOption) error {
	return transferhandler.ApplyOptions(o, opts...)
}

func (o *Options) SetEnforceTransport(enforce bool) {
	o.enforceTransport = &enforce
}

func (o *Options) IsTransportEnforced() bool {
	return optionutils.AsBool(o.enforceTransport)
}

func (o *Options) SetOverwrite(overwrite bool) {
	o.overwrite = &overwrite
}

func (o *Options) IsOverwrite() bool {
	return optionutils.AsBool(o.overwrite)
}

func (o *Options) SetSkipUpdate(skipupdate bool) {
	o.skipUpdate = &skipupdate
}

func (o *Options) IsSkipUpdate() bool {
	return optionutils.AsBool(o.skipUpdate)
}

func (o *Options) SetRecursive(recursive bool) {
	o.recursive = &recursive
}

func (o *Options) IsRecursive() bool {
	return optionutils.AsBool(o.recursive)
}

func (o *Options) SetResourcesByValue(resourcesByValue bool) {
	o.resourcesByValue = &resourcesByValue
}

func (o *Options) IsResourcesByValue() bool {
	return optionutils.AsBool(o.resourcesByValue)
}

func (o *Options) SetLocalResourcesByValue(resourcesByValue bool) {
	o.localByValue = &resourcesByValue
}

func (o *Options) IsLocalResourcesByValue() bool {
	return optionutils.AsBool(o.localByValue)
}

func (o *Options) SetSourcesByValue(sourcesByValue bool) {
	o.sourcesByValue = &sourcesByValue
}

func (o *Options) IsSourcesByValue() bool {
	return optionutils.AsBool(o.sourcesByValue)
}

func (o *Options) SetKeepGlobalAccess(keepGlobalAccess bool) {
	o.keepGlobalAccess = &keepGlobalAccess
}

func (o *Options) IsKeepGlobalAccess() bool {
	return optionutils.AsBool(o.keepGlobalAccess)
}

func (o *Options) SetRetries(retries int) {
	o.retries = &retries
}

func (o *Options) GetRetries() int {
	if o.retries == nil {
		return 0
	}
	return *o.retries
}

func (o *Options) SetResolver(resolver ocm.ComponentVersionResolver) {
	o.resolver = resolver
}

func (o *Options) GetResolver() ocm.ComponentVersionResolver {
	return o.resolver
}

func (o *Options) SetStopOnExistingVersion(stopOnExistingVersion bool) {
	o.stopOnExisting = &stopOnExistingVersion
}

func (o *Options) IsStopOnExistingVersion() bool {
	return optionutils.AsBool(o.stopOnExisting)
}

func (o *Options) SetOmittedAccessTypes(list ...string) {
	o.omitAccessTypes = set.New[string]()
	for _, t := range list {
		o.omitAccessTypes.Add(t)
	}
}

func (o *Options) AddOmittedAccessTypes(list ...string) {
	if o.omitAccessTypes == nil {
		o.omitAccessTypes = set.New[string]()
	}
	for _, t := range list {
		o.omitAccessTypes.Add(t)
	}
}

func (o *Options) GetOmittedAccessTypes() []string {
	if o.omitAccessTypes == nil {
		return nil
	}
	return maputils.OrderedKeys(o.omitAccessTypes)
}

func (o *Options) IsAccessTypeOmitted(t string) bool {
	if o.omitAccessTypes == nil {
		return false
	}
	if o.omitAccessTypes.Contains(t) {
		return true
	}
	k, _ := runtime.KindVersion(t)
	return o.omitAccessTypes.Contains(k)
}

func (o *Options) SetOmittedArtifactTypes(list ...string) {
	o.omitArtifactTypes = set.New[string]()
	for _, t := range list {
		o.omitArtifactTypes.Add(t)
	}
}

func (o *Options) AddOmittedArtifactTypes(list ...string) {
	if o.omitArtifactTypes == nil {
		o.omitArtifactTypes = set.New[string]()
	}
	for _, t := range list {
		o.omitArtifactTypes.Add(t)
	}
}

func (o *Options) GetOmittedArtifactTypes() []string {
	if o.omitArtifactTypes == nil {
		return nil
	}
	return maputils.OrderedKeys(o.omitArtifactTypes)
}

func (o *Options) IsArtifactTypeOmitted(t string) bool {
	if o.omitArtifactTypes == nil {
		return false
	}
	if o.omitArtifactTypes.Contains(t) {
		return true
	}
	return o.omitArtifactTypes.Contains(t)
}

//////////////////////////////////////////////////////////////////////////////

type EnforceTransportOption interface {
	SetEnforceTransport(bool)
	IsTransportEnforced() bool
}

type enforceTransportOption struct {
	TransferOptionsCreator
	enforce bool
}

func (o *enforceTransportOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	if eff, ok := to.(EnforceTransportOption); ok {
		eff.SetEnforceTransport(o.enforce)
		return nil
	} else {
		return errors.ErrNotSupported(transferhandler.KIND_TRANSFEROPTION, "enforceTransport")
	}
}

// EnforceTransport enforces a transport of a component version as it is.
// This controls whether transport is carried out
// as if the component version were not present at the destination.
func EnforceTransport(args ...bool) transferhandler.TransferOption {
	return &enforceTransportOption{
		enforce: optionutils.GetOptionFlag(args...),
	}
}

///////////////////////////////////////////////////////////////////////////////

type OverwriteOption interface {
	SetOverwrite(bool)
	IsOverwrite() bool
}

type overwriteOption struct {
	TransferOptionsCreator
	overwrite bool
}

func (o *overwriteOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	if eff, ok := to.(OverwriteOption); ok {
		eff.SetOverwrite(o.overwrite)
		return nil
	} else {
		return errors.ErrNotSupported(transferhandler.KIND_TRANSFEROPTION, "overwrite")
	}
}

// Overwrite enables the modification of digest relevant information in a component version.
func Overwrite(args ...bool) transferhandler.TransferOption {
	return &overwriteOption{
		overwrite: optionutils.GetOptionFlag(args...),
	}
}

///////////////////////////////////////////////////////////////////////////////

type SkipUpdateOption interface {
	SetSkipUpdate(bool)
	IsSkipUpdate() bool
}

type skipUpdateOption bool

func (o skipUpdateOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	if eff, ok := to.(SkipUpdateOption); ok {
		eff.SetSkipUpdate(bool(o))
		return nil
	} else {
		return errors.ErrNotSupported("skip-update")
	}
}

// SkipUpdate enables the modification of non-digest (volatile) relevant information in a component version.
func SkipUpdate(args ...bool) transferhandler.TransferOption {
	return skipUpdateOption(optionutils.GetOptionFlag(args...))
}

///////////////////////////////////////////////////////////////////////////////

type RetryOption interface {
	SetRetries(n int)
	GetRetries() int
}

type retryOption struct {
	TransferOptionsCreator
	retries int
}

func (o *retryOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	if eff, ok := to.(RetryOption); ok {
		eff.SetRetries(o.retries)
		return nil
	} else {
		return errors.ErrNotSupported(transferhandler.KIND_TRANSFEROPTION, "retry")
	}
}

// Retries sets the number of retries for failing update operations.
func Retries(retries int) transferhandler.TransferOption {
	return &retryOption{retries: retries}
}

///////////////////////////////////////////////////////////////////////////////

type RecursiveOption interface {
	SetRecursive(bool)
	IsRecursive() bool
}

type recursiveOption struct {
	TransferOptionsCreator
	flag bool
}

func (o *recursiveOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	if eff, ok := to.(RecursiveOption); ok {
		eff.SetRecursive(o.flag)
		return nil
	} else {
		return errors.ErrNotSupported(transferhandler.KIND_TRANSFEROPTION, "recursive")
	}
}

// Recursive enables the transport of the reference closure of a component version.
func Recursive(args ...bool) transferhandler.TransferOption {
	return &recursiveOption{flag: optionutils.GetOptionFlag(args...)}
}

///////////////////////////////////////////////////////////////////////////////

type ResourcesByValueOption interface {
	SetResourcesByValue(bool)
	IsResourcesByValue() bool
}

type resourcesByValueOption struct {
	TransferOptionsCreator
	flag bool
}

func (o *resourcesByValueOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	if eff, ok := to.(ResourcesByValueOption); ok {
		eff.SetResourcesByValue(o.flag)
		return nil
	} else {
		return errors.ErrNotSupported(transferhandler.KIND_TRANSFEROPTION, "resources-by-value")
	}
}

// ResourcesByValue enables the transport a resources by values instead of by-reference.
func ResourcesByValue(args ...bool) transferhandler.TransferOption {
	return &resourcesByValueOption{flag: optionutils.GetOptionFlag(args...)}
}

///////////////////////////////////////////////////////////////////////////////

type LocalResourcesByValueOption interface {
	SetLocalResourcesByValue(bool)
	IsLocalResourcesByValue() bool
}

type intrscsByValueOption struct {
	TransferOptionsCreator
	flag bool
}

func (o *intrscsByValueOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	if eff, ok := to.(LocalResourcesByValueOption); ok {
		eff.SetLocalResourcesByValue(o.flag)
		return nil
	} else {
		return errors.ErrNotSupported(transferhandler.KIND_TRANSFEROPTION, "local-resources by-value")
	}
}

// LocalResourcesByValue enables the transport a local (relation) resources by values instead of by-reference.
func LocalResourcesByValue(args ...bool) transferhandler.TransferOption {
	return &intrscsByValueOption{flag: optionutils.GetOptionFlag(args...)}
}

///////////////////////////////////////////////////////////////////////////////

type SourcesByValueOption interface {
	SetSourcesByValue(bool)
	IsSourcesByValue() bool
}

type sourcesByValueOption struct {
	TransferOptionsCreator
	flag bool
}

func (o *sourcesByValueOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	if eff, ok := to.(SourcesByValueOption); ok {
		eff.SetSourcesByValue(o.flag)
		return nil
	} else {
		return errors.ErrNotSupported(transferhandler.KIND_TRANSFEROPTION, "sources by-value")
	}
}

// SourcesByValue enables the transport a sources by values instead of by-reference.
func SourcesByValue(args ...bool) transferhandler.TransferOption {
	return &sourcesByValueOption{flag: optionutils.GetOptionFlag(args...)}
}

///////////////////////////////////////////////////////////////////////////////

type ResolverOption interface {
	GetResolver() ocm.ComponentVersionResolver
	SetResolver(ocm.ComponentVersionResolver)
}

type resolverOption struct {
	TransferOptionsCreator
	resolver ocm.ComponentVersionResolver
}

func (o *resolverOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	if eff, ok := to.(ResolverOption); ok {
		eff.SetResolver(o.resolver)
		return nil
	} else {
		return errors.ErrNotSupported(transferhandler.KIND_TRANSFEROPTION, "resolver")
	}
}

// Resolver specifies a resolver used to resolve nested component versions.
func Resolver(resolver ocm.ComponentVersionResolver) transferhandler.TransferOption {
	return &resolverOption{
		resolver: resolver,
	}
}

///////////////////////////////////////////////////////////////////////////////

type KeepGlobalAccessOption interface {
	SetKeepGlobalAccess(bool)
	IsKeepGlobalAccess() bool
}

type keepGlobalOption struct {
	TransferOptionsCreator
	flag bool
}

func (o *keepGlobalOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	if eff, ok := to.(KeepGlobalAccessOption); ok {
		eff.SetKeepGlobalAccess(o.flag)
		return nil
	} else {
		return errors.ErrNotSupported(transferhandler.KIND_TRANSFEROPTION, "keep-global-access")
	}
}

// KeepGlobalAccess enables to keep local blobs if uploaders are used to upload imported blobs.
func KeepGlobalAccess(args ...bool) transferhandler.TransferOption {
	return &keepGlobalOption{flag: optionutils.GetOptionFlag(args...)}
}

///////////////////////////////////////////////////////////////////////////////

type StopOnExistingVersionOption interface {
	SetStopOnExistingVersion(bool)
	IsStopOnExistingVersion() bool
}

type stopOnExistingVersionOption struct {
	TransferOptionsCreator
	flag bool
}

func (o *stopOnExistingVersionOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	if eff, ok := to.(StopOnExistingVersionOption); ok {
		eff.SetStopOnExistingVersion(o.flag)
		return nil
	} else {
		return errors.ErrNotSupported(transferhandler.KIND_TRANSFEROPTION, "stop-on-existing")
	}
}

// StopOnExistingVersion stops the recursion on component versions already present in target.
func StopOnExistingVersion(args ...bool) transferhandler.TransferOption {
	return &stopOnExistingVersionOption{flag: optionutils.GetOptionFlag(args...)}
}

///////////////////////////////////////////////////////////////////////////////

type OmitAccessTypesOption interface {
	SetOmittedAccessTypes(...string)
	GetOmittedAccessTypes() []string
}

type omitAccessTypesOption struct {
	TransferOptionsCreator
	add  bool
	list []string
}

func (o *omitAccessTypesOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	if eff, ok := to.(OmitAccessTypesOption); ok {
		if o.add {
			eff.SetOmittedAccessTypes(sliceutils.CopyAppend(eff.GetOmittedAccessTypes(), o.list...)...)
		} else {
			eff.SetOmittedAccessTypes(o.list...)
		}
		return nil
	} else {
		return errors.ErrNotSupported(transferhandler.KIND_TRANSFEROPTION, "omit-access-types")
	}
}

// OmitAccessTypes somits the specified access types from value transport.
func OmitAccessTypes(list ...string) transferhandler.TransferOption {
	return &omitAccessTypesOption{
		list: slices.Clone(list),
	}
}

func AddOmittedAccessTypes(list ...string) transferhandler.TransferOption {
	return &omitAccessTypesOption{
		add:  true,
		list: slices.Clone(list),
	}
}

///////////////////////////////////////////////////////////////////////////////

type OmitArtifactTypesOption interface {
	SetOmittedArtifactTypes(...string)
	GetOmittedArtifactTypes() []string
}

type omitArtifactTypesOption struct {
	add  bool
	list []string
}

func (o *omitArtifactTypesOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	if eff, ok := to.(OmitAccessTypesOption); ok {
		if o.add {
			eff.SetOmittedAccessTypes(append(eff.GetOmittedAccessTypes(), o.list...)...)
		} else {
			eff.SetOmittedAccessTypes(o.list...)
		}
		return nil
	} else {
		return errors.ErrNotSupported("omit-artifact-types")
	}
}

// OmitArtifactTypes somits the specified artifact types from value transport.
func OmitArtifactTypes(list ...string) transferhandler.TransferOption {
	return &omitArtifactTypesOption{
		list: slices.Clone(list),
	}
}

func AddOmittedArtifactTypes(list ...string) transferhandler.TransferOption {
	return &omitArtifactTypesOption{
		add:  true,
		list: slices.Clone(list),
	}
}
